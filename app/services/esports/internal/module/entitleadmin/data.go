package entitleadmin

import (
	"context"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"

	"cyber-ecosystem/app/services/esports/internal/module/sku"
)

// Repo ----------------------------------------------------------------------------------------------------------------

type skuFactsAdapter struct {
	inner sku.SkuRP
}

func NewSkuFactsRP(inner sku.SkuRP) SkuFactsRP {
	// Adapts the sku catalog into the grant's narrow needs: the one kind
	// it will sell by hand, and the window length that kind carries.
	return &skuFactsAdapter{inner: inner}
}

// Method --------------------------------------------------------------------------------------------------------------

func (a *skuFactsAdapter) MembershipSku(ctx context.Context, code string) (*SkuFacts, error) {
	s, err := a.inner.ByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, esportspb.ErrorEsportsSkuNotFound("")
	}
	if s.Kind != sku.KindMembership || s.PeriodDays == nil {
		// The hand-sell face sells windows only; a trial sku has nothing to
		// grant here and its permanent unlock rides the settle path.
		return nil, esportspb.ErrorEsportsSkuKindInvalid("")
	}
	return &SkuFacts{Name: s.Name, PeriodDays: *s.PeriodDays}, nil
}
