package sku

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/rs/xid"

	"cyber-ecosystem/shared-go/kratos/security"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"

	"cyber-ecosystem/app/services/esports/internal/module/compliance"
	"cyber-ecosystem/app/services/esports/internal/shared"
)

const (
	KindTrial      = "trial"
	KindMembership = "membership"
)

const skuCodePrefix = "sku_" // operator-minted rows; seed rows keep semantic codes (member_month) — a join key either way

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
	// Operator id of the last catalog write; nil on seed rows no operator
	// has touched (audit join).
	UpdatedBy *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SaveIn struct {
	// The catalog write minus the code — the mint owns that. Kind shape
	// rules (trial carries no window, membership must) answer at the save.
	Name         string
	Kind         string
	PriceFen     int64
	ListPriceFen *int64 // defaults to the selling price
	PeriodDays   *int
	Description  *string
}

// Port ----------------------------------------------------------------------------------------------------------------

type SkuRP interface {
	Save(ctx context.Context, s *Sku) (*Sku, error)                                                       // insert with the caller-minted code; a code collision fails as a constraint error
	UpdatePrice(ctx context.Context, code string, priceFen, listPriceFen int64, by *string) (*Sku, error) // nil, nil = no such code
	SetShelf(ctx context.Context, code string, onShelf bool, by *string) (*Sku, error)                    // nil, nil = no such code
	List(ctx context.Context) ([]*Sku, error)                                                             // on-shelf face for the member plane
	ListAll(ctx context.Context) ([]*Sku, error)                                                          // the whole catalog, shelf ignored — the operator face
	ByCode(ctx context.Context, code string) (*Sku, error)                                                // shelf ignored — nil, nil = no such code
}

// UC ------------------------------------------------------------------------------------------------------------------

type SkuUC struct {
	shared.UC
	skuRP        SkuRP
	complianceRP compliance.ComplianceRP
}

func NewSkuUC(logger *slog.Logger, tm shared.Transaction, skuRP SkuRP, complianceRP compliance.ComplianceRP) *SkuUC {
	return &SkuUC{UC: shared.NewUC(logger.With("module", "module/sku"), tm), skuRP: skuRP, complianceRP: complianceRP}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *SkuUC) Save(ctx context.Context, in *SaveIn) (*Sku, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok || subject == nil {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("sku save requires an admin session"))
	}
	if !kindShapeOK(in.Kind, in.PeriodDays) {
		return nil, esportspb.ErrorEsportsSkuKindInvalid("")
	}
	list := in.PriceFen
	if in.ListPriceFen != nil {
		list = *in.ListPriceFen
	}
	saved, err := uc.skuRP.Save(ctx, &Sku{
		Code: skuCodePrefix + xid.New().String(),
		Kind: in.Kind, Name: in.Name, Description: in.Description,
		PriceFen: in.PriceFen, ListPriceFen: list, PeriodDays: in.PeriodDays,
		OnShelf: true, UpdatedBy: &subject.UserID,
	})
	if err != nil {
		return nil, err
	}
	uc.Log.Info("sku created by operator", "code", saved.Code, "kind", saved.Kind, "price_fen", saved.PriceFen, "by", subject.UserID)
	return saved, nil
}

func (uc *SkuUC) UpdatePrice(ctx context.Context, code string, priceFen int64, listPriceFen *int64) (*Sku, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok || subject == nil {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("sku reprice requires an admin session"))
	}
	// The one home of the list-price default: absent keeps the selling
	// price. Save applies the same rule on the create side.
	list := priceFen
	if listPriceFen != nil {
		list = *listPriceFen
	}
	s, err := uc.skuRP.UpdatePrice(ctx, code, priceFen, list, &subject.UserID)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, esportspb.ErrorEsportsSkuNotFound("")
	}
	uc.Log.Info("sku price updated by operator", "code", code, "price_fen", priceFen, "list_price_fen", list, "by", subject.UserID)
	return s, nil
}

func (uc *SkuUC) SetShelf(ctx context.Context, code string, onShelf bool) (*Sku, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok || subject == nil {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("sku shelf requires an admin session"))
	}
	s, err := uc.skuRP.SetShelf(ctx, code, onShelf, &subject.UserID)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, esportspb.ErrorEsportsSkuNotFound("")
	}
	uc.Log.Info("sku shelf set by operator", "code", code, "on_shelf", onShelf, "by", subject.UserID)
	return s, nil
}

func (uc *SkuUC) List(ctx context.Context) ([]*Sku, error) {
	return uc.skuRP.List(ctx)
}

func (uc *SkuUC) PayGate(ctx context.Context) (*compliance.ComplianceConfig, error) {
	// The selling face answers the gate with a read it already owes the
	// wall — the member plane has no compliance read of its own, so the
	// catalog folds it (the member_until precedent: zero extra hop).
	return uc.complianceRP.Get(ctx)
}

func (uc *SkuUC) ListAll(ctx context.Context) ([]*Sku, error) {
	return uc.skuRP.ListAll(ctx)
}

// Private -------------------------------------------------------------------------------------------------------------

func kindShapeOK(kind string, periodDays *int) bool {
	switch kind {
	case KindTrial:
		return periodDays == nil
	case KindMembership:
		return periodDays != nil && *periodDays > 0
	default:
		return false
	}
}
