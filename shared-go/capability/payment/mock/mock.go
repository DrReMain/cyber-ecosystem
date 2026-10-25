package mock

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"cyber-ecosystem/shared-go/capability/payment"
)

// Channel values mirror wechat's declared set string for string (sibling
// backends never import each other), so a consumer's channel choice never
// changes when the platform swaps mock for wechat.
const (
	ChannelJSAPI  = payment.Channel("wechat_jsapi")
	ChannelH5     = payment.Channel("wechat_h5")
	ChannelNative = payment.Channel("wechat_native")
)

// Channels declares the backend's supported values.
func Channels() []payment.Channel {
	return []payment.Channel{ChannelJSAPI, ChannelH5, ChannelNative}
}

type order struct {
	in         payment.CreatePaymentInput
	credential payment.Credential
	state      payment.PaymentState
	paidAt     time.Time
	txID       string
	refunded   payment.AmountFen
	refunds    map[string]*refundEntry
	// notification caches the minted payment notification so repeated
	// delivery redelivers the same bytes — the duplicate-notification
	// behavior channels really have.
	notification *payment.NotificationInput
}

type refundEntry struct {
	in       payment.RefundInput
	state    payment.RefundState
	refundID string
	doneAt   time.Time
}

// backend is the in-memory state machine: one mutex, one order map, one
// mock keypair for the mini wire protocol. New/Open populate a payment
// container from a single instance so Gateway, Notifier and Reports share
// the same state.
type backend struct {
	mu      sync.Mutex
	orders  map[string]*order
	aead    []byte // 32 bytes, mock notification encryption
	hmacKey []byte
	// rejects forces a channel-style business rejection per trade number —
	// the conformance probe for the ErrChannelRejected path.
	rejects map[string]*payment.ChannelError
}

func newBackend() *backend {
	aead := make([]byte, 32)
	hmacKey := make([]byte, 32)
	if _, err := rand.Read(aead); err != nil {
		panic("mock: entropy failure: " + err.Error())
	}
	if _, err := rand.Read(hmacKey); err != nil {
		panic("mock: entropy failure: " + err.Error())
	}
	return &backend{
		orders:  make(map[string]*order),
		aead:    aead,
		hmacKey: hmacKey,
		rejects: make(map[string]*payment.ChannelError),
	}
}

// New returns a payment container over the in-memory state machine.
func New() *payment.Payment {
	b := newBackend()
	return &payment.Payment{Gateway: b, Notifier: b, Reports: b}
}

// Open returns the container plus its Controller (settle an order, mint
// and deliver notifications, force rejections). Production consumers use
// New and see only the contract; Open is the test/dev face.
func Open() (*payment.Payment, *Controller) {
	b := newBackend()
	return &payment.Payment{Gateway: b, Notifier: b, Reports: b}, &Controller{b: b}
}

// ControllerOf recovers the Controller from a container this package built —
// the dev/e2e face over a backend a platform assembled from config, where
// production code reaches only the contract.
func ControllerOf(p *payment.Payment) (*Controller, bool) {
	b, ok := p.Gateway.(*backend)
	if !ok {
		return nil, false
	}
	return &Controller{b: b}, true
}

func txIDFor(outTradeNo string) string {
	var nonce [4]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		panic("mock: entropy failure: " + err.Error())
	}
	return "mock-tx-" + hex.EncodeToString(nonce[:]) + "-" + outTradeNo
}

func refundIDFor(outRefundNo string) string {
	var nonce [4]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		panic("mock: entropy failure: " + err.Error())
	}
	return "mock-rf-" + hex.EncodeToString(nonce[:]) + "-" + outRefundNo
}
