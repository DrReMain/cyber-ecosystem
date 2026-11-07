package orderadmin

import (
	"context"
	"log/slog"
	"math"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"

	"cyber-ecosystem/app/services/esports/internal/module/order"
)

// Struct --------------------------------------------------------------------------------------------------------------

type OrderAdminService struct {
	esportspb.UnimplementedOrderAdminServiceServer

	log     *slog.Logger
	adminUC *order.OrderAdminUC
}

func NewOrderAdminService(logger *slog.Logger, adminUC *order.OrderAdminUC) *OrderAdminService {
	return &OrderAdminService{
		log:     logger.With("module", "module/orderadmin_service"),
		adminUC: adminUC,
	}
}

func (s *OrderAdminService) RegisterGRPC(srv *grpc.Server) {
	esportspb.RegisterOrderAdminServiceServer(srv, s)
}

func (s *OrderAdminService) RegisterHTTP(srv *http.Server) {
	esportspb.RegisterOrderAdminServiceHTTPServer(srv, s)
}

func (s *OrderAdminService) RegisterConnect(srv *connecttransport.Server) {
	esportspb.RegisterOrderAdminServiceConnectServer(srv, s)
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *OrderAdminService) ClosePaymentOrder(ctx context.Context, req *esportspb.ClosePaymentOrderRequest) (*esportspb.ClosePaymentOrderResponse, error) {
	row, err := s.adminUC.Close(ctx, req.GetOrderNo())
	if err != nil {
		return nil, err
	}
	return &esportspb.ClosePaymentOrderResponse{Order: adminOrderPB(row)}, nil
}

func (s *OrderAdminService) RefundPaymentOrder(ctx context.Context, req *esportspb.RefundPaymentOrderRequest) (*esportspb.RefundPaymentOrderResponse, error) {
	row, err := s.adminUC.Refund(ctx, req.GetOrderNo(), req.GetReason())
	if err != nil {
		return nil, err
	}
	return &esportspb.RefundPaymentOrderResponse{Order: adminOrderPB(row)}, nil
}

func (s *OrderAdminService) ListPaymentOrders(ctx context.Context, req *esportspb.ListPaymentOrdersRequest) (*esportspb.ListPaymentOrdersResponse, error) {
	in := &order.AdminListIn{PageRequest: req.GetPage()}
	if req.GetState() != esportspb.OrderState_ORDER_STATE_UNSPECIFIED {
		in.State = utils.Ptr(order.StateFromPB(req.GetState()))
	}
	if req.GetScope() != esportspb.OrderScope_ORDER_SCOPE_UNSPECIFIED {
		in.Scope = utils.Ptr(order.ScopeFromPB(req.GetScope()))
	}
	if req.GetKeyword() != "" {
		in.Keyword = utils.Ptr(req.GetKeyword())
	}
	out, err := s.adminUC.List(ctx, in)
	if err != nil {
		return nil, err
	}
	return &esportspb.ListPaymentOrdersResponse{
		Page:   out.PageResponse,
		Orders: utils.SliceMap(out.Rows, adminOrderPB),
	}, nil
}

func (s *OrderAdminService) GetPaymentOrder(ctx context.Context, req *esportspb.GetPaymentOrderRequest) (*esportspb.GetPaymentOrderResponse, error) {
	row, unlock, err := s.adminUC.Get(ctx, req.GetOrderNo())
	if err != nil {
		return nil, err
	}
	return &esportspb.GetPaymentOrderResponse{Order: adminOrderPB(row), Unlock: unlockPB(unlock)}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func adminOrderPB(r *order.AdminOrderRow) *esportspb.AdminOrder {
	return &esportspb.AdminOrder{
		Order:       order.MapOrderPB(r.Order),
		UserId:      r.UserID,
		Nickname:    utils.Wrap(r.Nickname, utils.StringW),
		PhoneMasked: utils.Wrap(r.PhoneMasked, utils.StringW),
		SkuName:     utils.Wrap(r.SkuName, utils.StringW),
		Refund:      refundViewPB(r.Order),
	}
}

func refundViewPB(o *order.Order) *esportspb.OrderRefundView {
	if o == nil || o.RefundNo == nil {
		return nil
	}
	return &esportspb.OrderRefundView{
		RefundNo:        utils.Wrap(o.RefundNo, utils.StringW),
		ChannelRefundId: utils.Wrap(o.RefundID, utils.StringW),
		State:           refundStatePB(o.RefundState),
		AmountFen:       utils.Wrap(o.RefundedFen, utils.Int64W),
		Reason:          utils.Wrap(o.RefundReason, utils.StringW),
		RefundedAt:      utils.ToTimestamp(o.RefundedAt),
	}
}

func refundStatePB(state *string) esportspb.RefundState {
	// nil folds to UNSPECIFIED — the view only exists when a refund does,
	// so an absent state never reaches a consumer's switch.
	switch utils.Deref(state, "") {
	case order.RefundProcessing:
		return esportspb.RefundState_REFUND_STATE_PROCESSING
	case order.RefundSucceeded:
		return esportspb.RefundState_REFUND_STATE_SUCCEEDED
	case order.RefundAbnormal:
		return esportspb.RefundState_REFUND_STATE_ABNORMAL
	default:
		return esportspb.RefundState_REFUND_STATE_UNSPECIFIED
	}
}

func unlockPB(v *order.UnlockView) *esportspb.OrderUnlockView {
	if v == nil {
		return nil
	}
	// An unlock count beyond int32 is unreachable (days-bounded windows);
	// the guard keeps the narrowing honest anyway.
	count := int64(math.MaxInt32)
	if int64(v.UnlockCount) < math.MaxInt32 {
		count = int64(v.UnlockCount)
	}
	out := &esportspb.OrderUnlockView{
		GrantedAt:   utils.ToTimestamp(&v.GrantedAt),
		ExpiresAt:   utils.ToTimestamp(&v.ExpiresAt),
		UnlockCount: utils.Int32W(int32(count)),
	}
	if v.RecordID != "" {
		out.RecordId = utils.StringW(v.RecordID)
	}
	return out
}
