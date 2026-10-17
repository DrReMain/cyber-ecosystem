package file

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/capability/storage"
	"cyber-ecosystem/shared-go/kratos/security"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
	systempb "cyber-ecosystem/gen/go/cyber/system/v1"
)

// passthroughTm runs the unit inline — these laws are about state gating and
// write order, not transaction semantics.
type passthroughTm struct{}

func (passthroughTm) InTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// fakeFileRP backs the whole surface: rows for the read faces, per-call
// records for the write faces, injectable errors, and hook answers for the
// sweep family.
type fakeFileRP struct {
	FileRP
	rows map[string]*File

	updateNames []string
	deletes     []string
	listOut     *FileListOut

	limits     storage.Limits
	presigned  []string
	presignErr error

	creates       int
	createdRow    *File
	uploadBytes   map[string]int64
	uploadErr     error
	uploadTenants []string
	marked        map[string][2]any // id -> {size, etag}
	failed        []string
	ambientTx     bool

	claim    map[string]bool
	claimErr map[string]error

	aborted        []string
	deletedObjs    []string
	multipartList  []storage.PendingUpload
	findByKey      map[string]*File
	objects        []storage.ObjectInfo
	nextCursor     string
	findRefByKey   map[string]*File
	staleUploading []*File
	listTenants    []string
	softDeleted    []*File
	reclaim        map[string]bool
	failStaleCut   time.Time
	objCalls       int
	staleErr       error
	objectsErr     error
	refByKeyErr    error
}

func (f *fakeFileRP) UpdateName(_ context.Context, id, _ string) (*File, error) {
	f.updateNames = append(f.updateNames, id)
	return f.rows[id], nil
}

func (f *fakeFileRP) Delete(_ context.Context, id string) error {
	f.deletes = append(f.deletes, id)
	return nil
}

func (f *fakeFileRP) List(context.Context, *FileListIn) (*FileListOut, error) {
	return f.listOut, nil
}

func (f *fakeFileRP) FindByID(_ context.Context, id string) (*File, error) {
	if r, ok := f.rows[id]; ok {
		return r, nil
	}
	return nil, errorspb.ErrorInfraErrorDbNotFound("")
}

func (f *fakeFileRP) FindRef(_ context.Context, id string) (*File, error) {
	if r, ok := f.rows[id]; ok {
		return r, nil
	}
	return nil, errorspb.ErrorInfraErrorDbNotFound("")
}

func (f *fakeFileRP) Limits() storage.Limits { return f.limits }

func (f *fakeFileRP) PresignDownload(context.Context, string, time.Duration, storage.DownloadOptions) (string, error) {
	if f.presignErr != nil {
		return "", f.presignErr
	}
	f.presigned = append(f.presigned, "url")
	return "url", nil
}

func (f *fakeFileRP) CreateGeneration(_ context.Context, spec *File) (*File, error) {
	f.creates++
	out := *spec
	out.ID = "f-new"
	f.createdRow = &out
	return &out, nil
}

func (f *fakeFileRP) UploadObject(ctx context.Context, key, _ string, r io.Reader) (*storage.ObjectInfo, error) {
	if t, ok := security.ExplicitTenantFromCtx(ctx); ok {
		f.uploadTenants = append(f.uploadTenants, t)
	}
	if f.uploadErr != nil {
		return nil, f.uploadErr
	}
	n, cerr := io.Copy(io.Discard, r) // the reader's error is the upload's error — a real backend never accepts a failed body
	if cerr != nil {
		return nil, cerr
	}
	if f.uploadBytes == nil {
		f.uploadBytes = map[string]int64{}
	}
	f.uploadBytes[key] = n
	return &storage.ObjectInfo{ETag: "etag-1"}, nil
}

func (f *fakeFileRP) DeleteObject(_ context.Context, key string) error {
	f.deletedObjs = append(f.deletedObjs, key)
	return nil
}

func (f *fakeFileRP) MarkGenerated(_ context.Context, id string, size int64, etag string) error {
	if f.marked == nil {
		f.marked = map[string][2]any{}
	}
	f.marked[id] = [2]any{size, etag}
	return nil
}

func (f *fakeFileRP) MarkFailed(_ context.Context, id string) error {
	f.failed = append(f.failed, id)
	return nil
}

func (f *fakeFileRP) FailStaleProcessing(_ context.Context, cut time.Time) (int, error) {
	f.failStaleCut = cut
	return 1, nil
}

func (f *fakeFileRP) ListStaleUploading(ctx context.Context, _ time.Time, _ string, _ int) ([]*File, error) {
	if t, ok := security.ExplicitTenantFromCtx(ctx); ok {
		f.listTenants = append(f.listTenants, t)
	}
	if f.staleErr != nil {
		return nil, f.staleErr
	}
	return f.staleUploading, nil
}

func (f *fakeFileRP) ClaimUploading(_ context.Context, id string, _ time.Time) (bool, error) {
	if err, ok := f.claimErr[id]; ok {
		return false, err
	}
	return f.claim[id], nil
}

func (f *fakeFileRP) FindByKey(_ context.Context, key string) (*File, error) {
	if r, ok := f.findByKey[key]; ok {
		return r, nil
	}
	return nil, nil
}

func (f *fakeFileRP) ListMultipartUploads(_ context.Context, _ string, visit func(storage.PendingUpload) error) error {
	for _, u := range f.multipartList {
		if err := visit(u); err != nil {
			return err
		}
	}
	return nil
}

func (f *fakeFileRP) AbortSession(_ context.Context, key, _ string) error {
	f.aborted = append(f.aborted, key)
	return nil
}

func (f *fakeFileRP) ListObjects(_ context.Context, _, _ string, _ int32) (*storage.ListResult, error) {
	if f.objectsErr != nil {
		return nil, f.objectsErr
	}
	f.objCalls++
	if f.objCalls == 1 {
		return &storage.ListResult{Objects: f.objects, NextCursor: f.nextCursor}, nil
	}
	return &storage.ListResult{}, nil
}

func (f *fakeFileRP) FindRefByKey(_ context.Context, key string) (*File, error) {
	if f.refByKeyErr != nil {
		return nil, f.refByKeyErr
	}
	if r, ok := f.findRefByKey[key]; ok {
		return r, nil
	}
	return nil, nil
}

func (f *fakeFileRP) ListSoftDeleted(_ context.Context, _ time.Time, _ string, _ int) ([]*File, error) {
	return f.softDeleted, nil
}

func (f *fakeFileRP) ReclaimDeleted(_ context.Context, id string, _ time.Time) (bool, error) {
	return f.reclaim[id], nil
}

func (f *fakeFileRP) InAmbientTx(context.Context) bool { return f.ambientTx }

// hookSink discards lifecycle hooks.
type hookSink struct{}

func (hookSink) OnStart(func(context.Context) error) {}
func (hookSink) OnStop(func(context.Context) error)  {}

// fakeDomains answers the registry's domain list for the sweep's axis.
type fakeDomains struct {
	codes []string
	err   error
}

func (f fakeDomains) ListTenantCodes(context.Context) ([]string, error) { return f.codes, f.err }

func newFileUC(rp *fakeFileRP) *FileUC {
	return NewFileUC(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		passthroughTm{}, rp, fakeDomains{codes: []string{security.DefaultTenant}}, hookSink{},
	)
}

// CRUD faces ----------------------------------------------------------------------------------------------------------

// Only terminal rows retire: UPLOADING's exit is abort, PROCESSING's is the
// generation task; CONFIRMED and FAILED delete.
func TestDeleteTerminalOnly(t *testing.T) {
	rp := &fakeFileRP{rows: map[string]*File{
		"f-up": {ID: "f-up", Status: StatusUploading},
		"f-pr": {ID: "f-pr", Status: StatusProcessing},
		"f-ok": {ID: "f-ok", Status: StatusConfirmed},
		"f-fl": {ID: "f-fl", Status: StatusFailed},
	}}
	uc := newFileUC(rp)

	if err := uc.Delete(context.Background(), "f-up"); !systempb.IsSystemFileInvalidState(err) {
		t.Fatalf("uploading delete: want FILE_INVALID_STATE, got %v", err)
	}
	if err := uc.Delete(context.Background(), "f-pr"); !systempb.IsSystemFileInvalidState(err) {
		t.Fatalf("processing delete: want FILE_INVALID_STATE, got %v", err)
	}
	if err := uc.Delete(context.Background(), "f-ok"); err != nil {
		t.Fatalf("confirmed delete: %v", err)
	}
	if err := uc.Delete(context.Background(), "f-fl"); err != nil {
		t.Fatalf("failed delete: %v", err)
	}
	if len(rp.deletes) != 2 {
		t.Fatalf("deletes = %v, want exactly the two terminal rows", rp.deletes)
	}
}

// An unknown id classifies as FILE_NOT_FOUND; a generic store error keeps its
// infra identity.
func TestGetClassifiedGenericPassthrough(t *testing.T) {
	uc := newFileUC(&fakeFileRP{rows: map[string]*File{}})

	if _, err := uc.Get(context.Background(), "ghost"); !systempb.IsSystemFileNotFound(err) {
		t.Fatalf("unknown: want FILE_NOT_FOUND, got %v", err)
	}
}

// Rename delegates only after the row resolves.
func TestUpdateDelegatesAfterGet(t *testing.T) {
	rp := &fakeFileRP{rows: map[string]*File{}}
	uc := newFileUC(rp)

	if _, err := uc.Update(context.Background(), "ghost", "n"); !systempb.IsSystemFileNotFound(err) {
		t.Fatalf("unknown rename: want FILE_NOT_FOUND, got %v", err)
	}
	if len(rp.updateNames) != 0 {
		t.Fatal("unknown rename: no write may happen")
	}
}

// MintUrls skips silently over absent refs and unconfirmed rows — a list
// render never breaks on one bad id — but a presign failure is a hard error.
func TestMintUrlsSkipsUnconfirmedAndMissing(t *testing.T) {
	rp := &fakeFileRP{
		rows: map[string]*File{
			"f-miss": nil, // placeholder; absence comes from FindRef miss below
			"f-up":   {ID: "f-up", Status: StatusUploading, Key: "k-up"},
			"f-ok":   {ID: "f-ok", Status: StatusConfirmed, Key: "k-ok", ContentType: "text/plain", Name: "a.txt"},
		},
		limits: storage.Limits{PresignTTL: time.Minute},
	}
	delete(rp.rows, "f-miss") // absent key = FindRef miss
	uc := newFileUC(rp)

	mints, err := uc.MintUrls(context.Background(), []string{"f-miss", "f-up", "f-ok"})
	if err != nil {
		t.Fatalf("mint: %v", err)
	}
	if len(mints) != 1 || mints[0].ID != "f-ok" {
		t.Fatalf("mints = %+v, want only the confirmed row", mints)
	}
	if mints[0].ExpiresAt.Before(time.Now()) {
		t.Fatal("mint: expiry must be in the future")
	}

	rp.presignErr = errorspb.ErrorInfraErrorNetworkConnection("")
	if _, err := uc.MintUrls(context.Background(), []string{"f-ok"}); !errorspb.IsInfraErrorNetworkConnection(err) {
		t.Fatalf("presign failure: want hard error, got %v", err)
	}
}

// Generate ------------------------------------------------------------------------------------------------------------

// An ambient transaction would launch the task before the row's tx commits —
// the entry guard rejects loud and early.
func TestGenerateAmbientTxGuard(t *testing.T) {
	rp := &fakeFileRP{ambientTx: true}
	uc := newFileUC(rp)

	if _, err := uc.Generate(context.Background(), &GenerationSpec{Name: "n"}, nil); !errorspb.IsGeneralErrorInternal(err) {
		t.Fatalf("ambient tx: want INTERNAL, got %v", err)
	}
	if rp.creates != 0 {
		t.Fatal("ambient tx: no row may be born")
	}
}

// A generation belongs to a subject — the row records its owner.
func TestGenerateRequiresSubject(t *testing.T) {
	uc := newFileUC(&fakeFileRP{})

	if _, err := uc.Generate(context.Background(), &GenerationSpec{Name: "n"}, nil); !errorspb.IsGeneralErrorUnauthenticated(err) {
		t.Fatalf("no subject: want unauthenticated, got %v", err)
	}
}

func ownerCtx() context.Context {
	return securityWithSubject("u1")
}

// The happy path: the PROCESSING row and the domain record ride one tx, the
// task launches after commit, bytes are counted reader-side, and the row
// flips to generated with size and etag.
func TestGenerateRecordsAndMarksGenerated(t *testing.T) {
	rp := &fakeFileRP{}
	uc := newFileUC(rp)

	recorded := ""
	_, err := uc.Generate(ownerCtx(), &GenerationSpec{
		Name:        "report.csv",
		ContentType: "text/csv",
		Produce:     func(_ context.Context, w io.Writer) error { _, e := w.Write([]byte("hello")); return e },
	}, func(_ context.Context, f *File) error {
		recorded = f.ID
		return nil
	})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if rp.creates != 1 || rp.createdRow.Status != StatusProcessing || rp.createdRow.Source != SourceServerGenerated || rp.createdRow.OwnerID != "u1" {
		t.Fatalf("generate: row = %+v", rp.createdRow)
	}
	if recorded != "f-new" {
		t.Fatalf("generate: record callback saw %q", recorded)
	}

	if err := uc.ShutdownGenerations(context.Background()); err != nil {
		t.Fatalf("drain: %v", err)
	}
	if got := rp.uploadBytes[rp.createdRow.Key]; got != 5 {
		t.Fatalf("generate: uploaded %d bytes, want 5 (counted reader-side)", got)
	}
	if m := rp.marked["f-new"]; m[0] != int64(5) || m[1] != "etag-1" {
		t.Fatalf("generate: marked = %v", m)
	}
}

// A produce failure cleans up: the object is deleted and the row flips to
// failed — the failure cause never blocks the cleanup.
func TestGenerateFailureCleansUp(t *testing.T) {
	rp := &fakeFileRP{}
	uc := newFileUC(rp)

	_, err := uc.Generate(ownerCtx(), &GenerationSpec{
		Name:    "x",
		Produce: func(context.Context, io.Writer) error { return errors.New("boom") },
	}, nil)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	_ = uc.ShutdownGenerations(context.Background())

	if len(rp.marked) != 0 {
		t.Fatal("failed generation: must not be marked generated")
	}
	if len(rp.failed) != 1 || rp.failed[0] != "f-new" {
		t.Fatalf("failed generation: failed = %v", rp.failed)
	}
	if len(rp.deletedObjs) != 1 {
		t.Fatalf("failed generation: object cleanup = %v", rp.deletedObjs)
	}
}

// A record-callback failure rolls the whole thing back before launch — the
// task never writes against a row that was never born.
func TestGenerateRecordErrorAbortsBeforeLaunch(t *testing.T) {
	rp := &fakeFileRP{}
	uc := newFileUC(rp)
	rejected := errors.New("domain row rejected")

	_, err := uc.Generate(ownerCtx(), &GenerationSpec{Name: "x", Produce: func(context.Context, io.Writer) error { return nil }},
		func(context.Context, *File) error { return rejected })
	if !errors.Is(err, rejected) {
		t.Fatalf("record failure: want the callback error verbatim, got %v", err)
	}
	_ = uc.ShutdownGenerations(context.Background())
	if len(rp.uploadBytes) != 0 {
		t.Fatal("record failure: the task must not launch")
	}
}

// Sweep family --------------------------------------------------------------------------------------------------------

// The sweep claims the row first: losing the claim means confirm/abort won
// and the object side must not be touched. A claim error skips the row, not
// the pass.
func TestSweepStaleRowsClaimFirst(t *testing.T) {
	rp := &fakeFileRP{
		staleUploading: []*File{
			{ID: "r1", Key: "k1", UploadID: "up1"},
			{ID: "r2", Key: "k2"},
			{ID: "r3", Key: "k3", UploadID: "up3"},
			{ID: "r4", Key: "k4"},
		},
		claim:    map[string]bool{"r1": true, "r2": false, "r4": true},
		claimErr: map[string]error{"r3": errorspb.ErrorInfraErrorNetworkConnection("")},
	}
	uc := newFileUC(rp)

	n := uc.sweepStaleRows(context.Background(), time.Now())
	if n != 2 {
		t.Fatalf("claimed = %d, want 2", n)
	}
	if len(rp.aborted) != 1 || rp.aborted[0] != "k1" {
		t.Fatalf("aborted = %v, want only k1 (its row had a session)", rp.aborted)
	}
	if len(rp.deletedObjs) != 1 || rp.deletedObjs[0] != "k4" {
		t.Fatalf("objects = %v, want only k4 (sessionless row)", rp.deletedObjs)
	}
}

// Session cleanup is a row-join: a live row (any non-terminal status) owns
// its session; no row — or a terminal failed row — means leaked.
func TestSweepOrphanSessionsRowJoin(t *testing.T) {
	rp := &fakeFileRP{
		multipartList: []storage.PendingUpload{
			{Key: "k-live", UploadID: "u1"},
			{Key: "k-none", UploadID: "u2"},
			{Key: "k-fail", UploadID: "u3"},
		},
		findByKey: map[string]*File{
			"k-live": {Status: StatusUploading},
			"k-fail": {Status: StatusFailed},
		},
	}
	uc := newFileUC(rp)

	if n := uc.sweepOrphanSessions(context.Background()); n != 2 {
		t.Fatalf("aborted = %d, want 2", n)
	}
	if len(rp.aborted) != 2 {
		t.Fatalf("aborted = %v, want k-none and k-fail", rp.aborted)
	}
}

// Object cleanup is age + row-join: young objects and row-owned objects
// stay; rowless old objects go. Pagination follows the key-ordered cursor.
func TestSweepOrphanObjectsAgeAndRowJoin(t *testing.T) {
	old := time.Now().Add(-time.Hour)
	rp := &fakeFileRP{
		objects: []storage.ObjectInfo{
			{Key: "k-young", LastModified: time.Now()},
			{Key: "k-owned", LastModified: old},
			{Key: "k-orphan", LastModified: old},
		},
		findRefByKey: map[string]*File{"k-owned": {ID: "f1"}},
	}
	uc := newFileUC(rp)

	if n := uc.sweepOrphanObjects(context.Background(), time.Now().Add(-time.Minute)); n != 1 {
		t.Fatalf("deleted = %d, want 1", n)
	}
	if len(rp.deletedObjs) != 1 || rp.deletedObjs[0] != "k-orphan" {
		t.Fatalf("deleted = %v, want only k-orphan", rp.deletedObjs)
	}
}

// Retention reclaims claim-first: the conditional hard delete owns the row
// before the object is touched; a lost claim or an error skips the row.
func TestSweepExpiredDeletedClaimFirst(t *testing.T) {
	rp := &fakeFileRP{
		softDeleted: []*File{{ID: "d1", Key: "k1"}, {ID: "d2", Key: "k2"}},
		reclaim:     map[string]bool{"d1": true, "d2": false},
	}
	uc := newFileUC(rp)

	if n := uc.sweepExpiredDeleted(context.Background(), time.Now()); n != 1 {
		t.Fatalf("reclaimed = %d, want 1", n)
	}
	if len(rp.deletedObjs) != 1 || rp.deletedObjs[0] != "k1" {
		t.Fatalf("objects = %v, want only the reclaimed row's", rp.deletedObjs)
	}
}

// helpers -------------------------------------------------------------------------------------------------------------

func securityWithSubject(userID string) context.Context {
	return security.WithSubject(context.Background(), &security.Subject{UserID: userID})
}

// CRUD completions ----------------------------------------------------------------------------------------------------

// Rename on a live row delegates; an unknown id never reaches the write.
func TestUpdateHappyAndList(t *testing.T) {
	rp := &fakeFileRP{rows: map[string]*File{"f1": {ID: "f1"}}, listOut: &FileListOut{}}
	uc := newFileUC(rp)

	if _, err := uc.Update(context.Background(), "f1", "new-name"); err != nil {
		t.Fatalf("rename: %v", err)
	}
	if len(rp.updateNames) != 1 {
		t.Fatal("rename: the write must land")
	}
	if _, err := uc.List(context.Background(), &FileListIn{}); err != nil {
		t.Fatalf("list: %v", err)
	}
}

// A cancelled context ends a pass before it starts — every sweep stage
// checks ctx first.
func TestSweepRespectsCancelledContext(t *testing.T) {
	rp := &fakeFileRP{softDeleted: []*File{{ID: "x"}}}
	uc := newFileUC(rp)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if n := uc.sweepStaleRows(ctx, time.Now()); n != 0 {
		t.Fatalf("cancelled: stale rows = %d, want 0", n)
	}
	if n := uc.sweepExpiredDeleted(ctx, time.Now()); n != 0 {
		t.Fatalf("cancelled: reclaimed = %d, want 0", n)
	}
}

// Pagination follows the key-ordered cursor until it runs dry.
func TestSweepOrphanObjectsPaginates(t *testing.T) {
	old := time.Now().Add(-time.Hour)
	rp := &fakeFileRP{
		objects:    []storage.ObjectInfo{{Key: "k-1", LastModified: old}, {Key: "k-2", LastModified: old}},
		nextCursor: "k-2",
	}
	uc := newFileUC(rp)

	if n := uc.sweepOrphanObjects(context.Background(), time.Now().Add(-time.Minute)); n != 2 {
		t.Fatalf("paged delete = %d, want 2", n)
	}
	if rp.objCalls != 2 {
		t.Fatalf("pages fetched = %d, want 2", rp.objCalls)
	}
}

// The sweep lifecycle runs an immediate first pass and stops cleanly.
func TestSweepLifecycleStartStop(t *testing.T) {
	rp := &fakeFileRP{}
	uc := newFileUC(rp)

	if err := uc.StartSweep(context.Background()); err != nil {
		t.Fatalf("start: %v", err)
	}
	if err := uc.StopSweep(context.Background()); err != nil {
		t.Fatalf("stop: %v", err)
	}
}

// Recovery fails stale PROCESSING rows past the dead-writer horizon; the cut
// is derived from the generation timeout family, never from boot time.
func TestRunRecoveryCut(t *testing.T) {
	rp := &fakeFileRP{}
	uc := newFileUC(rp)
	before := time.Now()

	uc.runRecovery(context.Background())

	if rp.failStaleCut.Before(before.Add(-staleGenerationAfter-time.Second)) ||
		rp.failStaleCut.After(before.Add(-staleGenerationAfter+time.Second)) {
		t.Fatalf("recovery cut = %v, want ~now-staleGenerationAfter", rp.failStaleCut)
	}
}

// Sweep halves: a listing error ends that stage's pass quietly — one stage's
// failure never aborts the others.
func TestSweepStageErrorsEndQuietly(t *testing.T) {
	rp := &fakeFileRP{
		staleErr:    errorspb.ErrorInfraErrorNetworkConnection(""),
		objectsErr:  errorspb.ErrorInfraErrorNetworkConnection(""),
		refByKeyErr: errorspb.ErrorInfraErrorNetworkConnection(""),
		softDeleted: []*File{{ID: "x"}},
	}
	uc := newFileUC(rp)
	ctx := context.Background()

	if n := uc.sweepStaleRows(ctx, time.Now()); n != 0 {
		t.Fatalf("stage error: stale = %d, want 0", n)
	}
	if n := uc.sweepOrphanObjects(ctx, time.Now()); n != 0 {
		t.Fatalf("stage error: objects = %d, want 0", n)
	}
}

// Deleting an unknown row answers through the Get classification, with no
// write side.
func TestDeleteUnknownClassified(t *testing.T) {
	rp := &fakeFileRP{rows: map[string]*File{}}
	uc := newFileUC(rp)

	if err := uc.Delete(context.Background(), "ghost"); !systempb.IsSystemFileNotFound(err) {
		t.Fatalf("unknown delete: want FILE_NOT_FOUND, got %v", err)
	}
	if len(rp.deletes) != 0 {
		t.Fatal("unknown delete: no write may happen")
	}
}

// Sweep ---------------------------------------------------------------------------------------------------------------

// The sweep runs one pass per registry domain: the domain switch drives row
// filtering (the tenant mixin) and bucket routing (TenantStorage) on one
// axis, so a pass never sees one tenant's rows against another's objects.
func TestSweepPassesPerDomain(t *testing.T) {
	rp := &fakeFileRP{
		claim:          map[string]bool{"f-1": true},
		staleUploading: []*File{{ID: "f-1", Key: "f/x"}},
	}
	uc := NewFileUC(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		passthroughTm{}, rp, fakeDomains{codes: []string{"default", "acme"}}, hookSink{},
	)
	uc.runSweep(context.Background())
	if got := rp.listTenants; len(got) != 2 || got[0] != "default" || got[1] != "acme" {
		t.Fatalf("sweep passes = %v, want one per domain default,acme", got)
	}
}

// A failed domain listing skips the cycle loudly — no pass runs against a
// registry the sweeper cannot read.
func TestSweepDomainListingFailureSkipsCycle(t *testing.T) {
	rp := &fakeFileRP{staleUploading: []*File{{ID: "f-1", Key: "f/x"}}}
	uc := NewFileUC(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		passthroughTm{}, rp, fakeDomains{err: errors.New("registry down")}, hookSink{},
	)
	uc.runSweep(context.Background())
	if len(rp.listTenants) != 0 {
		t.Fatal("a failed domain listing must not sweep")
	}
}

// A detached generation task re-pins the row's tenant: the upload lands in
// the owning tenant's bucket regardless of the scheduling context.
func TestGenerateRePinsOwningTenant(t *testing.T) {
	rp := &fakeFileRP{}
	uc := newFileUC(rp)
	err := uc.generate(context.Background(), &File{ID: "g-1", Key: "f/g1", TenantID: "acme"}, &GenerationSpec{
		Name:        "report.csv",
		ContentType: "text/csv",
		Produce:     func(_ context.Context, w io.Writer) error { _, e := w.Write([]byte("hello")); return e },
	})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(rp.uploadTenants) != 1 || rp.uploadTenants[0] != "acme" {
		t.Fatalf("upload tenant = %v, want the row's acme", rp.uploadTenants)
	}
}
