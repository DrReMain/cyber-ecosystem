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
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}
