package platform

// Probes over the sms platform assembly: the row's backend field is the
// whole channel choice — mock never consults credentials, dysmsapi demands
// the pair or the boot fails, and empty fails the boot exactly like a typo
// (the pick-by-credentials branch is gone by the explicit-config ruling).
// A regression here is invisible to every downstream test (they assemble
// providers directly), so the switch carries its own table.

import (
	"io"
	"log/slog"
	"testing"

	"cyber-ecosystem/shared-go/capability/sms/dysmsapi"
	smsmock "cyber-ecosystem/shared-go/capability/sms/mock"

	"cyber-ecosystem/app/services/system/internal/conf"
)

func testSmsLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func smsRowFor(backend, id, secret string) *conf.Auth_Provider {
	return &conf.Auth_Provider{
		Key: "sms", Kind: "sms", Enabled: true,
		Sms: &conf.Auth_Provider_Sms{Backend: backend, AccessKeyId: id, AccessKeySecret: secret},
	}
}

// TestNewSmsBackendSwitch: every branch of the explicit backend switch —
// the mock cells are the workflow's core (credentials present or not, a
// mock row never dials) and once failed open in exactly the
// with-credentials direction.
func TestNewSmsBackendSwitch(t *testing.T) {
	cases := []struct {
		name     string
		backend  string
		id       string
		secret   string
		wantMock bool
		wantReal bool
		wantErr  bool
	}{
		{"empty_backend_fails_boot", "", "", "", false, false, true},
		{"empty_backend_with_pair_fails_boot", "", "id", "secret", false, false, true},
		{"forced_mock_without_credentials", "mock", "", "", true, false, false},
		{"forced_mock_with_credentials", "mock", "id", "secret", true, false, false},
		{"forced_dysmsapi_without_credentials", "dysmsapi", "", "", false, false, true},
		{"forced_dysmsapi_half_pair", "dysmsapi", "id", "", false, false, true},
		{"forced_dysmsapi_full_pair", "dysmsapi", "id", "secret", false, true, false},
		{"typo_fails_boot", "dysmsap", "id", "secret", false, false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			auth := &conf.Auth{Providers: []*conf.Auth_Provider{smsRowFor(tc.backend, tc.id, tc.secret)}}
			sender, err := NewSms(auth, testSmsLogger())
			if tc.wantErr {
				if err == nil {
					t.Fatalf("want boot error, got %T", sender)
				}
				return
			}
			if err != nil {
				t.Fatalf("want a sender, got %v", err)
			}
			if _, isMock := smsmock.ControllerOf(sender); isMock != tc.wantMock {
				t.Fatalf("mock side = %v, want %v", isMock, tc.wantMock)
			}
			_, isReal := sender.(*dysmsapi.Sender)
			if isReal != tc.wantReal {
				t.Fatalf("real side = %v, want %v", isReal, tc.wantReal)
			}
		})
	}
}

// TestNewSmsRowSelection: no auth, no enabled row, and a disabled row all
// yield the mute placeholder — behind an auth face that refuses the
// provider at request time.
func TestNewSmsRowSelection(t *testing.T) {
	if sender, err := NewSms(nil, testSmsLogger()); err != nil || sender == nil {
		t.Fatalf("nil auth: sender=%v err=%v", sender, err)
	}

	for name, rows := range map[string][]*conf.Auth_Provider{
		"no_sms_row":   {{Key: "wechat", Kind: "wechat_mp", Enabled: true}},
		"disabled_row": {{Key: "sms", Kind: "sms", Enabled: false, Sms: &conf.Auth_Provider_Sms{Backend: "dysmsapi", AccessKeyId: "id", AccessKeySecret: "secret"}}},
		"empty_rows":   {},
	} {
		t.Run(name, func(t *testing.T) {
			sender, err := NewSms(&conf.Auth{Providers: rows}, testSmsLogger())
			if err != nil {
				t.Fatalf("want the mute placeholder, got %v", err)
			}
			if _, isMock := smsmock.ControllerOf(sender); !isMock {
				t.Fatalf("want the placeholder backend, got %T", sender)
			}
		})
	}

	// An enabled row behind a disabled one still wins.
	auth := &conf.Auth{Providers: []*conf.Auth_Provider{
		{Key: "sms", Kind: "sms", Enabled: false, Sms: &conf.Auth_Provider_Sms{Backend: "mock"}},
		smsRowFor("dysmsapi", "id", "secret"),
	}}
	sender, err := NewSms(auth, testSmsLogger())
	if err != nil {
		t.Fatalf("enabled row behind disabled: %v", err)
	}
	if _, isReal := sender.(*dysmsapi.Sender); !isReal {
		t.Fatalf("want the real backend from the enabled row, got %T", sender)
	}
}
