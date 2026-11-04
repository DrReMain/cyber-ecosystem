package order

import (
	"io"
	"log/slog"
	nethttp "net/http"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	"cyber-ecosystem/shared-go/capability/payment/mock"
	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	"cyber-ecosystem/app/services/esports/internal/conf"
	"cyber-ecosystem/app/services/esports/internal/platform"
)

const cashierMarkPaidPath = "/api/v1/esports/dev/cashier/mark-paid"

type MockCashier struct {
	log     *slog.Logger
	payConf *conf.Payment
	pf      *platform.Platform
	orderUC *OrderUC
}

func NewMockCashier(logger *slog.Logger, payConf *conf.Payment, pf *platform.Platform, orderUC *OrderUC) *MockCashier {
	return &MockCashier{
		// The module is the log unit — this second service-face instance
		// rides the service tag and tells itself apart by field.
		log:     logger.With("module", "module/order_service").With("registrar", "mock_cashier"),
		payConf: payConf,
		pf:      pf,
		orderUC: orderUC,
	}
}

func (c *MockCashier) RegisterGRPC(*grpc.Server) {}

func (c *MockCashier) RegisterConnect(*connecttransport.Server) {}

func (c *MockCashier) RegisterHTTP(srv *http.Server) {
	// The mock channel's simulated cashier — the browser face over the dev
	// backend's controller. MarkPaid mints the channel-signed push the real
	// channel would deliver; this route hands it to the same notify
	// pipeline NotifyPayment serves, so a dev walkthrough exercises the
	// true settle path, not a shortcut. It mounts only when the mock
	// provider is the default: a production topology never sees the route
	// at all. The raw handler carries no proto annotation by design — its
	// operation never starts with the protected namespace, so the access
	// ladder passes it through the way it passes framework built-ins; the
	// mount gate below is the entire security story, which is why it stays
	// this blunt.
	if c.payConf.GetDefaultProvider() != "mock" {
		c.log.Info("mock cashier not mounted: default provider is not mock", "provider", c.payConf.GetDefaultProvider())
		return
	}
	backend, err := c.pf.PaymentFor("mock")
	if err != nil {
		c.log.Warn("mock cashier not mounted: no mock backend", "err", err)
		return
	}
	ctrl, ok := mock.ControllerOf(backend)
	if !ok {
		c.log.Warn("mock cashier not mounted: backend is not the mock controller")
		return
	}
	srv.HandleFunc(cashierMarkPaidPath, func(w nethttp.ResponseWriter, r *nethttp.Request) {
		if r.Method != nethttp.MethodPost {
			w.WriteHeader(nethttp.StatusMethodNotAllowed)
			return
		}
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(nethttp.StatusBadRequest)
			return
		}
		in, err := utils.Unmarshal[struct {
			OrderNo string `json:"order_no"`
		}](raw)
		if err != nil || in.OrderNo == "" {
			w.WriteHeader(nethttp.StatusBadRequest)
			return
		}
		ni, err := ctrl.MarkPaid(in.OrderNo)
		if err != nil {
			c.log.Warn("cashier mark-paid: channel refused", "order_no", in.OrderNo, "err", err)
			w.WriteHeader(nethttp.StatusConflict)
			return
		}
		hdrs := make(map[string]string, len(ni.Headers))
		for k, vs := range ni.Headers {
			if len(vs) > 0 {
				hdrs[k] = vs[0]
			}
		}
		handled, err := c.orderUC.HandleNotification(r.Context(), "mock", hdrs, ni.Body)
		if err != nil {
			c.log.Error("cashier mark-paid: delivery failed", "order_no", in.OrderNo, "err", err)
			w.WriteHeader(nethttp.StatusInternalServerError)
			return
		}
		body, err := utils.Marshal(struct {
			Handled bool `json:"handled"`
		}{Handled: handled})
		if err != nil {
			w.WriteHeader(nethttp.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	})
	c.log.Info("mock cashier mounted", "path", cashierMarkPaidPath)
}
