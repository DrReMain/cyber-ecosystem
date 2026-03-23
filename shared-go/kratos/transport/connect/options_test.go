package connect

import (
	"crypto/tls"
	"net"
	"net/url"
	"testing"
	"time"
)

func TestServerOptions(t *testing.T) {
	srv := NewServer()
	Network("tcp4")(srv)
	Address("127.0.0.1:18080")(srv)
	Timeout(2 * time.Second)(srv)
	TLSConfig(&tls.Config{})(srv)
	DisableReflection()(srv)
	DisableH2C()(srv)
	ReflectionServices("acme.echo.v1.EchoService")(srv)
	if srv.network != "tcp4" {
		t.Fatalf("network = %q", srv.network)
	}
	if srv.address != "127.0.0.1:18080" {
		t.Fatalf("address = %q", srv.address)
	}
	if srv.timeout != 2*time.Second {
		t.Fatalf("timeout = %v", srv.timeout)
	}
	if srv.tlsConf == nil {
		t.Fatal("tls config should be set")
	}
	if !srv.disableReflection {
		t.Fatal("disableReflection should be true")
	}
	if _, ok := srv.reflectionServices["acme.echo.v1.EchoService"]; !ok {
		t.Fatal("reflection service should be recorded")
	}
	if srv.enableH2C {
		t.Fatal("enableH2C should be false after DisableH2C")
	}
}

func TestEndpointOption(t *testing.T) {
	u, err := url.Parse("connect://127.0.0.1:13000")
	if err != nil {
		t.Fatal(err)
	}
	srv := NewServer(Endpoint(u))
	got, err := srv.Endpoint()
	if err != nil {
		t.Fatal(err)
	}
	if got.String() != u.String() {
		t.Fatalf("Endpoint() = %q, want %q", got.String(), u.String())
	}
}

func TestListenerOption(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer lis.Close()

	srv := NewServer(Listener(lis))
	got, err := srv.Endpoint()
	if err != nil {
		t.Fatal(err)
	}
	if got.Host == "" {
		t.Fatal("endpoint host should not be empty")
	}
}
