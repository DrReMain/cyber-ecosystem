package auth

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-kratos/kratos/v3/errors"
	"github.com/go-kratos/kratos/v3/transport"

	"cyber-ecosystem/shared-go/kratos/security"
	"cyber-ecosystem/shared-go/kratos/security/internal/teststub"
)

func ctxWithHeader(h http.Header) context.Context {
	return transport.NewServerContext(context.Background(), teststub.NewTransportWithHeader("", h))
}

type fakeAuthenticator struct {
	subject *security.Subject
	err     error
	token   string
}

func (f *fakeAuthenticator) AuthenticateSession(_ context.Context, token string) (*security.Subject, error) {
	f.token = token
	return f.subject, f.err
}

func TestSessionAuth(t *testing.T) {
	subject := &security.Subject{UserID: "u1", SessionID: "tok"}
	tests := []struct {
		name        string
		header      http.Header
		authErr     error
		nilSubject  bool
		wantErr     error
		wantToken   string
		wantSubject bool
	}{
		{
			name:        "cookie extracted among siblings",
			header:      http.Header{"Cookie": {"other=x; session_token=tok"}},
			wantToken:   "tok",
			wantSubject: true,
		},
		{
			name:        "cookie found across split cookie headers",
			header:      http.Header{"Cookie": {"other=x", "session_token=tok"}},
			wantToken:   "tok",
			wantSubject: true,
		},
		{
			name:    "no cookie header",
			header:  http.Header{},
			wantErr: ErrMissingSession,
		},
		{
			name:    "empty cookie value rejected",
			header:  http.Header{"Cookie": {"session_token="}},
			wantErr: ErrMissingSession,
		},
		{
			name:    "wrong cookie name rejected",
			header:  http.Header{"Cookie": {"access_token=tok"}},
			wantErr: ErrMissingSession,
		},
		{
			name:    "authorization header is not a session credential",
			header:  http.Header{"Authorization": {"Bearer tok"}},
			wantErr: ErrMissingSession,
		},
		{
			name:      "authenticator error passes through",
			header:    http.Header{"Cookie": {"session_token=tok"}},
			authErr:   errors.Unauthorized("GENERAL_ERROR_UNAUTHENTICATED", ""),
			wantErr:   errors.Unauthorized("GENERAL_ERROR_UNAUTHENTICATED", ""),
			wantToken: "tok",
		},
		{
			name:       "nil subject with nil error denied",
			header:     http.Header{"Cookie": {"session_token=tok"}},
			nilSubject: true,
			wantErr:    ErrMissingSession,
			wantToken:  "tok",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authn := &fakeAuthenticator{subject: subject, err: tt.authErr}
			if tt.nilSubject {
				authn = &fakeAuthenticator{}
			}
			var gotSubject *security.Subject
			handler := SessionAuth(authn, SessionCookie("session_token"))(func(ctx context.Context, _ any) (any, error) {
				var ok bool
				gotSubject, ok = security.SubjectFromCtx(ctx)
				if !ok {
					t.Fatal("handler ran without subject in context")
				}
				return nil, nil
			})
			_, err := handler(ctxWithHeader(tt.header), nil)
			switch {
			case tt.wantErr != nil && (err == nil || err.Error() != tt.wantErr.Error()):
				t.Fatalf("err = %v; want %v", err, tt.wantErr)
			case tt.wantErr == nil && err != nil:
				t.Fatalf("err = %v; want nil", err)
			}
			if authn.token != tt.wantToken {
				t.Fatalf("authenticator token = %q; want %q", authn.token, tt.wantToken)
			}
			if tt.wantSubject && gotSubject != subject {
				t.Fatalf("subject = %v, want the authenticator's subject", gotSubject)
			}
		})
	}
}

func TestSessionAuthNoTransport(t *testing.T) {
	handler := SessionAuth(&fakeAuthenticator{}, SessionCookie("session_token"))(func(ctx context.Context, _ any) (any, error) {
		t.Fatal("handler must not run without a server transport")
		return nil, nil
	})
	if _, err := handler(context.Background(), nil); err == nil || err.Error() != ErrMissingSession.Error() {
		t.Fatalf("err = %v; want %v", err, ErrMissingSession)
	}
}

func TestSessionAuthRequiresCookieName(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("SessionAuth without SessionCookie must panic at construction")
		}
	}()
	_ = SessionAuth(&fakeAuthenticator{})
}
