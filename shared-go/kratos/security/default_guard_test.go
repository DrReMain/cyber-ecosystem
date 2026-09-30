package security_test

import (
	"context"
	"strings"
	"testing"

	"github.com/go-kratos/kratos/v3/transport"

	"cyber-ecosystem/shared-go/kratos/security"
	"cyber-ecosystem/shared-go/kratos/security/internal/teststub"
)

func TestDefaultGuard(t *testing.T) {
	guard := security.DefaultGuard()
	var handlerRan bool
	handler := func(ctx context.Context, req any) (any, error) {
		handlerRan = true
		return "ok", nil
	}

	tests := []struct {
		name        string
		ctx         context.Context
		wantErr     bool
		wantHandler bool
	}{
		{
			name:    "business RPC under cyber namespace is rejected",
			ctx:     transport.NewServerContext(context.Background(), teststub.NewTransport("/cyber.system.v1.UserService/GetUser")),
			wantErr: true,
		},
		{
			name:        "framework built-in outside cyber namespace passes",
			ctx:         transport.NewServerContext(context.Background(), teststub.NewTransport("/grpc.health.v1.Health/Check")),
			wantErr:     false,
			wantHandler: true,
		},
		{
			name:    "no transport context is rejected",
			ctx:     context.Background(),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlerRan = false
			_, err := guard(handler)(tt.ctx, nil)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tt.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), "MISSING_ANNOTATION") {
				t.Fatalf("reason = %q, want MISSING_ANNOTATION", err.Error())
			}
			if handlerRan != tt.wantHandler {
				t.Fatalf("handler ran = %v, want %v", handlerRan, tt.wantHandler)
			}
		})
	}
}
