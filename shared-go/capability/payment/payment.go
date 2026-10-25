package payment

// Payment is the payment capability root. Backend constructors (mock.New /
// wechat.New) populate every field — a nil field nil-derefs deep in a request
// path, not at startup; the underlying channel credentials and cert state
// belong to the backend provider's construction, not to this container.
type Payment struct {
	Gateway  Gateway
	Notifier Notifier
	Reports  Reports
}
