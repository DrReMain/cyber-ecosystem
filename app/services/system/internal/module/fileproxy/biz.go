package fileproxy

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/rs/xid"

	"cyber-ecosystem/shared-go/capability/storage"
	"cyber-ecosystem/shared-go/kratos/security"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/module/file"
	"cyber-ecosystem/app/services/system/internal/shared"
)

// DO ------------------------------------------------------------------------------------------------------------------

type CreateIn struct {
	Name        string
	ContentType string
	Data        []byte
}

// Port ----------------------------------------------------------------------------------------------------------------

type FileProxyRP interface {
	Create(ctx context.Context, f *file.File) (*file.File, error)
	DeleteObject(ctx context.Context, key string) error
	Limits() storage.Limits
	PutObject(ctx context.Context, key string, data []byte, contentType string) (*storage.ObjectInfo, error)
}

// UC ------------------------------------------------------------------------------------------------------------------

type FileProxyUC struct {
	shared.UC
	fileRP FileProxyRP
}

func NewFileProxyUC(logger *slog.Logger, tm shared.Transaction, fileRP FileProxyRP) *FileProxyUC {
	return &FileProxyUC{
		UC:     shared.NewUC(logger.With("module", "module/fileproxy"), tm),
		fileRP: fileRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *FileProxyUC) Create(ctx context.Context, in *CreateIn) (*file.File, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("no subject in context"))
	}
	limits := uc.fileRP.Limits()
	size := int64(len(in.Data))
	if limits.MaxFileSize > 0 && size > limits.MaxFileSize {
		return nil, systempb.ErrorSystemFileTooLarge("")
	}
	f := &file.File{
		Key:         newKey(),
		Name:        in.Name,
		ContentType: in.ContentType,
		Size:        size,
		Source:      file.SourceClientUpload,
		Status:      file.StatusConfirmed,
		OwnerID:     subject.UserID,
	}
	obj, err := uc.fileRP.PutObject(ctx, f.Key, in.Data, f.ContentType)
	if err != nil {
		return nil, err
	}
	f.ETag = obj.ETag
	created, err := uc.fileRP.Create(ctx, f)
	if err != nil {
		if dErr := uc.fileRP.DeleteObject(ctx, f.Key); dErr != nil {
			uc.Log.Warn("proxy upload orphan cleanup failed", "key", f.Key, "error", dErr)
		}
		return nil, err
	}
	return created, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func newKey() string {
	return "f/" + xid.New().String()
}
