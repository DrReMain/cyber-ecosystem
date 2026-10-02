package file

import (
	"context"
	"errors"
	"io"
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
	return MapFile(updated), nil
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
		List:         utils.SliceMap(fs, MapFile),
	}, nil
}

func (rp *fileRP) FindByID(ctx context.Context, id string) (*File, error) {
	d, err := rp.Platform.GetClient(ctx).File.Get(ctx, id)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return MapFile(d), nil
}

func (rp *fileRP) FindRef(ctx context.Context, id string) (*File, error) {
	uctx := local_mixins.Unscoped(ctx)
	d, err := rp.Platform.GetClient(uctx).File.Get(uctx, id)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return MapFile(d), nil
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

func (rp *fileRP) CreateGeneration(ctx context.Context, f *File) (*File, error) {
	created, err := rp.Platform.GetClient(ctx).File.Create().
		SetKey(f.Key).
		SetName(f.Name).
		SetContentType(f.ContentType).
		SetSource(f.Source).
		SetStatus(f.Status).
		SetOwnerID(f.OwnerID).
		Save(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return MapFile(created), nil
}

func (rp *fileRP) UploadObject(ctx context.Context, key, contentType string, r io.Reader) (*storage.ObjectInfo, error) {
	// size<0: unknown at produce time, enforced by the streaming counter.
	info, err := rp.Platform.GetStorage().Object.Upload(ctx, key, r, -1, contentType)
	if err != nil {
		return nil, rp.Platform.HandleStorageError(err)
	}
	return info, nil
}

func (rp *fileRP) DeleteObject(ctx context.Context, key string) error {
	if err := rp.Platform.GetStorage().Object.Delete(ctx, key); err != nil {
		return rp.Platform.HandleStorageError(err)
	}
	return nil
}

func (rp *fileRP) MarkGenerated(ctx context.Context, id string, size int64, etag string) error {
	_, err := rp.Platform.GetClient(ctx).File.UpdateOneID(id).
		SetStatus(StatusConfirmed).
		SetSize(size).
		SetEtag(etag).
		Save(ctx)
	if err != nil {
		return rp.Platform.HandleEntError(err)
	}
	return nil
}

func (rp *fileRP) MarkFailed(ctx context.Context, id string) error {
	_, err := rp.Platform.GetClient(ctx).File.UpdateOneID(id).
		SetStatus(StatusFailed).
		Save(ctx)
	if err != nil {
		return rp.Platform.HandleEntError(err)
	}
	return nil
}

func (rp *fileRP) FailStaleProcessing(ctx context.Context, createdBefore time.Time) (int, error) {
	n, err := rp.Platform.GetClient(ctx).File.Update().
		Where(
			entfile.StatusEQ(StatusProcessing),
			entfile.CreatedAtLT(createdBefore),
		).
		SetStatus(StatusFailed).
		Save(ctx)
	if err != nil {
		return 0, rp.Platform.HandleEntError(err)
	}
	return n, nil
}

func (rp *fileRP) ListStaleUploading(ctx context.Context, createdBefore time.Time, afterID string, limit int) ([]*File, error) {
	rows, err := rp.Platform.GetClient(ctx).File.Query().
		Where(
			entfile.StatusEQ(StatusUploading),
			entfile.CreatedAtLT(createdBefore),
			entfile.IDGT(afterID),
		).
		Order(ent.Asc(entfile.FieldID)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return utils.SliceMap(rows, MapFile), nil
}

func (rp *fileRP) ClaimUploading(ctx context.Context, id string, createdBefore time.Time) (bool, error) {
	// The atomic claim: hard delete only when the row is still a stale
	// uploading one. Soft delete is bypassed on purpose — a swept row must
	// vanish, not linger as a tombstone. Both ctx sites need the bypass:
	// hooks read the mutation ctx, GetClient only resolves the connection.
	uctx := local_mixins.SkipSoftDelete(ctx)
	n, err := rp.Platform.GetClient(uctx).File.Delete().
		Where(
			entfile.ID(id),
			entfile.StatusEQ(StatusUploading),
			entfile.CreatedAtLT(createdBefore),
		).
		Exec(uctx)
	if err != nil {
		return false, rp.Platform.HandleEntError(err)
	}
	return n > 0, nil
}

func (rp *fileRP) FindByKey(ctx context.Context, key string) (*File, error) {
	d, err := rp.Platform.GetClient(ctx).File.Query().
		Where(entfile.KeyEQ(key)).
		Only(ctx)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return MapFile(d), nil
}

func (rp *fileRP) ListMultipartUploads(ctx context.Context, prefix string, visit func(storage.PendingUpload) error) error {
	if err := rp.Platform.GetStorage().Multipart.ListMultipartUploads(ctx, prefix, visit); err != nil {
		return rp.Platform.HandleStorageError(err)
	}
	return nil
}

func (rp *fileRP) AbortSession(ctx context.Context, key, uploadID string) error {
	err := rp.Platform.GetStorage().Multipart.Abort(ctx, key, uploadID)
	// Already completed or aborted is the sweep's goal state, not a failure.
	if errors.Is(err, storage.ErrNotFound) {
		return nil
	}
	if err != nil {
		return rp.Platform.HandleStorageError(err)
	}
	return nil
}

func (rp *fileRP) ListObjects(ctx context.Context, prefix, cursor string, maxKeys int32) (*storage.ListResult, error) {
	res, err := rp.Platform.GetStorage().List.List(ctx, prefix, cursor, maxKeys)
	if err != nil {
		return nil, rp.Platform.HandleStorageError(err)
	}
	return res, nil
}

func (rp *fileRP) FindRefByKey(ctx context.Context, key string) (*File, error) {
	uctx := local_mixins.Unscoped(ctx)
	d, err := rp.Platform.GetClient(uctx).File.Query().
		Where(entfile.KeyEQ(key)).
		Only(uctx)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return MapFile(d), nil
}

func (rp *fileRP) ListSoftDeleted(ctx context.Context, deletedBefore time.Time, afterID string, limit int) ([]*File, error) {
	uctx := local_mixins.SkipSoftDelete(ctx)
	rows, err := rp.Platform.GetClient(uctx).File.Query().
		Where(
			entfile.DeletedAtLT(deletedBefore),
			entfile.IDGT(afterID),
		).
		Order(ent.Asc(entfile.FieldID)).
		Limit(limit).
		All(uctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return utils.SliceMap(rows, MapFile), nil
}

func (rp *fileRP) ReclaimDeleted(ctx context.Context, id string, deletedBefore time.Time) (bool, error) {
	uctx := local_mixins.SkipSoftDelete(ctx)
	n, err := rp.Platform.GetClient(uctx).File.Delete().
		Where(
			entfile.ID(id),
			entfile.DeletedAtLT(deletedBefore),
		).
		Exec(uctx)
	if err != nil {
		return false, rp.Platform.HandleEntError(err)
	}
	return n > 0, nil
}

func (rp *fileRP) InAmbientTx(ctx context.Context) bool {
	return ent.TxFromContext(ctx) != nil
}

// Private -------------------------------------------------------------------------------------------------------------

func MapFile(d *ent.File) *File {
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
