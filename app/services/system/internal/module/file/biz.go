package file

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"cyber-ecosystem/shared-go/capability/storage"

	commonpb "cyber-ecosystem/gen/go/cyber/shared/common/v1"
	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
	systempb "cyber-ecosystem/gen/go/cyber/system/v1"

	"cyber-ecosystem/app/services/system/internal/shared"
)

const (
	StatusUploading = "uploading"
	StatusConfirmed = "confirmed"

	SourceClientUpload    = "client_upload"
	SourceServerGenerated = "server_generated"
)

// DO ------------------------------------------------------------------------------------------------------------------

type File struct {
	ID          string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	TenantID    string
	Key         string
	Name        string
	ContentType string
	Size        int64
	Source      string
	Status      string
	UploadID    string
	ETag        string
	OwnerID     string
}

type FileListIn struct {
	*commonpb.PageRequest
	OrderBy     []string
	Name        *string
	ContentType *string
	Status      *string
	Source      *string
}

type FileListOut struct {
	*commonpb.PageResponse
	List []*File
}

type UrlMint struct {
	ID        string
	URL       string
	ExpiresAt time.Time
}

// Port ----------------------------------------------------------------------------------------------------------------

type FileRP interface {
	UpdateName(ctx context.Context, id, name string) (*File, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, in *FileListIn) (*FileListOut, error)
	FindByID(ctx context.Context, id string) (*File, error)
	FindRef(ctx context.Context, id string) (*File, error)
	Limits() storage.Limits
	PresignDownload(ctx context.Context, key string, ttl time.Duration, opts storage.DownloadOptions) (string, error)
}

// UC ------------------------------------------------------------------------------------------------------------------

type FileUC struct {
	shared.UC
	fileRP FileRP
}

func NewFileUC(logger *slog.Logger, tm shared.Transaction, fileRP FileRP) *FileUC {
	return &FileUC{
		UC:     shared.NewUC(logger.With("module", "module/file"), tm),
		fileRP: fileRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *FileUC) Update(ctx context.Context, id, name string) (*File, error) {
	if _, err := uc.Get(ctx, id); err != nil {
		return nil, err
	}
	return uc.fileRP.UpdateName(ctx, id, name)
}

func (uc *FileUC) Delete(ctx context.Context, id string) error {
	f, err := uc.Get(ctx, id)
	if err != nil {
		return err
	}
	// An UPLOADING row belongs to the upload channel's lifecycle (its exit is
	// abort); the metadata plane only retires confirmed files.
	if f.Status != StatusConfirmed {
		return systempb.ErrorSystemFileInvalidState("")
	}
	return uc.fileRP.Delete(ctx, id)
}

func (uc *FileUC) List(ctx context.Context, in *FileListIn) (*FileListOut, error) {
	return uc.fileRP.List(ctx, in)
}

func (uc *FileUC) Get(ctx context.Context, id string) (*File, error) {
	f, err := uc.fileRP.FindByID(ctx, id)
	if err != nil {
		if errorspb.IsInfraErrorDbNotFound(err) {
			return nil, systempb.ErrorSystemFileNotFound("")
		}
		return nil, err
	}
	return f, nil
}

func (uc *FileUC) MintUrls(ctx context.Context, ids []string) ([]*UrlMint, error) {
	ttl := uc.fileRP.Limits().PresignTTL
	mints := make([]*UrlMint, 0, len(ids))
	for _, id := range ids {
		f, err := uc.fileRP.FindRef(ctx, id)
		if err != nil {
			if errorspb.IsInfraErrorDbNotFound(err) {
				continue
			}
			return nil, err
		}
		if f.Status != StatusConfirmed {
			continue
		}
		url, err := uc.fileRP.PresignDownload(ctx, f.Key, ttl, storage.DownloadOptions{
			ContentType:        f.ContentType,
			ContentDisposition: fmt.Sprintf("inline; filename=%q", f.Name),
		})
		if err != nil {
			return nil, err
		}
		mints = append(mints, &UrlMint{ID: id, URL: url, ExpiresAt: time.Now().Add(ttl)})
	}
	return mints, nil
}
