package data

import (
	"context"
	"io"

	"github.com/rs/xid"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/apps/template/services/base/internal/biz"
	"cyber-ecosystem/apps/template/services/base/internal/ent"
	"cyber-ecosystem/apps/template/services/base/internal/platform"
)

type fileRP struct {
	RP
}

func NewFileRP(logger log.Logger, p *platform.Platform) biz.FileRP {
	return &fileRP{
		RP: RP{
			log:      log.NewHelper(log.With(logger, "module", "data/file_rp")),
			platform: p,
		},
	}
}

// region[rgba(0,188,212,0.12)] 🩵 Repo --------------------------------------------------------------------------------

func (rp *fileRP) Upload(ctx context.Context, name string, reader io.Reader, size int64, contentType string) (*biz.File, error) {
	id := xid.New().String()

	info, err := rp.platform.GetStorage().Upload(ctx, id, name, reader, size, contentType)
	if err != nil {
		return nil, rp.platform.HandleStorageError(err)
	}

	created, err := rp.platform.GetClient(ctx).File.Create().
		SetID(id).
		SetName(info.Name).
		SetContentType(info.ContentType).
		SetSize(info.Size).
		SetStatus("attached").
		Save(ctx)
	if err != nil {
		_ = rp.platform.GetStorage().Delete(context.Background(), id)
		return nil, rp.platform.HandleEntError(err)
	}

	return mapFile(created), nil
}

func (rp *fileRP) Download(ctx context.Context, id string) (io.ReadCloser, *biz.File, error) {
	f, err := rp.platform.GetClient(ctx).File.Get(ctx, id)
	if err != nil {
		return nil, nil, rp.platform.HandleEntError(err)
	}

	body, _, err := rp.platform.GetStorage().Download(ctx, id)
	if err != nil {
		return nil, nil, rp.platform.HandleStorageError(err)
	}

	return body, mapFile(f), nil
}

func (rp *fileRP) Delete(ctx context.Context, id string) error {
	if err := rp.platform.GetStorage().Delete(context.Background(), id); err != nil {
		return rp.platform.HandleStorageError(err)
	}

	if err := rp.platform.GetClient(ctx).File.DeleteOneID(id).Exec(ctx); err != nil {
		rp.log.Errorf("S3 object %s deleted but metadata removal failed: %v", id, err)
		return rp.platform.HandleEntError(err)
	}
	return nil
}

// region[rgba(144,164,174,0.10)] ⚪ Private ---------------------------------------------------------------------------

func mapFile(f *ent.File) *biz.File {
	return &biz.File{
		ID:          f.ID,
		CreatedAt:   &f.CreatedAt,
		UpdatedAt:   &f.UpdatedAt,
		Name:        f.Name,
		ContentType: f.ContentType,
		Size:        f.Size,
		Status:      f.Status,
	}
}
