package platform

import (
	"fmt"
	"log/slog"

	"cyber-ecosystem/shared-go/capability/sms"
	"cyber-ecosystem/shared-go/capability/sms/dysmsapi"
	smsmock "cyber-ecosystem/shared-go/capability/sms/mock"

	"cyber-ecosystem/app/services/system/internal/conf"
)

func NewSms(auth *conf.Auth, logger *slog.Logger) (sms.Sender, error) {
	var row *conf.Auth_Provider
	// The row's backend field is the whole channel choice — "dysmsapi"
	// demands the credential pair or the boot fails, "mock" records sends
	// and never dials regardless of credentials; empty is a missing dial,
	// not a third behavior, and fails the boot the same way as a typo
	// (the auto-pick-by-credentials branch is gone by the explicit-config
	// ruling: a boot that silently switched channels on which fields
	// happened to be set made the yaml lie about what the service does).
	// "sms" is the auth module's registered provider kind; modules import
	// platform, never the reverse, so the literal is pinned here with the
	// row's kind field as the join.
	if auth != nil {
		for _, r := range auth.GetProviders() {
			if r.GetKind() == "sms" && r.GetEnabled() {
				row = r
				break
			}
		}
	}
	if row == nil {
		// An absent or disabled row is an explicit deployment dial: this
		// sender is the mute placeholder behind an auth face that refuses
		// the provider at request time; nothing here can dial.
		logger.Warn("sms channel resolved", "backend", "none", "mode", "mute-placeholder", "reason", "no enabled provider row")
		return smsmock.New(), nil
	}
	sc := row.GetSms()
	switch backend := sc.GetBackend(); backend {
	case "mock":
		logger.Info("sms channel resolved", "provider", row.GetKey(), "backend", "mock", "mode", "record-only")
		return smsmock.New(), nil
	case "dysmsapi":
		id, secret := sc.GetAccessKeyId(), sc.GetAccessKeySecret()
		if id == "" || secret == "" {
			return nil, fmt.Errorf("sms provider %q: backend dysmsapi demands the credential pair (access_key_id=%s, access_key_secret=%s)", row.GetKey(), fieldState(id), fieldState(secret))
		}
		sender, err := dysmsapi.New(dysmsapi.Config{
			AccessKeyID:     id,
			AccessKeySecret: secret,
			Endpoint:        sc.GetEndpoint(),
		})
		if err != nil {
			return nil, fmt.Errorf("sms provider %q: %w", row.GetKey(), err)
		}
		logger.Info("sms channel resolved", "provider", row.GetKey(), "backend", "dysmsapi", "mode", "live-dial")
		return sender, nil
	default:
		return nil, fmt.Errorf("sms provider %q: backend is required and must be \"dysmsapi\" or \"mock\" (got %q)", row.GetKey(), backend)
	}
}

func fieldState(v string) string {
	if v == "" {
		return "missing"
	}
	return "set"
}
