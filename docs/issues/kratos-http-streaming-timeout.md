# Kratos v3 HTTP Streaming Timeout

> Status: Issue filed, waiting for official response.

## Problem

Kratos v3 HTTP server applies `context.WithTimeout` uniformly to **all** requests, including SSE and WebSocket streaming connections. This causes long-running streams to be killed when the configured timeout is reached.

In contrast, the gRPC transport intentionally **skips** timeout for streaming calls — only unary calls get `context.WithTimeout`. This is an inconsistency between the two transports.

## Evidence

### gRPC unary interceptor applies timeout

`transport/grpc/interceptor.go` (unary path):

```go
if s.timeout > 0 {
    ctx, cancel = context.WithTimeout(ctx, s.timeout)
    defer cancel()
}
```

### gRPC stream interceptor skips timeout

`transport/grpc/interceptor.go` (stream path):

No `WithTimeout` call — uses the merged context directly.

### HTTP server applies timeout to all requests

`transport/http/server.go`:

```go
if s.timeout > 0 {
    ctx, cancel = context.WithTimeout(req.Context(), s.timeout)
}
```

No distinction between unary and streaming requests. SSE and WebSocket connections inherit this context and are cancelled when the timeout fires.

## Reproduction

1. Define a server streaming method (SSE) that sends 5 events at 500ms intervals (~2.5s total).
2. Configure HTTP server `timeout: 1s`.
3. gRPC client receives all 5 events.
4. HTTP SSE client receives only ~2 events, then gets `context deadline exceeded`.

Verified in `admin_bff` with `TransferService.Subscribe`.

## Expected Behavior

HTTP streaming requests (SSE, WebSocket) should not be subject to the server-level timeout, matching gRPC stream interceptor behavior. Alternatively, provide a way to opt out of timeout per-route or per-method.

## Workaround

Until this is fixed, set HTTP `timeout` to `0` (disabled) and rely on per-request or per-method timeouts instead. This is not ideal as it removes timeout protection for unary endpoints.

## Related

- Kratos Issue #3775 (similar report, less structured)
- Kratos v3 added native SSE (server streaming) and WebSocket (client/bidi streaming) support to the HTTP transport, but the timeout logic was not updated to be streaming-aware.
