package sample

import "github.com/google/wire"

// ProviderSet wires this domain module's use case and service. No repo — this
// is a pure ACL facade (no own persistence); the ports are satisfied by client.
var ProviderSet = wire.NewSet(
	NewSampleUC,
	NewSampleService,
)
