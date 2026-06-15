package comparative

// timeout_comparative_test.go — Step 1 of the timeout-mapping consistency
// investigation.
//
// QUESTION: when a unary handler returns a BARE context error (i.e. it does NOT
// wrap ctx.Err() into a kratos error and lets the server's own unary timeout
// fire), what status CODE does each native client observe?
//
//   - gRPC client: status.Code(err)
//   - Connect client: (*connect.Error).Code()
//
// This is the gating check for whether Connect has a consistency bug relative
// to kratos gRPC. We LOG both codes (and assert only that *an* error surfaces
// and that the timeout actually fired) — the verdict is decided by reading the
// logged values, not by hard-coding an expectation here. See the investigation
// report for the conclusion + any follow-up fix in errors.go.
//
// Why a dedicated harness (not startAll): startAll sets Timeout(0) on every
// transport and shares the resource service (no Raw method). We need a server
// unary timeout on BOTH transports AND a handler that returns the bare ctx
// error, so a small purpose-built harness is cleaner than parameterizing
// startAll.

import (
	"context"
	"errors"
	"testing"
	"time"

	connectrpc "connectrpc.com/connect"
	kratosgrpc "github.com/go-kratos/kratos/v3/transport/grpc"
	"google.golang.org/genproto/googleapis/api/httpbody"
	grpcstatus "google.golang.org/grpc/status"

	connect "cyber-ecosystem/shared-go/kratos/transport/connect"

	mobilepb "cyber-ecosystem/gen/go/cyber/mobile/v1"
	mobilev1connect "cyber-ecosystem/gen/go/cyber/mobile/v1/v1connect"
)

// bareCtxErrService is a TransferService whose unary Raw blocks until the
// server-imposed context deadline fires, then returns the BARE ctx.Err()
// (context.DeadlineExceeded) — exactly the "handler didn't catch the timeout"
// scenario under test.
type bareCtxErrService struct {
	mobilepb.UnimplementedMobileTransferServiceServer
}

func (bareCtxErrService) Raw(ctx context.Context, _ *mobilepb.RawRequest) (*httpbody.HttpBody, error) {
	<-ctx.Done() // block until the server unary timeout fires
	return nil, ctx.Err()
}

// startTimeoutPair brings up a kratos grpc.Server and a connect.Server, each
// with the SAME short unary Timeout, each serving bareCtxErrService, and
// returns native clients for both. The server timeout (not a client timeout)
// is what fires — so the codes we observe reflect each transport's server-side
// error mapping of a bare ctx.Err().
func startTimeoutPair(t *testing.T, serverTimeout time.Duration) (grpcRaw mobilepb.MobileTransferServiceClient, connRaw mobilev1connect.MobileTransferServiceClient, stop func()) {
	t.Helper()
	ctx := context.Background()
	svc := bareCtxErrService{}
	var stops []func()
	cleanup := func() {
		for i := len(stops) - 1; i >= 0; i-- {
			stops[i]()
		}
	}
	fail := func(format string, args ...any) {
		cleanup()
		t.Fatalf(format, args...)
	}

	// --- gRPC server (kratos transport/grpc) with a real unary timeout ---
	grpcSrv := kratosgrpc.NewServer(
		kratosgrpc.Address("127.0.0.1:0"),
		kratosgrpc.Timeout(serverTimeout),
	)
	mobilepb.RegisterMobileTransferServiceServer(grpcSrv, svc)
	grpcEP, err := grpcSrv.Endpoint()
	if err != nil {
		fail("grpc endpoint: %v", err)
	}
	go func() { _ = grpcSrv.Start(ctx) }()
	waitReady(t, grpcEP.Host)

	grpcConn, err := kratosgrpc.NewClient(ctx,
		kratosgrpc.WithEndpoint("direct:///"+grpcEP.Host),
		kratosgrpc.WithTimeout(0), // client waits forever; the SERVER times out
	)
	if err != nil {
		fail("grpc client: %v", err)
	}
	grpcRaw = mobilepb.NewMobileTransferServiceClient(grpcConn)
	stops = append(stops, func() { _ = grpcConn.Close() })
	stops = append(stops, func() { _ = grpcSrv.Stop(ctx) })

	// --- Connect server (our shared-go transport/connect) with the SAME timeout ---
	connSrv := connect.NewServer(
		connect.Address("127.0.0.1:0"),
		connect.Timeout(serverTimeout),
	)
	mobilepb.RegisterMobileTransferServiceConnectServer(connSrv, svc)
	connEP, err := connSrv.Endpoint()
	if err != nil {
		fail("connect endpoint: %v", err)
	}
	go func() { _ = connSrv.Start(ctx) }()
	waitReady(t, connEP.Host)

	connClient, err := connect.Dial(ctx,
		connect.WithEndpoint(connEP.Host),
		connect.WithH2C(true),
		connect.WithTimeout(0), // client waits forever; the SERVER times out
	)
	if err != nil {
		fail("connect dial: %v", err)
	}
	connRaw = mobilev1connect.NewMobileTransferServiceClient(connClient.HTTPClient(), connClient.BaseURL(), connClient.ClientOptions()...)
	stops = append(stops, func() { _ = connClient.Close() })
	stops = append(stops, func() { _ = connSrv.Stop(ctx) })

	return grpcRaw, connRaw, cleanup
}

// TestTimeoutMappingGRPCvsConnect is the Step-1 comparative probe. It runs the
// SAME bare-ctx.Err() timeout scenario through a gRPC client and a Connect
// client and LOGS the observed codes. It does not assert the code values
// (those are the thing under investigation); it asserts only that an error
// surfaces on both and that the server timeout fired well before the handler
// could complete on its own.
//
// Reading the LOG output gives the verdict:
//
//	grpc code = DeadlineExceeded  (gRPC-go maps ctx errors via status.FromContextError)
//	connect code = ???             (ErrorToConnect -> kerrors.FromError path; pre-fix this is Internal)
//
// If grpc=DeadlineExceeded but connect=Internal, Connect diverges and the
// errors.go fix in Step 2 applies. If both are the same, no fix is needed.
func TestTimeoutMappingGRPCvsConnect(t *testing.T) {
	const serverTimeout = 300 * time.Millisecond

	grpcCli, connCli, stop := startTimeoutPair(t, serverTimeout)
	defer stop()

	req := &mobilepb.RawRequest{Data: []byte("x")}

	// --- gRPC leg ---
	grpcStart := time.Now()
	_, grpcErr := grpcCli.Raw(context.Background(), req)
	grpcElapsed := time.Since(grpcStart)
	grpcCode := grpcstatus.Code(grpcErr)

	// --- Connect leg ---
	connStart := time.Now()
	_, connErr := connCli.Raw(context.Background(), connectrpc.NewRequest(req))
	connElapsed := time.Since(connStart)
	connCode := connectrpc.CodeUnknown
	if connErr != nil {
		var ce *connectrpc.Error
		if errors.As(connErr, &ce) {
			connCode = ce.Code()
		}
	}

	// LOG both codes — this is the primary output of the investigation.
	t.Logf("TIMEOUT-MAPPING: grpc code = %v (err=%v, elapsed=%v)", grpcCode, grpcErr, grpcElapsed)
	t.Logf("TIMEOUT-MAPPING: connect code = %v (err=%v, elapsed=%v)", connCode, connErr, connElapsed)

	// We DO assert the invariants that are not under investigation:
	// both transports must surface an error, and it must be the SERVER timeout
	// firing (elapsed well under the handler's natural lifetime), not the
	// handler completing or the client giving up.
	if grpcErr == nil {
		t.Fatal("grpc: expected an error from Raw when the server unary timeout fires, got nil")
	}
	if connErr == nil {
		t.Fatal("connect: expected an error from Raw when the server unary timeout fires, got nil")
	}
	// Generous bound: the server timeout is 300ms; allow up to 1s for scheduler
	// / CI jitter. If we crossed ~1s the timeout did not fire.
	if grpcElapsed > 1*time.Second {
		t.Fatalf("grpc elapsed %v looks like the server timeout did NOT fire (server timeout=%v)", grpcElapsed, serverTimeout)
	}
	if connElapsed > 1*time.Second {
		t.Fatalf("connect elapsed %v looks like the server timeout did NOT fire (server timeout=%v)", connElapsed, serverTimeout)
	}

	// Sanity: grpc status must carry SOME code (not OK) for the log to be meaningful.
	if grpcCode == 0 { // codes.OK
		t.Fatalf("grpc: status.Code returned OK for a non-nil error: %v", grpcErr)
	}

	// NOTE: no assertion on grpcCode == DeadlineExceeded or connCode == ??? here.
	// The verdict is read from the LOG lines above. See the investigation notes.
}
