package filepresign

import (
	"context"
	"log/slog"
	"time"

	"cyber-ecosystem/shared-go/capability/storage"

	"cyber-ecosystem/app/services/system/internal/ent"
	"cyber-ecosystem/app/services/system/internal/ent/schema/local_mixins"
	"cyber-ecosystem/app/services/system/internal/module/file"
	"cyber-ecosystem/app/services/system/internal/platform"
	"cyber-ecosystem/app/services/system/internal/shared"
)

// Repo ----------------------------------------------------------------------------------------------------------------

type filePresignRP struct {
	shared.RP
}

func NewFilePresignRP(logger *slog.Logger, p *platform.Platform) FilePresignRP {
	return &filePresignRP{
		RP: shared.NewRP(logger.With("module", "module/filepresign_rp"), p),
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (rp *filePresignRP) Create(ctx context.Context, f *file.File) (*file.File, error) {
	created, err := rp.Platform.GetClient(ctx).File.Create().
		SetKey(f.Key).
		SetName(f.Name).
		SetContentType(f.ContentType).
		SetSize(f.Size).
		SetSource(f.Source).
		SetStatus(f.Status).
		SetUploadID(f.UploadID).
		SetOwnerID(f.OwnerID).
		Save(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapFile(created), nil
}

func (rp *filePresignRP) MarkConfirmed(ctx context.Context, id string, size int64, contentType, etag string) (*file.File, error) {
	updated, err := rp.Platform.GetClient(ctx).File.UpdateOneID(id).
		SetStatus(file.StatusConfirmed).
		SetSize(size).
		SetContentType(contentType).
		SetEtag(etag).
		Save(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapFile(updated), nil
}

func (rp *filePresignRP) Delete(ctx context.Context, id string) error {
	// Aborted sessions hard-delete: an uploading row has no recycle value.
	uctx := local_mixins.SkipSoftDelete(ctx)
	if err := rp.Platform.GetClient(uctx).File.DeleteOneID(id).Exec(uctx); err != nil {
		return rp.Platform.HandleEntError(err)
	}
	return nil
}

func (rp *filePresignRP) FindByID(ctx context.Context, id string) (*file.File, error) {
	d, err := rp.Platform.GetClient(ctx).File.Get(ctx, id)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapFile(d), nil
}

func (rp *filePresignRP) Limits() storage.Limits {
	return rp.Platform.GetStorage().Limits()
}

func (rp *filePresignRP) PresignUpload(ctx context.Context, key string, size int64, ttl time.Duration) (string, error) {
	url, err := rp.Platform.GetStorage().Presign.PresignUpload(ctx, key, size, ttl)
	if err != nil {
		return "", rp.Platform.HandleStorageError(err)
	}
	return url, nil
}

func (rp *filePresignRP) PresignUploadPart(ctx context.Context, key, uploadID string, partNum int32, ttl time.Duration) (string, error) {
	url, err := rp.Platform.GetStorage().Presign.PresignUploadPart(ctx, key, uploadID, partNum, ttl)
	if err != nil {
		return "", rp.Platform.HandleStorageError(err)
	}
	return url, nil
}

func (rp *filePresignRP) CreateMultipart(ctx context.Context, key, contentType string) (string, error) {
	uploadID, err := rp.Platform.GetStorage().Multipart.Create(ctx, key, contentType)
	if err != nil {
		return "", rp.Platform.HandleStorageError(err)
	}
	return uploadID, nil
}

func (rp *filePresignRP) ListParts(ctx context.Context, key, uploadID string) ([]storage.CompletedPart, error) {
	parts, err := rp.Platform.GetStorage().Multipart.ListParts(ctx, key, uploadID)
	if err != nil {
		return nil, rp.Platform.HandleStorageError(err)
	}
	return parts, nil
}

func (rp *filePresignRP) CompleteMultipart(ctx context.Context, key, uploadID string, parts []storage.CompletedPart) error {
	if _, err := rp.Platform.GetStorage().Multipart.Complete(ctx, key, uploadID, parts); err != nil {
		return rp.Platform.HandleStorageError(err)
	}
	return nil
}

func (rp *filePresignRP) AbortMultipart(ctx context.Context, key, uploadID string) error {
	if err := rp.Platform.GetStorage().Multipart.Abort(ctx, key, uploadID); err != nil {
		return rp.Platform.HandleStorageError(err)
	}
	return nil
}

func (rp *filePresignRP) Stat(ctx context.Context, key string) (*storage.ObjectInfo, error) {
	obj, err := rp.Platform.GetStorage().Object.Stat(ctx, key)
	if err != nil {
		return nil, rp.Platform.HandleStorageError(err)
	}
	return obj, nil
}

func (rp *filePresignRP) DeleteObject(ctx context.Context, key string) error {
	if err := rp.Platform.GetStorage().Object.Delete(ctx, key); err != nil {
		return rp.Platform.HandleStorageError(err)
	}
	return nil
}

// Private -------------------------------------------------------------------------------------------------------------

func mapFile(d *ent.File) *file.File {
	return &file.File{
		ID:          d.ID,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
		TenantID:    d.TenantID,
		Key:         d.Key,
		Name:        d.Name,
		ContentType: d.ContentType,
		Size:        d.Size,
		Source:      d.Source,
		Status:      d.Status,
		UploadID:    d.UploadID,
		ETag:        d.Etag,
		OwnerID:     d.OwnerID,
	}
}
