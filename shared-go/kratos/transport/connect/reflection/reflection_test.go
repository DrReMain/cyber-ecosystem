package reflection_test

import (
	"net/http"
	"testing"

	"cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/kratos/transport/connect/reflection"
)

func TestRegisterSmoke(t *testing.T) {
	srv := connect.NewServer()
	srv.Register("/pkg.Svc/Method", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	if err := reflection.Register(srv); err != nil {
		t.Fatalf("Register: %v", err)
	}
	if len(srv.RegisteredServices()) == 0 {
		t.Fatal("no registered services")
	}
}
