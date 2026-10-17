package filepresign

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/capability/storage"
	"cyber-ecosystem/shared-go/kratos/security"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/module/file"
)

// passthroughTm runs the unit inline.
type passthroughTm struct{}

func (passthroughTm) InTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// fakePresignRP answers the presign surface with per-call records and
// injectable storage state.
type fakePresignRP struct {
	FilePresignRP
	rows map[string]*file.File

	limits storage.Limits

	creates     int
	createdRow  *file.File
	createErr   error
	marked      []string
	deletedRows []string

	presignedUploads  []string
	presignedParts    int
	multipartSessions int
	uploadID          string

	listedParts  []storage.CompletedPart
	listErr      error
	uploadResets []string
	completeErr  error
	completed    int
	aborted      []string
	deletedObjs  []string

	statObj        *storage.ObjectInfo
	statErr        error
	abortErr       error
	statOnComplete bool
}

func (f *fakePresignRP) Create(_ context.Context, spec *file.File) (*file.File, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	f.creates++
	out := *spec
	out.ID = "f-new"
	f.createdRow = &out
	return &out, nil
}

func (f *fakePresignRP) MarkConfirmed(_ context.Context, id string, _ int64, _, _ string) (*file.File, error) {
	f.marked = append(f.marked, id)
	return f.rows[id], nil
}

func (f *fakePresignRP) Delete(_ context.Context, id string) error {
	f.deletedRows = append(f.deletedRows, id)
	return nil
}

func (f *fakePresignRP) FindByID(_ context.Context, id string) (*file.File, error) {
	if r, ok := f.rows[id]; ok {
		return r, nil
	}
	return nil, errorspb.ErrorInfraErrorDbNotFound("")
}

func (f *fakePresignRP) Limits() storage.Limits { return f.limits }

func (f *fakePresignRP) PresignUpload(context.Context, string, int64, time.Duration) (string, error) {
	f.presignedUploads = append(f.presignedUploads, "u")
	return "put-url", nil
}

func (f *fakePresignRP) PresignUploadPart(context.Context, string, string, int32, time.Duration) (string, error) {
	f.presignedParts++
	return "part-url", nil
}

func (f *fakePresignRP) CreateMultipart(context.Context, string, string) (string, error) {
	f.multipartSessions++
	if f.uploadID == "" {
		return "up-1", nil
	}
	return f.uploadID, nil
}

func (f *fakePresignRP) ListParts(context.Context, string, string) ([]storage.CompletedPart, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.listedParts, nil
}

func (f *fakePresignRP) SetUploadID(_ context.Context, id, uploadID string) error {
	f.uploadResets = append(f.uploadResets, id+"->"+uploadID)
	return nil
}

func (f *fakePresignRP) CompleteMultipart(context.Context, string, string, []storage.CompletedPart) error {
	f.completed++
	if f.statOnComplete && f.statObj == nil {
		f.statObj = &storage.ObjectInfo{Size: 100, ETag: "E", ContentType: "application/octet-stream"}
	}
	return f.completeErr
}

func (f *fakePresignRP) AbortMultipart(_ context.Context, _, uploadID string) error {
	f.aborted = append(f.aborted, uploadID)
	return f.abortErr
}

func (f *fakePresignRP) Stat(context.Context, string) (*storage.ObjectInfo, error) {
	if f.statErr != nil {
		return nil, f.statErr
	}
	if f.statObj == nil {
		// An object that never landed answers the classified miss the real
		// backend would — never a nil object with a nil error.
		return nil, errorspb.ErrorInfraErrorStorageNotFound("")
	}
	return f.statObj, nil
}

func (f *fakePresignRP) DeleteObject(_ context.Context, key string) error {
	f.deletedObjs = append(f.deletedObjs, key)
	return nil
}

func newPresignUC(rp *fakePresignRP) *FilePresignUC {
	return NewFilePresignUC(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		passthroughTm{}, rp,
	)
}

func userCtx(userID string) context.Context {
	return security.WithSubject(context.Background(), &security.Subject{UserID: userID})
}

func testLimits() storage.Limits {
	return storage.Limits{
		MaxFileSize:        1024,
		MultipartThreshold: 100,
		PartSize:           40,
		PresignTTL:         time.Minute,
	}
}

// Create --------------------------------------------------------------------------------------------------------------

// Below the multipart threshold a single PUT URL is minted; the row is born
// UPLOADING owned by the caller.
func TestCreateSingleUnderThreshold(t *testing.T) {
	rp := &fakePresignRP{limits: testLimits()}
	uc := newPresignUC(rp)

	out, err := uc.Create(userCtx("u1"), &CreateIn{Name: "a", ContentType: "text/plain", Size: 50})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if rp.multipartSessions != 0 || len(rp.presignedUploads) != 1 {
		t.Fatalf("create: sessions=%d uploads=%v", rp.multipartSessions, rp.presignedUploads)
	}
	if out.File.Status != file.StatusUploading || out.File.OwnerID != "u1" || out.Single == nil || out.Multipart != nil {
		t.Fatalf("create: out = %+v", out)
	}
}

// At/above the threshold the plan is multipart: session + per-part URLs, part
// count from the plan.
func TestCreateMultipartOverThreshold(t *testing.T) {
	rp := &fakePresignRP{limits: testLimits()} // size 100 / partSize 40 → 3 parts
	uc := newPresignUC(rp)

	out, err := uc.Create(userCtx("u1"), &CreateIn{Name: "a", Size: 100})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if rp.multipartSessions != 1 || len(rp.presignedUploads) != 0 {
		t.Fatalf("create: sessions=%d single=%v", rp.multipartSessions, rp.presignedUploads)
	}
	if out.Multipart == nil || out.Multipart.PartCount != 3 || out.Multipart.PartSize != 40 || len(out.Multipart.PartURLs) != 3 {
		t.Fatalf("create: plan = %+v", out.Multipart)
	}
}

// An explicit size cap rejects before anything is minted.
func TestCreateRejectsOversize(t *testing.T) {
	rp := &fakePresignRP{limits: testLimits()}
	uc := newPresignUC(rp)

	if _, err := uc.Create(userCtx("u1"), &CreateIn{Name: "a", Size: 2000}); !systempb.IsSystemFileTooLarge(err) {
		t.Fatalf("oversize: want FILE_TOO_LARGE, got %v", err)
	}
	if rp.creates != 0 {
		t.Fatal("oversize: no row may be born")
	}
}

// A row-create failure after the session opened aborts the session — no
// leaked multipart.
func TestCreateMultipartRollbackAbortsSession(t *testing.T) {
	rp := &fakePresignRP{limits: testLimits(), createErr: errorspb.ErrorInfraErrorDbConstraint("")}
	uc := newPresignUC(rp)

	if _, err := uc.Create(userCtx("u1"), &CreateIn{Name: "a", Size: 100}); !errorspb.IsInfraErrorDbConstraint(err) {
		t.Fatalf("create failure: want infra DbConstraint passthrough, got %v", err)
	}
	if len(rp.aborted) != 1 || rp.aborted[0] != "up-1" {
		t.Fatalf("orphan cleanup: aborted = %v", rp.aborted)
	}
}

// Confirm -------------------------------------------------------------------------------------------------------------

// A foreign row answers FILE_NOT_FOUND, not FORBIDDEN — the id is an
// unguessable xid and the answer leaks no existence either way.
func TestGetOwnedForeignIsNotFound(t *testing.T) {
	rp := &fakePresignRP{rows: map[string]*file.File{
		"f1": {ID: "f1", OwnerID: "someone-else", Status: file.StatusUploading},
	}}
	uc := newPresignUC(rp)

	if _, err := uc.Confirm(userCtx("u1"), "f1", nil); !systempb.IsSystemFileNotFound(err) {
		t.Fatalf("foreign row: want FILE_NOT_FOUND, got %v", err)
	}
	if _, err := uc.Confirm(userCtx("u1"), "ghost", nil); !systempb.IsSystemFileNotFound(err) {
		t.Fatalf("unknown row: want FILE_NOT_FOUND, got %v", err)
	}
}

// Confirming an already-confirmed row is idempotent — no storage round-trip.
func TestConfirmIdempotent(t *testing.T) {
	rp := &fakePresignRP{rows: map[string]*file.File{
		"f1": {ID: "f1", OwnerID: "u1", Status: file.StatusConfirmed, Key: "k1"},
	}}
	uc := newPresignUC(rp)

	out, err := uc.Confirm(userCtx("u1"), "f1", nil)
	if err != nil || out.Status != file.StatusConfirmed {
		t.Fatalf("idempotent confirm: out=%+v err=%v", out, err)
	}
	if rp.statObj != nil || rp.statErr != nil || len(rp.marked) != 0 {
		t.Fatal("idempotent confirm: no storage may be touched")
	}
}

// The object's size is the truth: a mismatch refuses to confirm.
func TestConfirmSizeMismatch(t *testing.T) {
	rp := &fakePresignRP{
		rows: map[string]*file.File{
			"f1": {ID: "f1", OwnerID: "u1", Status: file.StatusUploading, Key: "k1", Size: 100},
		},
		statObj: &storage.ObjectInfo{Size: 99, ETag: "e", ContentType: "text/plain"},
	}
	uc := newPresignUC(rp)

	if _, err := uc.Confirm(userCtx("u1"), "f1", nil); !systempb.IsSystemFileSizeMismatch(err) {
		t.Fatalf("size mismatch: want FILE_SIZE_MISMATCH, got %v", err)
	}
	if len(rp.marked) != 0 {
		t.Fatal("size mismatch: no confirm may be written")
	}
}

// Abort ----------------------------------------------------------------------------------------------------------------

// Only UPLOADING aborts — the exit of a confirmed row is delete, not abort.
func TestAbortOnlyUploading(t *testing.T) {
	rp := &fakePresignRP{rows: map[string]*file.File{
		"f-ok": {ID: "f-ok", OwnerID: "u1", Status: file.StatusConfirmed},
	}}
	uc := newPresignUC(rp)

	if err := uc.Abort(userCtx("u1"), "f-ok"); !systempb.IsSystemFileInvalidState(err) {
		t.Fatalf("confirmed abort: want FILE_INVALID_STATE, got %v", err)
	}
}

// A sessioned upload aborts the multipart session; a single-PUT upload
// deletes the object; both drop the row.
func TestAbortSessionAndSinglePaths(t *testing.T) {
	rp := &fakePresignRP{rows: map[string]*file.File{
		"f-mp": {ID: "f-mp", OwnerID: "u1", Status: file.StatusUploading, Key: "k-mp", UploadID: "up-9"},
		"f-1":  {ID: "f-1", OwnerID: "u1", Status: file.StatusUploading, Key: "k-1"},
	}}
	uc := newPresignUC(rp)

	if err := uc.Abort(userCtx("u1"), "f-mp"); err != nil {
		t.Fatalf("abort sessioned: %v", err)
	}
	if len(rp.aborted) != 1 || rp.aborted[0] != "up-9" {
		t.Fatalf("abort sessioned: aborted = %v", rp.aborted)
	}

	if err := uc.Abort(userCtx("u1"), "f-1"); err != nil {
		t.Fatalf("abort single: %v", err)
	}
	if len(rp.deletedObjs) != 1 || rp.deletedObjs[0] != "k-1" {
		t.Fatalf("abort single: objects = %v", rp.deletedObjs)
	}
	if len(rp.deletedRows) != 2 {
		t.Fatalf("abort: rows dropped = %v", rp.deletedRows)
	}
}

// ListParts (resume) ---------------------------------------------------------------------------------------------------

// A single-PUT session is the degenerate one-part plan: no part state to
// list — the object either landed (resume = confirm) or the whole object is
// re-minted.
func TestListPartsSingleResumeDegenerate(t *testing.T) {
	rp := &fakePresignRP{
		rows: map[string]*file.File{
			"f1": {ID: "f1", OwnerID: "u1", Status: file.StatusUploading, Key: "k1", Size: 50},
		},
		limits: testLimits(),
	}
	uc := newPresignUC(rp)

	// Object not landed: one missing "part" = the whole object.
	out, err := uc.ListParts(userCtx("u1"), "f1")
	if err != nil {
		t.Fatalf("resume miss: %v", err)
	}
	if out.PartSize != 50 || len(out.Missing) != 1 || out.Missing[0].PartNumber != 1 {
		t.Fatalf("resume miss: %+v", out)
	}

	// Object landed: the resume is a confirm-in-place.
	rp.statObj = &storage.ObjectInfo{Size: 50, ETag: "e"}
	out, err = uc.ListParts(userCtx("u1"), "f1")
	if err != nil {
		t.Fatalf("resume hit: %v", err)
	}
	if len(out.Uploaded) != 1 || out.Uploaded[0].PartNumber != 1 || len(out.Missing) != 0 {
		t.Fatalf("resume hit: %+v", out)
	}
}

// A misaligned uploaded part (wrong size on a non-final part, or an
// out-of-plan part number) refuses the resume.
func TestListPartsRejectsMisalignedParts(t *testing.T) {
	rp := &fakePresignRP{
		rows: map[string]*file.File{
			"f1": {ID: "f1", OwnerID: "u1", Status: file.StatusUploading, Key: "k1", UploadID: "up-1", Size: 100},
		},
		limits:      testLimits(),                                                  // partSize 40, partCount 3
		listedParts: []storage.CompletedPart{{PartNumber: 1, Size: 39, ETag: "e"}}, // non-final part with wrong size
	}
	uc := newPresignUC(rp)

	if _, err := uc.ListParts(userCtx("u1"), "f1"); !systempb.IsSystemFileInvalidState(err) {
		t.Fatalf("misaligned: want FILE_INVALID_STATE, got %v", err)
	}
}

// completeSession -----------------------------------------------------------------------------------------------------

// The client's claimed etags must match the backend's truth exactly, part
// for part — a mismatched or short claim refuses to complete.
func TestCompleteSessionVerifiesClaimedParts(t *testing.T) {
	// size 100 / partSize 40 → parts 1..3; backend truth uploaded fully.
	truth := []storage.CompletedPart{
		{PartNumber: 1, Size: 40, ETag: "e1"},
		{PartNumber: 2, Size: 40, ETag: "e2"},
		{PartNumber: 3, Size: 20, ETag: "e3"},
	}
	rp := &fakePresignRP{limits: testLimits(), listedParts: truth}
	uc := newPresignUC(rp)
	f := &file.File{Key: "k1", UploadID: "up-1", Size: 100}

	// Wrong etag on part 2.
	bad := []PartETag{{PartNumber: 1, ETag: "e1"}, {PartNumber: 2, ETag: "WRONG"}, {PartNumber: 3, ETag: "e3"}}
	if _, err := uc.completeSession(context.Background(), f, bad); !systempb.IsSystemFileInvalidState(err) {
		t.Fatalf("wrong etag: want FILE_INVALID_STATE, got %v", err)
	}

	// Short claim.
	short := []PartETag{{PartNumber: 1, ETag: "e1"}}
	if _, err := uc.completeSession(context.Background(), f, short); !systempb.IsSystemFileInvalidState(err) {
		t.Fatalf("short claim: want FILE_INVALID_STATE, got %v", err)
	}

	// Honest claim completes and stats the result.
	rp.statObj = &storage.ObjectInfo{Size: 100, ETag: "E", ContentType: "application/octet-stream"}
	good := []PartETag{{PartNumber: 1, ETag: "e1"}, {PartNumber: 2, ETag: "e2"}, {PartNumber: 3, ETag: "e3"}}
	obj, err := uc.completeSession(context.Background(), f, good)
	if err != nil || obj.ETag != "E" {
		t.Fatalf("good claim: obj=%+v err=%v", obj, err)
	}
	if rp.completed != 1 {
		t.Fatalf("good claim: completed = %d", rp.completed)
	}
}

// planOf --------------------------------------------------------------------------------------------------------------

// The plan sizes parts off the backend's part size, raising the floor when
// the S3 max-parts cap demands it.
func TestPlanOf(t *testing.T) {
	limits := storage.Limits{PartSize: 100}
	// 250 bytes / 100 = 3 parts.
	if ps, n := planOf(limits, 250); ps != 100 || n != 3 {
		t.Fatalf("plan(250) = %d/%d, want 100/3", ps, n)
	}
	// Exact multiple is exactly that many parts.
	if _, n := planOf(limits, 300); n != 3 {
		t.Fatalf("plan(300) parts = %d, want 3", n)
	}
	// Beyond the max-parts cap the part size floor rises: 10001 parts of 100
	// would exceed 10000, so the plan rounds up.
	if ps, n := planOf(limits, 10000*100+1); ps <= 100 || n > int32(10000) {
		t.Fatalf("plan(cap-buster) = %d/%d, want floor raised within cap", ps, n)
	}
}

// A single-PUT confirm stats the landed object, checks its size against the
// declared one, and writes the confirm.
func TestConfirmSingleByStat(t *testing.T) {
	rp := &fakePresignRP{
		rows: map[string]*file.File{
			"f1": {ID: "f1", OwnerID: "u1", Status: file.StatusUploading, Key: "k1", Size: 100},
		},
		statObj: &storage.ObjectInfo{Size: 100, ETag: "e1", ContentType: "text/plain"},
	}
	uc := newPresignUC(rp)

	out, err := uc.Confirm(userCtx("u1"), "f1", nil)
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if len(rp.marked) != 1 || rp.marked[0] != "f1" {
		t.Fatalf("confirm: marked = %v", rp.marked)
	}
	if out == nil {
		t.Fatal("confirm: out must be the marked row")
	}
}

// When the object has not landed, a single-PUT row (no session, no parts)
// cannot confirm — there is nothing on the backend to confirm against.
func TestConfirmSingleWithoutObjectRejected(t *testing.T) {
	rp := &fakePresignRP{
		rows: map[string]*file.File{
			"f1": {ID: "f1", OwnerID: "u1", Status: file.StatusUploading, Key: "k1", Size: 100},
		},
	}
	uc := newPresignUC(rp)

	if _, err := uc.Confirm(userCtx("u1"), "f1", nil); !systempb.IsSystemFileInvalidState(err) {
		t.Fatalf("no object: want FILE_INVALID_STATE, got %v", err)
	}
}

// A multipart confirm rides the session: the claimed parts are verified
// against backend truth, completed, statted, and only then confirmed.
func TestConfirmCompletesSessionEndToEnd(t *testing.T) {
	rp := &fakePresignRP{
		rows: map[string]*file.File{
			"f1": {ID: "f1", OwnerID: "u1", Status: file.StatusUploading, Key: "k1", UploadID: "up-1", Size: 100},
		},
		limits: testLimits(), // 3 parts of 40
		listedParts: []storage.CompletedPart{
			{PartNumber: 1, Size: 40, ETag: "e1"},
			{PartNumber: 2, Size: 40, ETag: "e2"},
			{PartNumber: 3, Size: 20, ETag: "e3"},
		},
		statOnComplete: true,
	}
	uc := newPresignUC(rp)

	parts := []PartETag{{PartNumber: 1, ETag: "e1"}, {PartNumber: 2, ETag: "e2"}, {PartNumber: 3, ETag: "e3"}}
	if _, err := uc.Confirm(userCtx("u1"), "f1", parts); err != nil {
		t.Fatalf("session confirm: %v", err)
	}
	if rp.completed != 1 || len(rp.marked) != 1 {
		t.Fatalf("session confirm: completed=%d marked=%v", rp.completed, rp.marked)
	}
}

// A backend error on abort or object delete surfaces — cleanup failures are
// not swallowed here.
func TestAbortSurfacesCleanupErrors(t *testing.T) {
	rp := &fakePresignRP{
		rows: map[string]*file.File{
			"f1": {ID: "f1", OwnerID: "u1", Status: file.StatusUploading, Key: "k1", UploadID: "up-1"},
		},
		abortErr: errorspb.ErrorInfraErrorNetworkConnection(""),
	}
	uc := newPresignUC(rp)

	if err := uc.Abort(userCtx("u1"), "f1"); !errorspb.IsInfraErrorNetworkConnection(err) {
		t.Fatalf("abort failure: want infra passthrough, got %v", err)
	}
}

// Resuming a multipart session mints URLs only for the missing parts; the
// uploaded set carries the backend's truth.
func TestListPartsMultipartMintsMissing(t *testing.T) {
	rp := &fakePresignRP{
		rows: map[string]*file.File{
			"f1": {ID: "f1", OwnerID: "u1", Status: file.StatusUploading, Key: "k1", UploadID: "up-1", Size: 100},
		},
		limits: testLimits(),
		listedParts: []storage.CompletedPart{
			{PartNumber: 1, Size: 40, ETag: "e1"},
		},
	}
	uc := newPresignUC(rp)

	out, err := uc.ListParts(userCtx("u1"), "f1")
	if err != nil {
		t.Fatalf("resume: %v", err)
	}
	if len(out.Uploaded) != 1 || len(out.Missing) != 2 {
		t.Fatalf("resume: uploaded=%d missing=%d, want 1/2", len(out.Uploaded), len(out.Missing))
	}
	if rp.presignedParts != 2 {
		t.Fatalf("resume: minted %d part URLs, want exactly the 2 missing", rp.presignedParts)
	}
}

// A multipart session the backend forgot restarts transparently: fresh id
// pinned to the row, every part missing, no error and no abort of the dead id.
func TestListPartsRestartsDeadSession(t *testing.T) {
	rp := &fakePresignRP{
		rows: map[string]*file.File{
			"f1": {ID: "f1", OwnerID: "u1", Status: file.StatusUploading, Key: "k1", UploadID: "up-dead", Size: 100},
		},
		limits:  testLimits(),
		listErr: errorspb.ErrorInfraErrorStorageNotFound(""),
	}
	uc := newPresignUC(rp)

	out, err := uc.ListParts(userCtx("u1"), "f1")
	if err != nil {
		t.Fatalf("restart: %v", err)
	}
	if rp.multipartSessions != 1 || len(rp.uploadResets) != 1 || rp.uploadResets[0] != "f1->up-1" {
		t.Fatalf("restart: sessions=%d resets=%v", rp.multipartSessions, rp.uploadResets)
	}
	if len(out.Uploaded) != 0 || len(out.Missing) != 3 || rp.presignedParts != 3 {
		t.Fatalf("restart: uploaded=%d missing=%d minted=%d, want 0/3/3", len(out.Uploaded), len(out.Missing), rp.presignedParts)
	}
	if len(rp.aborted) != 0 {
		t.Fatalf("restart: dead id must not be aborted, aborted=%v", rp.aborted)
	}
}

// No subject, no resume.
func TestListPartsRequiresSubject(t *testing.T) {
	uc := newPresignUC(&fakePresignRP{})

	if _, err := uc.ListParts(context.Background(), "f1"); !errorspb.IsGeneralErrorUnauthenticated(err) {
		t.Fatalf("no subject: want unauthenticated, got %v", err)
	}
}

// A misaligned backend truth inside completion refuses — the plan is the
// contract on both sides of the session.
func TestCompleteSessionRejectsMisalignedTruth(t *testing.T) {
	rp := &fakePresignRP{
		limits: testLimits(), // 3 parts of 40, size 100
		listedParts: []storage.CompletedPart{
			{PartNumber: 1, Size: 40, ETag: "e1"},
			{PartNumber: 2, Size: 35, ETag: "e2"}, // non-final part, wrong size
		},
	}
	uc := newPresignUC(rp)
	f := &file.File{Key: "k1", UploadID: "up-1", Size: 100}

	claim := []PartETag{{PartNumber: 1, ETag: "e1"}, {PartNumber: 2, ETag: "e2"}}
	if _, err := uc.completeSession(context.Background(), f, claim); !systempb.IsSystemFileInvalidState(err) {
		t.Fatalf("misaligned truth: want FILE_INVALID_STATE, got %v", err)
	}
}

// A complete that lost the race to a concurrent winner (session already
// closed) re-stats and lets the size check judge the winner's object.
func TestCompleteSessionRaceLostToWinner(t *testing.T) {
	rp := &fakePresignRP{
		limits: testLimits(),
		listedParts: []storage.CompletedPart{
			{PartNumber: 1, Size: 40, ETag: "e1"},
			{PartNumber: 2, Size: 40, ETag: "e2"},
			{PartNumber: 3, Size: 20, ETag: "e3"},
		},
		completeErr: errorspb.ErrorInfraErrorStorageNotFound(""),
		statObj:     &storage.ObjectInfo{Size: 100, ETag: "E", ContentType: "application/octet-stream"},
	}
	uc := newPresignUC(rp)
	f := &file.File{Key: "k1", UploadID: "up-1", Size: 100}

	claim := []PartETag{{PartNumber: 1, ETag: "e1"}, {PartNumber: 2, ETag: "e2"}, {PartNumber: 3, ETag: "e3"}}
	obj, err := uc.completeSession(context.Background(), f, claim)
	if err != nil || obj.ETag != "E" {
		t.Fatalf("race lost: obj=%+v err=%v, want the winner's object", obj, err)
	}
}

// No subject, no abort.
func TestAbortRequiresSubject(t *testing.T) {
	uc := newPresignUC(&fakePresignRP{})

	if err := uc.Abort(context.Background(), "f1"); !errorspb.IsGeneralErrorUnauthenticated(err) {
		t.Fatalf("no subject: want unauthenticated, got %v", err)
	}
}
