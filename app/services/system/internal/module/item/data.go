package item

import (
	"context"
	"log/slog"

	"entgo.io/ent/dialect/sql"
	"roci.dev/fracdex"

	"cyber-ecosystem/shared-go/helper"
	"cyber-ecosystem/shared-go/utils"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"

	"cyber-ecosystem/app/services/system/internal/ent"
	"cyber-ecosystem/app/services/system/internal/ent/item"
	"cyber-ecosystem/app/services/system/internal/ent/predicate"
	"cyber-ecosystem/app/services/system/internal/platform"
	"cyber-ecosystem/app/services/system/internal/shared"
)

type itemRP struct {
	shared.RP
}

func NewItemRP(logger *slog.Logger, p *platform.Platform) ItemRP {
	return &itemRP{
		RP: shared.NewRP(logger.With("module", "module/item_rp"), p),
	}
}

// Repo ----------------------------------------------------------------------------------------------------------------

func (rp *itemRP) Create(ctx context.Context, a *Item) (*Item, error) {
	created, err := rp.Platform.GetClient(ctx).Item.Create().
		SetName(*a.Name).
		SetNillableDescription(a.Description).
		Save(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapItem(created), nil
}

func (rp *itemRP) Update(ctx context.Context, fieldsMask []string, a *Item) (*Item, error) {
	updater := rp.Platform.GetClient(ctx).Item.UpdateOneID(a.ID)
	helper.Handler{
		"name": {
			Condition: a.Name != nil,
			OnTrue:    func() { updater.SetName(*a.Name) },
			OnFalse:   func() {},
		},
		"description": {
			Condition: a.Description != nil,
			OnTrue:    func() { updater.SetDescription(*a.Description) },
			OnFalse:   func() { updater.SetDescription("") },
		},
		"status": {
			Condition: a.Status != nil,
			OnTrue:    func() { updater.SetStatus(*a.Status) },
			OnFalse:   func() {},
		},
	}.Emit(fieldsMask)

	updated, err := updater.Save(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapItem(updated), nil
}

func (rp *itemRP) UpdateStatus(ctx context.Context, id, status string) (*Item, error) {
	updated, err := rp.Platform.GetClient(ctx).Item.UpdateOneID(id).SetStatus(status).Save(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapItem(updated), nil
}

func (rp *itemRP) Delete(ctx context.Context, id string) (string, error) {
	if err := rp.Platform.GetClient(ctx).Item.DeleteOneID(id).Exec(ctx); err != nil {
		return "", rp.Platform.HandleEntError(err)
	}
	return id, nil
}

func (rp *itemRP) List(ctx context.Context, in *ItemListIn) (*ItemListOut, error) {
	query := rp.Platform.GetClient(ctx).Item.Query()
	helper.WherePtr(query, utils.FromTimestamp(in.CreatedAtA), item.CreatedAtGTE)
	helper.WherePtr(query, utils.FromTimestamp(in.CreatedAtZ), item.CreatedAtLTE)
	helper.WherePtr(query, utils.FromTimestamp(in.UpdatedAtA), item.UpdatedAtGTE)
	helper.WherePtr(query, utils.FromTimestamp(in.UpdatedAtZ), item.UpdatedAtLTE)
	helper.Where(query, in.ID != nil, func() predicate.Item { return item.IDEQ(*in.ID) })
	helper.Where(query, in.Name != nil, func() predicate.Item { return item.NameContainsFold(*in.Name) })
	helper.Where(query, in.Status != nil, func() predicate.Item { return item.StatusEQ(*in.Status) })
	helper.ApplyOrderBy(helper.ParseOrderBy(in.OrderBy), ent.Asc, ent.Desc, helper.FOMapping{
		"name":      func(sel helper.SQLSelector) { query.Order(sel(item.FieldName)) },
		"createdAt": func(sel helper.SQLSelector) { query.Order(sel(item.FieldCreatedAt)) },
		"updatedAt": func(sel helper.SQLSelector) { query.Order(sel(item.FieldUpdatedAt)) },
		"sort":      func(sel helper.SQLSelector) { query.Order(sel(item.FieldSort)) },
	})
	query.Order(func(s *sql.Selector) { s.OrderBy(s.C(item.FieldSort)) })

	total, offset, limit, err := helper.ApplyPagination(ctx, query, in.PageRequest,
		helper.NewPageConfig(helper.DefaultPageSize, helper.DefaultPageSizeUnlimit),
		errorspb.ErrorGeneralErrorPaginationInvalidArgument(""),
	)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	items, err := query.All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return &ItemListOut{
		PageResponse: helper.BuildPageResponse(total, offset, limit),
		List:         utils.SliceMap(items, mapItem),
	}, nil
}

func (rp *itemRP) Get(ctx context.Context, id string) (*Item, error) {
	d, err := rp.Platform.GetClient(ctx).Item.Get(ctx, id)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapItem(d), nil
}

func (rp *itemRP) Sort(ctx context.Context, id string, prevID, nextID *string) (*Item, error) {
	var prevSort, nextSort string
	client := rp.Platform.GetClient(ctx)

	if prevID != nil {
		d, err := client.Item.Get(ctx, *prevID)
		if err != nil {
			return nil, rp.Platform.HandleEntError(err)
		}
		prevSort = d.Sort
	}

	if nextID != nil {
		d, err := client.Item.Get(ctx, *nextID)
		if err != nil {
			return nil, rp.Platform.HandleEntError(err)
		}
		nextSort = d.Sort
	}

	newSort, err := fracdex.KeyBetween(prevSort, nextSort)
	if err != nil {
		return nil, err
	}

	updated, err := client.Item.UpdateOneID(id).SetSort(newSort).Save(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapItem(updated), nil
}

// Private -------------------------------------------------------------------------------------------------------------

func mapItem(d *ent.Item) *Item {
	return &Item{
		ID:          d.ID,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
		Name:        &d.Name,
		Description: &d.Description,
		Status:      &d.Status,
	}
}
