package fileproxy

import (
	"bytes"
	"context"
	"log/slog"

	"cyber-ecosystem/shared-go/capability/storage"

	"cyber-ecosystem/app/services/system/internal/ent"
	"cyber-ecosystem/app/services/system/internal/module/file"
	"cyber-ecosystem/app/services/system/internal/platform"
	"cyber-ecosystem/app/services/system/internal/shared"
)

// Repo ----------------------------------------------------------------------------------------------------------------

type fileProxyRP struct {
	shared.RP
}

func NewFileProxyRP(logger *slog.Logger, p *platform.Platform) FileProxyRP {
	return &fileProxyRP{
		RP: shared.NewRP(logger.With("module", "module/fileproxy_rp"), p),
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (rp *fileProxyRP) Create(ctx context.Context, f *file.File) (*file.File, error) {
	created, err := rp.Platform.GetClient(ctx).File.Create().
		SetKey(f.Key).
		SetName(f.Name).
		SetContentType(f.ContentType).
		SetSize(f.Size).
		SetSource(f.Source).
		SetStatus(f.Status).
		SetUploadID(f.UploadID).
		SetEtag(f.ETag).
		SetOwnerID(f.OwnerID).
		Save(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapFile(created), nil
}

func (rp *fileProxyRP) DeleteObject(ctx context.Context, key string) error {
	if err := rp.Platform.GetStorage().Object.Delete(ctx, key); err != nil {
		return rp.Platform.HandleStorageError(err)
	}
	return nil
}

func (rp *fileProxyRP) Limits() storage.Limits {
	return rp.Platform.GetStorage().Limits()
}

func (rp *fileProxyRP) PutObject(ctx context.Context, key string, data []byte, contentType string) (*storage.ObjectInfo, error) {
	obj, err := rp.Platform.GetStorage().Object.Upload(ctx, key, bytes.NewReader(data), int64(len(data)), contentType)
	if err != nil {
		return nil, rp.Platform.HandleStorageError(err)
	}
	return obj, nil
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
