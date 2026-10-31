package sku

import (
	"context"
	"log/slog"

	"cyber-ecosystem/shared-go/utils"

	"cyber-ecosystem/app/services/esports/internal/ent"
	"cyber-ecosystem/app/services/esports/internal/ent/sku"
	"cyber-ecosystem/app/services/esports/internal/platform"
	"cyber-ecosystem/app/services/esports/internal/shared"
)

// Repo ----------------------------------------------------------------------------------------------------------------

type skuRP struct {
	shared.RP
}

func NewSkuRP(logger *slog.Logger, p *platform.Platform) SkuRP {
	return &skuRP{RP: shared.NewRP(logger.With("module", "module/sku_rp"), p)}
}

// Method --------------------------------------------------------------------------------------------------------------

func (rp *skuRP) Save(ctx context.Context, s *Sku) (*Sku, error) {
	create := rp.Platform.GetClient(ctx).Sku.Create().
		SetCode(s.Code).
		SetKind(sku.Kind(s.Kind)).
		SetName(s.Name).
		SetPriceFen(s.PriceFen).
		SetListPriceFen(s.ListPriceFen).
		SetOnShelf(s.OnShelf).
		SetNillableUpdatedBy(s.UpdatedBy)
	if s.Description != nil {
		create = create.SetDescription(*s.Description)
	}
	if s.PeriodDays != nil {
		create = create.SetPeriodDays(*s.PeriodDays)
	}
	row, err := create.Save(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapSku(row), nil
}

func (rp *skuRP) UpdatePrice(ctx context.Context, code string, priceFen, listPriceFen int64, by *string) (*Sku, error) {
	affected, err := rp.Platform.GetClient(ctx).Sku.Update().
		Where(sku.CodeEQ(code)).
		SetPriceFen(priceFen).
		SetListPriceFen(listPriceFen).
		SetNillableUpdatedBy(by).
		Save(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	if affected == 0 {
		return nil, nil
	}
	return rp.ByCode(ctx, code)
}

func (rp *skuRP) SetShelf(ctx context.Context, code string, onShelf bool, by *string) (*Sku, error) {
	affected, err := rp.Platform.GetClient(ctx).Sku.Update().
		Where(sku.CodeEQ(code)).
		SetOnShelf(onShelf).
		SetNillableUpdatedBy(by).
		Save(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	if affected == 0 {
		return nil, nil
	}
	return rp.ByCode(ctx, code)
}

func (rp *skuRP) List(ctx context.Context) ([]*Sku, error) {
	rows, err := rp.Platform.GetClient(ctx).Sku.Query().
		Where(sku.OnShelfEQ(true)).
		Order(ent.Asc(sku.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return utils.SliceMap(rows, mapSku), nil
}

func (rp *skuRP) ListAll(ctx context.Context) ([]*Sku, error) {
	rows, err := rp.Platform.GetClient(ctx).Sku.Query().
		Order(ent.Asc(sku.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return utils.SliceMap(rows, mapSku), nil
}

func (rp *skuRP) ByCode(ctx context.Context, code string) (*Sku, error) {
	row, err := rp.Platform.GetClient(ctx).Sku.Query().
		Where(sku.CodeEQ(code)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapSku(row), nil
}

// Private -------------------------------------------------------------------------------------------------------------

func mapSku(row *ent.Sku) *Sku {
	return &Sku{
		Code:         row.Code,
		Kind:         string(row.Kind),
		Name:         row.Name,
		Description:  row.Description,
		PriceFen:     row.PriceFen,
		ListPriceFen: row.ListPriceFen,
		PeriodDays:   row.PeriodDays,
		OnShelf:      row.OnShelf,
		UpdatedBy:    row.UpdatedBy,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}
