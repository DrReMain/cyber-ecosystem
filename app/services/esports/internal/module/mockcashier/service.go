package mockcashier

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	"cyber-ecosystem/shared-go/capability/payment/mock"
	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"

	"cyber-ecosystem/app/services/esports/internal/conf"
	"cyber-ecosystem/app/services/esports/internal/module/order"
	"cyber-ecosystem/app/services/esports/internal/platform"
)

// Struct --------------------------------------------------------------------------------------------------------------

type MockCashierService struct {
	esportspb.UnimplementedMockCashierServiceServer

	log     *slog.Logger
	payConf *conf.Payment
	pf      *platform.Platform
	orderUC *order.OrderUC
}

func NewMockCashierService(logger *slog.Logger, payConf *conf.Payment, pf *platform.Platform, orderUC *order.OrderUC) *MockCashierService {
	return &MockCashierService{
		log:     logger.With("module", "module/mockcashier_service"),
		payConf: payConf,
		pf:      pf,
		orderUC: orderUC,
	}
}

func (c *MockCashierService) RegisterGRPC(srv *grpc.Server) {
	esportspb.RegisterMockCashierServiceServer(srv, c)
}

func (c *MockCashierService) RegisterHTTP(srv *http.Server) {
	esportspb.RegisterMockCashierServiceHTTPServer(srv, c)
}

func (c *MockCashierService) RegisterConnect(srv *connecttransport.Server) {
	esportspb.RegisterMockCashierServiceConnectServer(srv, c)
}

// Handler -------------------------------------------------------------------------------------------------------------

func (c *MockCashierService) MarkPaid(ctx context.Context, in *esportspb.MarkPaidRequest) (*esportspb.MarkPaidResponse, error) {
	// The browser sheet a mock environment renders instead of the
	// channel's own cashier: this face mints the channel-signed push the
	// real channel would deliver and hands it to the same notify pipeline
	// NotifyPayment serves, so a dev walkthrough exercises the true
	// settle path, never a shortcut. The gate answers at call time, not
	// mount time — a topology whose default provider is not mock refuses
	// every order, so the simulated cashier never exists in production
	// even though the contract carries the face.
	if c.payConf.GetDefaultProvider() != "mock" {
		return nil, esportspb.ErrorEsportsPayCashierDisabled("").WithCause(
			fmt.Errorf("default provider %q never serves a cashier", c.payConf.GetDefaultProvider()))
	}
	// The UC's me-face read owns the session and ownership refusals —
	// another member's order answers NOT_FOUND, never material.
	if _, err := c.orderUC.Get(ctx, in.GetOrderNo()); err != nil {
		return nil, err
	}
	backend, err := c.pf.PaymentFor("mock")
	if err != nil {
		return nil, esportspb.ErrorEsportsPayCashierDisabled("").WithCause(err)
	}
	ctrl, cashier := mock.ControllerOf(backend)
	if !cashier {
		return nil, esportspb.ErrorEsportsPayCashierDisabled("").WithCause(fmt.Errorf("the mock backend is not the controller"))
	}
	ni, err := ctrl.MarkPaid(in.GetOrderNo())
	if err != nil {
		c.log.Warn("cashier mark-paid: channel refused", "order_no", in.GetOrderNo(), "err", err)
		return nil, esportspb.ErrorEsportsOrderStateConflict("").WithCause(err)
	}
	hdrs := make(map[string]string, len(ni.Headers))
	for k, vs := range ni.Headers {
		if len(vs) > 0 {
			hdrs[k] = vs[0]
		}
	}
	handled, err := c.orderUC.HandleNotification(ctx, "mock", hdrs, ni.Body)
	if err != nil {
		c.log.Error("cashier mark-paid: delivery failed", "order_no", in.GetOrderNo(), "err", err)
		return nil, err
	}
	return &esportspb.MarkPaidResponse{Handled: handled}, nil
}
