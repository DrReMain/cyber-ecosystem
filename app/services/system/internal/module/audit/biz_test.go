package audit

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"
	"testing"

	kaudit "cyber-ecosystem/shared-go/kratos/audit"
	"cyber-ecosystem/shared-go/kratos/security"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
)

// The CSV export's formula-injection guard: attacker-controlled cells
// (user agent, operation, path, deny reason, actor, ip) start life as
// request data, and a leading =, +, -, @, or tab makes Excel/Sheets execute
// the cell as a formula when an admin opens the export. The sanitizer
// prefixes such cells with an apostrophe; every other value passes through
// byte-identical, and numeric columns never route through it (a negative
// number is data, not a formula).
func TestSanitizeAuditCell(t *testing.T) {
	cases := []struct{ in, want string }{
		{"=HYPERLINK(\"http://evil\",\"click\")", "'=HYPERLINK(\"http://evil\",\"click\")"},
		{"+SUM(A1:A9)", "'+SUM(A1:A9)"},
		{"-2+3|cmd", "'-2+3|cmd"},
		{"@SUM(A1)", "'@SUM(A1)"},
		{"\tTAB(1)", "'\tTAB(1)"},
		{"Mozilla/5.0 (Macintosh)", "Mozilla/5.0 (Macintosh)"},
		{"/cyber.system.v1.UserService/CreateUser", "/cyber.system.v1.UserService/CreateUser"},
		{"a=b stays literal", "a=b stays literal"}, // only the first byte can arm a formula
		{"", ""},
	}
	for _, c := range cases {
		if got := sanitizeAuditCell(c.in); got != c.want {
			t.Fatalf("sanitizeAuditCell(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// fakeAuditRP answers the read faces and records everything the UC owes:
// the tenant stamp on the list input, the export ledger write inside the
// generate transaction, and the event handoff on the sink face. Faces a
// test never reaches panic — they are not under test here.
type fakeAuditRP struct {
	AuditRP

	listIn  *ListIn
	listOut *ListOut

	exportListIn  *ExportListIn
	exportListOut *ExportListOut

	afterBatches [][]*AuditLog
	afterCalls   []string
	afterErr     error

	idsChunks [][]string
	idsRows   []*AuditLog
	idsErr    error

	createdExports []struct{ fileID, ownerID string }
	createErr      error

	published []*kaudit.Event
}

func (f *fakeAuditRP) ListAuditLogs(_ context.Context, in *ListIn) (*ListOut, error) {
	f.listIn = in
	return f.listOut, nil
}

func (f *fakeAuditRP) ListExports(_ context.Context, in *ExportListIn) (*ExportListOut, error) {
	f.exportListIn = in
	return f.exportListOut, nil
}

func (f *fakeAuditRP) ListLogsAfter(_ context.Context, _, after string, _ int) ([]*AuditLog, error) {
	if f.afterErr != nil {
		return nil, f.afterErr
	}
	f.afterCalls = append(f.afterCalls, after)
	batch := f.afterBatches[0]
	f.afterBatches = f.afterBatches[1:]
	return batch, nil
}

func (f *fakeAuditRP) FindLogsByIDs(_ context.Context, _ string, ids []string) ([]*AuditLog, error) {
	if f.idsErr != nil {
		return nil, f.idsErr
	}
	f.idsChunks = append(f.idsChunks, ids)
	return f.idsRows, nil
}

func (f *fakeAuditRP) CreateExport(_ context.Context, fileID, ownerID string) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.createdExports = append(f.createdExports, struct{ fileID, ownerID string }{fileID, ownerID})
	return nil
}

func (f *fakeAuditRP) Publish(_ context.Context, ev *kaudit.Event) {
	f.published = append(f.published, ev)
}

func newAuditUC(rp *fakeAuditRP) *AuditUC {
	return NewAuditUC(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		nil, rp, nil,
	)
}

// The read rule lives on the UC, not the query: audit_log carries no
// tenant mixin (denied and subject-less events must audit), so List
// stamps the caller's tenant onto the input right before the repo read —
// and refuses to run at all without a subject.
func TestListStampsCallerTenant(t *testing.T) {
	rp := &fakeAuditRP{listOut: &ListOut{}}
	uc := newAuditUC(rp)
	ctx := security.WithSubject(context.Background(), &security.Subject{UserID: "u1", TenantID: "t1"})

	if _, err := uc.List(ctx, &ListIn{}); err != nil {
		t.Fatalf("list: %v", err)
	}
	if rp.listIn == nil || rp.listIn.Tenant != "t1" {
		t.Fatalf("tenant stamp: %+v", rp.listIn)
	}

	if _, err := uc.List(context.Background(), &ListIn{}); !errorspb.IsGeneralErrorUnauthenticated(err) {
		t.Fatalf("subjectless list: want unauthenticated, got %v", err)
	}
}

// ListExports is a pure delegation — paging and ordering live in the repo.
func TestListExportsDelegates(t *testing.T) {
	want := &ExportListOut{}
	rp := &fakeAuditRP{exportListOut: want}
	uc := newAuditUC(rp)

	in := &ExportListIn{}
	out, err := uc.ListExports(context.Background(), in)
	if err != nil || out != want {
		t.Fatalf("list exports: out=%v err=%v", out, err)
	}
	if rp.exportListIn != in {
		t.Fatal("list exports: input must pass through untouched")
	}
}

// The sink face: whatever the ladder hands the UC reaches the emitter
// verbatim — the UC adds nothing, drops nothing.
func TestEmitPublishesVerbatim(t *testing.T) {
	rp := &fakeAuditRP{}
	uc := newAuditUC(rp)
	ev := &kaudit.Event{Operation: "e2e op"}

	uc.Emit(context.Background(), ev)
	if len(rp.published) != 1 || rp.published[0] != ev {
		t.Fatalf("publish: %+v", rp.published)
	}
}

// The export refuses without a subject — the tenant rule and the owner
// ledger both hang off it. The full generate chain (spec → file UC →
// export row inside the transaction) is the file module's own heavily
// tested territory plus a live walkthrough; this arm pins the gate.
func TestExportRequiresSubject(t *testing.T) {
	uc := newAuditUC(&fakeAuditRP{})
	if _, err := uc.Export(context.Background(), &ExportIn{}); !errorspb.IsGeneralErrorUnauthenticated(err) {
		t.Fatalf("subjectless export: want unauthenticated, got %v", err)
	}
}

func csvRows(t *testing.T, w *bytes.Buffer) []string {
	t.Helper()
	rows := strings.Split(strings.TrimSuffix(w.String(), "\n"), "\n")
	if len(rows) == 0 || rows[0] != strings.Join(auditCSVHeader, ",") {
		t.Fatalf("csv header: %q", rows[0])
	}
	return rows
}

func batchOf(n int, prefix string) []*AuditLog {
	out := make([]*AuditLog, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, &AuditLog{ID: fmt.Sprintf("%s-%04d", prefix, i), Tenant: "t1", Status: 200, LatencyMs: 5})
	}
	return out
}

// The full-table export walks the keyset path: batches of the export
// size, resuming from the last id each hop, stopping on the short batch —
// memory stays O(batch) whatever the row count.
func TestWriteAuditCSVKeysetPagination(t *testing.T) {
	full, tail := batchOf(auditExportBatch, "a"), batchOf(3, "b")
	rp := &fakeAuditRP{afterBatches: [][]*AuditLog{full, tail}}
	uc := newAuditUC(rp)

	var w bytes.Buffer
	if err := uc.writeAuditCSV(context.Background(), &w, "t1", nil); err != nil {
		t.Fatalf("csv: %v", err)
	}
	rows := csvRows(t, &w)
	if len(rows) != 1+len(full)+len(tail) {
		t.Fatalf("csv rows: %d, want %d", len(rows), 1+len(full)+len(tail))
	}
	// The cursor resumes from the first batch's last id — the second hop
	// names it, the first hop starts from the beginning.
	if len(rp.afterCalls) != 2 || rp.afterCalls[0] != "" || rp.afterCalls[1] != full[len(full)-1].ID {
		t.Fatalf("keyset cursors: %v", rp.afterCalls)
	}
}

// The selected export walks the id-chunk path: ids travel in
// export-size chunks, and unknown ids simply export fewer rows.
func TestWriteAuditCSVByIDChunks(t *testing.T) {
	ids := make([]string, auditExportBatch+100)
	for i := range ids {
		ids[i] = strconv.Itoa(i)
	}
	rp := &fakeAuditRP{idsRows: batchOf(2, "s")}
	uc := newAuditUC(rp)

	var w bytes.Buffer
	if err := uc.writeAuditCSV(context.Background(), &w, "t1", ids); err != nil {
		t.Fatalf("csv: %v", err)
	}
	rows := csvRows(t, &w)
	// Each chunk answers the same fixed page, so rows = header + pages × rows.
	if len(rows) != 1+len(rp.idsChunks)*len(rp.idsRows) {
		t.Fatalf("csv rows: %d", len(rows))
	}
	if len(rp.idsChunks) != 2 || len(rp.idsChunks[0]) != auditExportBatch || len(rp.idsChunks[1]) != 100 {
		t.Fatalf("chunks: %d %v", len(rp.idsChunks), rp.idsChunks)
	}
}

// The free-text columns route through the formula guard on the way out,
// and a repo failure mid-walk surfaces instead of a half-silent file.
func TestWriteAuditCSVSanitizesAndSurfaces(t *testing.T) {
	t.Run("sanitizes free text", func(t *testing.T) {
		rp := &fakeAuditRP{afterBatches: [][]*AuditLog{{{
			ID: "x1", Tenant: "t1", Actor: "=cmd", Operation: "op", Status: 1, LatencyMs: 2,
		}}}}
		uc := newAuditUC(rp)
		var w bytes.Buffer
		if err := uc.writeAuditCSV(context.Background(), &w, "t1", nil); err != nil {
			t.Fatalf("csv: %v", err)
		}
		rows := csvRows(t, &w)
		if !strings.Contains(rows[1], "'=cmd") {
			t.Fatalf("sanitized row: %q", rows[1])
		}
	})
	t.Run("repo failure surfaces", func(t *testing.T) {
		uc := newAuditUC(&fakeAuditRP{afterErr: fmt.Errorf("boom")})
		var w bytes.Buffer
		if err := uc.writeAuditCSV(context.Background(), &w, "t1", nil); err == nil || err.Error() != "boom" {
			t.Fatalf("walk failure: got %v", err)
		}
	})
}
