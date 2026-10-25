package pay

import (
	"context"
	"log/slog"

	"cyber-ecosystem/shared-go/capability/payment"
	"cyber-ecosystem/shared-go/helper"
	"cyber-ecosystem/shared-go/utils"

	"cyber-ecosystem/app/services/esports/internal/ent"
	"cyber-ecosystem/app/services/esports/internal/ent/entitlement"
	"cyber-ecosystem/app/services/esports/internal/ent/payorder"
	"cyber-ecosystem/app/services/esports/internal/ent/sku"
	"cyber-ecosystem/app/services/esports/internal/platform"
	"cyber-ecosystem/app/services/esports/internal/shared"
)

// Repo ----------------------------------------------------------------------------------------------------------------

type skuRP struct {
	shared.RP
}

func NewSkuRP(logger *slog.Logger, p *platform.Platform) SkuRP {
	return &skuRP{RP: shared.NewRP(logger.With("module", "module/pay_sku_rp"), p)}
}

type payOrderRP struct {
	shared.RP
}

func NewPayOrderRP(logger *slog.Logger, p *platform.Platform) OrderRP {
	return &payOrderRP{RP: shared.NewRP(logger.With("module", "module/pay_order_rp"), p)}
}

type entitlementRP struct {
	shared.RP
}

func NewEntitlementRP(logger *slog.Logger, p *platform.Platform) EntitlementRP {
	return &entitlementRP{RP: shared.NewRP(logger.With("module", "module/pay_ent_rp"), p)}
}

type payChannelRP struct {
	shared.RP
}

func NewPayChannelRP(logger *slog.Logger, p *platform.Platform) PayChannelRP {
	return &payChannelRP{RP: shared.NewRP(logger.With("module", "module/pay_channel_rp"), p)}
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

func (rp *payOrderRP) Create(ctx context.Context, o *Order) (*Order, error) {
	create := rp.Platform.GetClient(ctx).PayOrder.Create().
		SetOrderNo(o.OrderNo).
		SetUserID(o.UserID).
		SetSkuCode(o.SKUCode).
		SetScopeKind(payorder.ScopeKind(o.ScopeKind)).
		SetProvider(o.Provider).
		SetChannel(o.Channel).
		SetAmountFen(o.AmountFen).
		SetState(payorder.State(o.State)).
		SetCredentialKind(o.CredentialKind).
		SetCredentialValue(o.CredentialValue).
		SetExpiresAt(o.ExpiresAt)
	if o.ScopeRef != nil {
		create = create.SetScopeRef(*o.ScopeRef)
	}
	row, err := create.Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			// The pending-unique lost the same-scope race — the survivor
			// is the order to return; its stored credential is what the
			// payer continues with.
			if survivor, reErr := rp.ActiveByScope(ctx, o.UserID, o.ScopeKind, o.ScopeRef); reErr == nil && survivor != nil {
				return survivor, nil
			}
		}
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapOrder(row), nil
}

func (rp *payOrderRP) ByNo(ctx context.Context, orderNo string) (*Order, error) {
	row, err := rp.Platform.GetClient(ctx).PayOrder.Query().
		Where(payorder.OrderNoEQ(orderNo)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapOrder(row), nil
}

func (rp *payOrderRP) ActiveByScope(ctx context.Context, userID, scopeKind string, scopeRef *string) (*Order, error) {
	query := rp.Platform.GetClient(ctx).PayOrder.Query().
		Where(
			payorder.UserIDEQ(userID),
			payorder.ScopeKindEQ(payorder.ScopeKind(scopeKind)),
			payorder.StateEQ(payorder.StatePending),
		).
		Order(ent.Desc(payorder.FieldCreatedAt))
	if scopeRef != nil {
		query = query.Where(payorder.ScopeRefEQ(*scopeRef))
	} else {
		// A period order's null ref: concurrent creations may leave more
		// than one live (NULLS DISTINCT never clashes) — newest wins.
		query = query.Where(payorder.ScopeRefIsNil())
	}
	row, err := query.First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapOrder(row), nil
}

func (rp *payOrderRP) ClosePending(ctx context.Context, orderNo string) (*Order, error) {
	rows, err := rp.Platform.GetClient(ctx).PayOrder.Update().
		Where(payorder.OrderNoEQ(orderNo), payorder.StateEQ(payorder.StatePending)).
		SetState(payorder.StateClosed).
		Save(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	if rows == 0 {
		return nil, nil
	}
	return rp.ByNo(ctx, orderNo)
}

func (rp *payOrderRP) List(ctx context.Context, in *ListIn) (*ListOut, error) {
	query := rp.Platform.GetClient(ctx).PayOrder.Query().
		Where(payorder.UserIDEQ(in.UserID)).
		Order(ent.Desc(payorder.FieldCreatedAt))
	total, offset, limit, err := helper.Paginate(ctx, query, in.PageRequest, helper.DefaultPageSizeMax)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	rows, err := query.All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return &ListOut{
		PageResponse: helper.BuildPageResponse(total, offset, limit),
		Orders:       utils.SliceMap(rows, mapOrder),
	}, nil
}

func (rp *entitlementRP) Exists(ctx context.Context, userID, scopeKind string, scopeRef *string) (bool, error) {
	query := rp.Platform.GetClient(ctx).Entitlement.Query().
		Where(
			entitlement.UserIDEQ(userID),
			entitlement.ScopeKindEQ(entitlement.ScopeKind(scopeKind)),
		)
	if scopeRef != nil {
		query = query.Where(entitlement.ScopeRefEQ(*scopeRef))
	} else {
		query = query.Where(entitlement.ScopeRefIsNil())
	}
	return query.Exist(ctx)
}

func (rp *payChannelRP) CreatePayment(ctx context.Context, provider string, in *payment.CreatePaymentInput) (*payment.PrepayResult, error) {
	backend, err := rp.Platform.PaymentFor(provider)
	if err != nil {
		return nil, rp.Platform.HandlePaymentError(err)
	}
	out, err := backend.Gateway.CreatePayment(ctx, in)
	if err != nil {
		return nil, rp.Platform.HandlePaymentError(err)
	}
	return out, nil
}

func (rp *payChannelRP) QueryPayment(ctx context.Context, provider, outTradeNo string) (*payment.PaymentStatusResult, error) {
	backend, err := rp.Platform.PaymentFor(provider)
	if err != nil {
		return nil, rp.Platform.HandlePaymentError(err)
	}
	out, err := backend.Gateway.QueryPayment(ctx, outTradeNo)
	if err != nil {
		return nil, rp.Platform.HandlePaymentError(err)
	}
	return out, nil
}

func (rp *payChannelRP) ClosePayment(ctx context.Context, provider, outTradeNo string) error {
	backend, err := rp.Platform.PaymentFor(provider)
	if err != nil {
		return rp.Platform.HandlePaymentError(err)
	}
	if err := backend.Gateway.ClosePayment(ctx, outTradeNo); err != nil {
		return rp.Platform.HandlePaymentError(err)
	}
	return nil
}

func (rp *payChannelRP) Channels(provider string) ([]payment.Channel, error) {
	channels, err := rp.Platform.Payments().ChannelsFor(provider)
	if err != nil {
		return nil, rp.Platform.HandlePaymentError(err)
	}
	return channels, nil
}

func (rp *payChannelRP) DefaultProvider() string { return rp.Platform.Payments().DefaultProvider() }

func (rp *payChannelRP) NotifyBaseURL() string { return rp.Platform.Payments().NotifyBaseURL() }

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

func mapOrder(row *ent.PayOrder) *Order {
	return &Order{
		OrderNo:         row.OrderNo,
		UserID:          row.UserID,
		SKUCode:         row.SkuCode,
		ScopeKind:       string(row.ScopeKind),
		ScopeRef:        row.ScopeRef,
		Provider:        row.Provider,
		Channel:         row.Channel,
		AmountFen:       row.AmountFen,
		State:           string(row.State),
		CredentialKind:  row.CredentialKind,
		CredentialValue: row.CredentialValue,
		TransactionID:   row.TransactionID,
		CreatedAt:       row.CreatedAt,
		ExpiresAt:       row.ExpiresAt,
		PaidAt:          row.PaidAt,
	}
}
