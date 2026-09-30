// Package connect is a clean-room Connect transport for Kratos v3. Its
// contract has two halves, and neither may quietly win over the other:
//
//   - Kratos semantics first. Middleware, transport contexts, error mapping
//     and lifecycles follow the official http/grpc transports — this package
//     is "the binding kratos would have shipped", not a policy maker. Where
//     the two official bindings disagree (streaming middleware: http chains
//     unary middleware per-RPC, grpc applies stream middleware per-message),
//     connect runs the UNION: per-RPC unary middleware (adapter.go chainUnary)
//     plus per-message stream middleware (middlewareStream). Connect is never
//     weaker than either official binding on any middleware behavior. The
//     comparative/ test matrix enforces parity.
//
//   - Connect-native strengths where the protocol is strictly better: typed
//     procedure paths instead of stringly-typed routes (no magic URL values),
//     end-to-end generated typing (proto → Go + TS clients), streaming over
//     plain HTTP, and the connect error model (code + details) mapping losslessly
//     onto kratos errors. Deviations from kratos semantics are allowed ONLY
//     where the connect protocol is strictly better, and must be documented.
//
// The one invented extension is the WebSocket upgrade dialect for
// client/bidi procedures (ws.go), forced by browser reality: fetch cannot
// stream request bodies, so the connect protocol's client/bidi legs are
// unreachable from browsers. The dialect lives on the same port and procedure
// paths, and its handler semantics mirror the connect streaming legs exactly
// (including chainUnary), keeping one procedure's behavior identical across
// legs.
package connect
