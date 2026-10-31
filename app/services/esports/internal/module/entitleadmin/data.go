package entitleadmin

import (
	"context"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"

	"cyber-ecosystem/app/services/esports/internal/module/pay"
)

// Repo ----------------------------------------------------------------------------------------------------------------

type skuFactsAdapter struct {
	inner pay.SkuRP
}

func NewSkuFactsRP(inner pay.SkuRP) SkuFactsRP {
	// Adapts the pay catalog into the grant's narrow needs: the one kind
	// it will sell by hand, and the window length that kind carries.
	return &skuFactsAdapter{inner: inner}
}

// Method --------------------------------------------------------------------------------------------------------------

func (a *skuFactsAdapter) MembershipSku(ctx context.Context, code string) (*SkuFacts, error) {
	sku, err := a.inner.ByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if sku == nil {
		return nil, esportspb.ErrorEsportsSkuNotFound("")
	}
	if sku.Kind != pay.SkuKindMembership || sku.PeriodDays == nil {
		// The hand-sell face sells windows only; a trial sku has nothing to
		// grant here and its permanent unlock rides the settle path.
		return nil, esportspb.ErrorEsportsSkuKindInvalid("")
	}
	return &SkuFacts{Name: sku.Name, PeriodDays: *sku.PeriodDays}, nil
}
