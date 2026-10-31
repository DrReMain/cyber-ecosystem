package order

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	"cyber-ecosystem/shared-go/capability/payment"
	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
)

// Struct --------------------------------------------------------------------------------------------------------------

type OrderService struct {
	esportspb.UnimplementedOrderServiceServer

	log     *slog.Logger
	orderUC *OrderUC
}

func NewOrderService(logger *slog.Logger, orderUC *OrderUC) *OrderService {
	return &OrderService{
		log:     logger.With("module", "module/order_service"),
		orderUC: orderUC,
	}
}

func (s *OrderService) RegisterGRPC(srv *grpc.Server) {
	esportspb.RegisterOrderServiceServer(srv, s)
}

func (s *OrderService) RegisterHTTP(srv *http.Server) {
	esportspb.RegisterOrderServiceHTTPServer(srv, s)
}

func (s *OrderService) RegisterConnect(srv *connecttransport.Server) {
	esportspb.RegisterOrderServiceConnectServer(srv, s)
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *OrderService) CreateOrder(ctx context.Context, in *esportspb.CreateOrderRequest) (*esportspb.CreateOrderResponse, error) {
	out, err := s.orderUC.Create(ctx, &CreateIn{
		SKUCode:     in.GetSkuCode(),
		Channel:     payment.Channel(in.GetChannel()),
		ScopeRef:    in.ScopeRef,
		PayerOpenID: in.PayerOpenid,
		ReturnURL:   in.ReturnUrl,
	})
	if err != nil {
		return nil, err
	}
	return &esportspb.CreateOrderResponse{
		Order:      mapOrderPB(out.Order),
		Credential: &esportspb.PayCredential{Kind: mapCredentialKindPB(out.Credential.Kind), Value: out.Credential.Value},
	}, nil
}

func (s *OrderService) CloseOrder(ctx context.Context, in *esportspb.CloseOrderRequest) (*esportspb.CloseOrderResponse, error) {
	order, err := s.orderUC.Close(ctx, in.GetOrderNo())
	if err != nil {
		return nil, err
	}
	return &esportspb.CloseOrderResponse{Order: mapOrderPB(order)}, nil
}

func (s *OrderService) ListOrders(ctx context.Context, in *esportspb.ListOrdersRequest) (*esportspb.ListOrdersResponse, error) {
	out, err := s.orderUC.List(ctx, in.GetPage())
	if err != nil {
		return nil, err
	}
	return &esportspb.ListOrdersResponse{
		Page:   out.PageResponse,
		Orders: utils.SliceMap(out.Orders, mapOrderPB),
	}, nil
}

func (s *OrderService) GetOrder(ctx context.Context, in *esportspb.GetOrderRequest) (*esportspb.GetOrderResponse, error) {
	order, err := s.orderUC.Get(ctx, in.GetOrderNo())
	if err != nil {
		return nil, err
	}
	return &esportspb.GetOrderResponse{Order: mapOrderPB(order)}, nil
}

func (s *OrderService) NotifyPayment(ctx context.Context, in *esportspb.NotifyPaymentRequest) (*esportspb.NotifyPaymentResponse, error) {
	handled, err := s.orderUC.HandleNotification(ctx, in.GetProvider(), in.GetHeaders(), in.GetBody())
	if err != nil {
		return nil, err
	}
	return &esportspb.NotifyPaymentResponse{Handled: handled}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func mapOrderPB(o *Order) *esportspb.Order {
	return &esportspb.Order{
		OrderNo:       utils.StringW(o.OrderNo),
		SkuCode:       utils.StringW(o.SKUCode),
		Scope:         mapScopePB(o.ScopeKind),
		ScopeRef:      utils.Wrap(o.ScopeRef, utils.StringW),
		Provider:      utils.StringW(o.Provider),
		Channel:       utils.StringW(o.Channel),
		AmountFen:     utils.Int64W(o.AmountFen),
		State:         mapStatePB(o.State),
		TransactionId: utils.Wrap(o.TransactionID, utils.StringW),
		CreatedAt:     utils.ToTimestamp(&o.CreatedAt),
		ExpiresAt:     utils.ToTimestamp(&o.ExpiresAt),
		PaidAt:        utils.ToTimestamp(o.PaidAt),
	}
}

func mapScopePB(scope string) esportspb.OrderScope {
	switch scope {
	case ScopeTrial:
		return esportspb.OrderScope_ORDER_SCOPE_TRIAL
	case ScopePeriod:
		return esportspb.OrderScope_ORDER_SCOPE_PERIOD
	default:
		return esportspb.OrderScope_ORDER_SCOPE_UNSPECIFIED
	}
}

func mapStatePB(state string) esportspb.OrderState {
	switch state {
	case StatePending:
		return esportspb.OrderState_ORDER_STATE_PENDING
	case StatePaid:
		return esportspb.OrderState_ORDER_STATE_PAID
	case StateClosed:
		return esportspb.OrderState_ORDER_STATE_CLOSED
	default:
		return esportspb.OrderState_ORDER_STATE_UNSPECIFIED
	}
}

func mapCredentialKindPB(kind payment.CredentialKind) esportspb.PayCredentialKind {
	switch kind {
	case payment.CredentialInvoke:
		return esportspb.PayCredentialKind_PAY_CREDENTIAL_KIND_INVOKE
	case payment.CredentialJumpURL:
		return esportspb.PayCredentialKind_PAY_CREDENTIAL_KIND_JUMP_URL
	case payment.CredentialQRCode:
		return esportspb.PayCredentialKind_PAY_CREDENTIAL_KIND_QR_CODE
	default:
		return esportspb.PayCredentialKind_PAY_CREDENTIAL_KIND_UNSPECIFIED
	}
}
