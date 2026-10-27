package fileproxy

import (
	"context"
	"io"
	"log/slog"
	"testing"

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

// fakeProxyRP records the proxy write side.
type fakeProxyRP struct {
	FileProxyRP
	limits      storage.Limits
	putObj      *storage.ObjectInfo
	putErr      error
	creates     int
	created     *file.File
	createErr   error
	deletedObjs []string
}

func (f *fakeProxyRP) Create(_ context.Context, spec *file.File) (*file.File, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	f.creates++
	out := *spec
	out.ID = "f-new"
	f.created = &out
	return &out, nil
}

func (f *fakeProxyRP) DeleteObject(_ context.Context, key string) error {
	f.deletedObjs = append(f.deletedObjs, key)
	return nil
}

func (f *fakeProxyRP) Limits() storage.Limits { return f.limits }

func (f *fakeProxyRP) PutObject(context.Context, string, []byte, string) (*storage.ObjectInfo, error) {
	if f.putErr != nil {
		return nil, f.putErr
	}
	return f.putObj, nil
}

func newProxyUC(rp *fakeProxyRP) *FileProxyUC {
	return NewFileProxyUC(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		passthroughTm{}, rp,
	)
}

func userCtx(userID string) context.Context {
	return security.WithSubject(context.Background(), &security.Subject{UserID: userID})
}

// The proxy path lands the object first, then the row — born CONFIRMED with
// the backend's etag, owned by the caller.
func TestCreateHappy(t *testing.T) {
	rp := &fakeProxyRP{
		limits: storage.Limits{MaxFileSize: 100},
		putObj: &storage.ObjectInfo{ETag: "e1"},
	}
	uc := newProxyUC(rp)

	out, err := uc.Create(userCtx("u1"), &CreateIn{Name: "a", ContentType: "text/plain", Data: []byte("hello")})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if rp.creates != 1 || out.Status != file.StatusConfirmed || out.ETag != "e1" || out.OwnerID != "u1" || out.Size != 5 {
		t.Fatalf("create: out = %+v", out)
	}
}

// The size cap applies to the buffered body too.
func TestCreateRejectsOversize(t *testing.T) {
	rp := &fakeProxyRP{limits: storage.Limits{MaxFileSize: 2}}
	uc := newProxyUC(rp)

	if _, err := uc.Create(userCtx("u1"), &CreateIn{Name: "a", Data: []byte("hello")}); !systempb.IsSystemFileTooLarge(err) {
		t.Fatalf("oversize: want FILE_TOO_LARGE, got %v", err)
	}
	if rp.creates != 0 {
		t.Fatal("oversize: no row may be born")
	}
}

// A row failure after the object landed cleans the object up — the proxy
// path owns the orphan it almost made.
func TestCreateRollbackCleansObject(t *testing.T) {
	rp := &fakeProxyRP{
		limits:    storage.Limits{},
		putObj:    &storage.ObjectInfo{ETag: "e1"},
		createErr: errorspb.ErrorInfraErrorDbConstraint(""),
	}
	uc := newProxyUC(rp)

	if _, err := uc.Create(userCtx("u1"), &CreateIn{Name: "a", Data: []byte("x")}); !errorspb.IsInfraErrorDbConstraint(err) {
		t.Fatalf("row failure: want infra DbConstraint passthrough, got %v", err)
	}
	if len(rp.deletedObjs) != 1 {
		t.Fatalf("orphan cleanup: objects = %v", rp.deletedObjs)
	}
}
