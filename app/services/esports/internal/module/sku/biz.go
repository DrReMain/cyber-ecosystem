package sku

import (
	"context"
	"log/slog"
	"time"

	"cyber-ecosystem/app/services/esports/internal/shared"
)

const (
	KindTrial      = "trial"
	KindMembership = "membership"
)

// DO ------------------------------------------------------------------------------------------------------------------

type Sku struct {
	Code         string
	Kind         string
	Name         string
	Description  *string
	PriceFen     int64
	ListPriceFen int64
	PeriodDays   *int
	OnShelf      bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Port ----------------------------------------------------------------------------------------------------------------

type SkuRP interface {
	List(ctx context.Context) ([]*Sku, error)              // on-shelf face for the member plane
	ByCode(ctx context.Context, code string) (*Sku, error) // shelf ignored — nil, nil = no such code
}

// UC ------------------------------------------------------------------------------------------------------------------

type SkuUC struct {
	shared.UC
	skuRP SkuRP
}

func NewSkuUC(logger *slog.Logger, tm shared.Transaction, skuRP SkuRP) *SkuUC {
	return &SkuUC{UC: shared.NewUC(logger.With("module", "module/sku"), tm), skuRP: skuRP}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *SkuUC) List(ctx context.Context) ([]*Sku, error) {
	return uc.skuRP.List(ctx)
}
