package response

import (
	"testing"

	connectrpc "connectrpc.com/connect"
	"github.com/DrReMain/cyber-ecosystem/gen/go/common"
	kerrors "github.com/go-kratos/kratos/v2/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestBuildSuccess(t *testing.T) {
	resp, err := Build(&emptypb.Empty{}, nil)
	if err != nil {
		t.Fatalf("Build() err = %v", err)
	}
	if !resp.Success {
		t.Fatalf("resp.Success = false, want true")
	}
	if resp.Result == nil {
		t.Fatalf("resp.Result = nil, want non-nil")
	}
	if resp.Err != nil {
		t.Fatalf("resp.Err = %#v, want nil", resp.Err)
	}
}

func TestBuildFromKratosError(t *testing.T) {
	resp, err := Build(nil, kerrors.BadRequest("ERROR_REASON_VALIDATOR", "invalid input"))
	if err != nil {
		t.Fatalf("Build() err = %v", err)
	}
	if resp.Success {
		t.Fatalf("resp.Success = true, want false")
	}
	if resp.Err == nil {
		t.Fatalf("resp.Err = nil, want non-nil")
	}
	if resp.Err.Reason != "ERROR_REASON_VALIDATOR" {
		t.Fatalf("resp.Err.Reason = %q", resp.Err.Reason)
	}
}

func TestExtractErrorBodyFromConnectDetail(t *testing.T) {
	ce := connectrpc.NewError(connectrpc.CodeInvalidArgument, kerrors.BadRequest("ERROR_REASON_VALIDATOR", "invalid input"))
	detail, err := connectrpc.NewErrorDetail(&common.ErrorBody{
		Reason:  "ERROR_REASON_VALIDATOR",
		Message: "请求参数校验失败。",
	})
	if err != nil {
		t.Fatalf("NewErrorDetail() err = %v", err)
	}
	ce.AddDetail(detail)
	body := ExtractErrorBody(ce)
	if body == nil {
		t.Fatalf("ExtractErrorBody() = nil")
	}
	if body.Reason != "ERROR_REASON_VALIDATOR" {
		t.Fatalf("body.Reason = %q", body.Reason)
	}
}

func TestExtractErrorBodyFromGRPCDetail(t *testing.T) {
	st := status.New(codes.NotFound, "not found")
	withDetails, err := st.WithDetails(&common.ErrorBody{
		Reason:  "ERROR_REASON_ENT_NOT_FOUND",
		Message: "资源不存在。",
	})
	if err != nil {
		t.Fatalf("WithDetails() err = %v", err)
	}
	body := ExtractErrorBody(withDetails.Err())
	if body == nil {
		t.Fatalf("ExtractErrorBody() = nil")
	}
	if body.Reason != "ERROR_REASON_ENT_NOT_FOUND" {
		t.Fatalf("body.Reason = %q", body.Reason)
	}
}
