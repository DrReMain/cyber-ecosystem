package file

import (
	"context"
	"log/slog"
	"time"

	"cyber-ecosystem/shared-go/capability/storage"
	"cyber-ecosystem/shared-go/helper"
	"cyber-ecosystem/shared-go/utils"

	"cyber-ecosystem/app/services/system/internal/ent"
	entfile "cyber-ecosystem/app/services/system/internal/ent/file"
	"cyber-ecosystem/app/services/system/internal/ent/predicate"
	"cyber-ecosystem/app/services/system/internal/ent/schema/local_mixins"
	"cyber-ecosystem/app/services/system/internal/platform"
	"cyber-ecosystem/app/services/system/internal/shared"
)

// Repo ----------------------------------------------------------------------------------------------------------------

type fileRP struct {
	shared.RP
}

func NewFileRP(logger *slog.Logger, p *platform.Platform) FileRP {
	return &fileRP{
		RP: shared.NewRP(logger.With("module", "module/file_rp"), p),
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (rp *fileRP) UpdateName(ctx context.Context, id, name string) (*File, error) {
	updated, err := rp.Platform.GetClient(ctx).File.UpdateOneID(id).SetName(name).Save(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapFile(updated), nil
}

func (rp *fileRP) Delete(ctx context.Context, id string) error {
	if err := rp.Platform.GetClient(ctx).File.DeleteOneID(id).Exec(ctx); err != nil {
		return rp.Platform.HandleEntError(err)
	}
	return nil
}

func (rp *fileRP) List(ctx context.Context, in *FileListIn) (*FileListOut, error) {
	query := rp.Platform.GetClient(ctx).File.Query()
	helper.Where(query, in.Name != nil, func() predicate.File { return entfile.NameContainsFold(*in.Name) })
	helper.Where(query, in.ContentType != nil, func() predicate.File { return entfile.ContentTypeEQ(*in.ContentType) })
	helper.Where(query, in.Status != nil, func() predicate.File { return entfile.StatusEQ(*in.Status) })
	helper.Where(query, in.Source != nil, func() predicate.File { return entfile.SourceEQ(*in.Source) })
	helper.ApplyOrderBy(helper.ParseOrderBy(in.OrderBy), ent.Asc, ent.Desc, helper.FOMapping{
		"name":      func(sel helper.SQLSelector) { query.Order(sel(entfile.FieldName)) },
		"size":      func(sel helper.SQLSelector) { query.Order(sel(entfile.FieldSize)) },
		"createdAt": func(sel helper.SQLSelector) { query.Order(sel(entfile.FieldCreatedAt)) },
		"updatedAt": func(sel helper.SQLSelector) { query.Order(sel(entfile.FieldUpdatedAt)) },
	})
	total, offset, limit, err := shared.Paginate(ctx, query, in.PageRequest, helper.DefaultPageSizeUnlimit)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	fs, err := query.All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return &FileListOut{
		PageResponse: helper.BuildPageResponse(total, offset, limit),
		List:         utils.SliceMap(fs, mapFile),
	}, nil
}

func (rp *fileRP) FindByID(ctx context.Context, id string) (*File, error) {
	d, err := rp.Platform.GetClient(ctx).File.Get(ctx, id)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapFile(d), nil
}

func (rp *fileRP) FindRef(ctx context.Context, id string) (*File, error) {
	uctx := local_mixins.Unscoped(ctx)
	d, err := rp.Platform.GetClient(uctx).File.Get(uctx, id)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapFile(d), nil
}

func (rp *fileRP) Limits() storage.Limits {
	return rp.Platform.GetStorage().Limits()
}

func (rp *fileRP) PresignDownload(ctx context.Context, key string, ttl time.Duration, opts storage.DownloadOptions) (string, error) {
	url, err := rp.Platform.GetStorage().Presign.PresignDownload(ctx, key, ttl, opts)
	if err != nil {
		return "", rp.Platform.HandleStorageError(err)
	}
	return url, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func mapFile(d *ent.File) *File {
	return &File{
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
