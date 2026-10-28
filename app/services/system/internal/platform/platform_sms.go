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
	// The explicit switch the payment platform runs, on the row's own
	// backend field: "mock" never dials regardless of credentials (the
	// dev loop after the one real-channel verification), "dysmsapi"
	// demands the credential pair or the boot fails, and empty picks by
	// credentials — a full pair dials, anything else falls to mock so dev
	// and e2e stacks run with zero channel credentials. Any other value
	// is a typo and fails the boot rather than silently dialing. "sms" is
	// the auth module's registered provider kind; modules import platform,
	// never the reverse, so the literal is pinned here with the row's kind
	// field as the join.
	if auth != nil {
		for _, r := range auth.GetProviders() {
			if r.GetKind() == "sms" && r.GetEnabled() {
				row = r
				break
			}
		}
	}
	if row == nil {
		logger.Warn("sms: no enabled provider row; the mock backend records sends instead of dialing")
		return smsmock.New(), nil
	}
	sc := row.GetSms()
	backend := sc.GetBackend()
	switch backend {
	case "mock":
		logger.Info("sms: forced mock backend; sends are recorded, nothing dials", "provider", row.GetKey())
		return smsmock.New(), nil
	case "", "dysmsapi":
	default:
		return nil, fmt.Errorf("sms provider %q: unknown backend %q", row.GetKey(), backend)
	}
	id, secret := sc.GetAccessKeyId(), sc.GetAccessKeySecret()
	if id == "" && secret == "" {
		if backend == "dysmsapi" {
			return nil, fmt.Errorf("sms provider %q: backend dysmsapi demands the credential pair", row.GetKey())
		}
		logger.Warn("sms: no credential pair; the mock backend records sends instead of dialing", "provider", row.GetKey())
		return smsmock.New(), nil
	}
	sender, err := dysmsapi.New(dysmsapi.Config{
		AccessKeyID:     id,
		AccessKeySecret: secret,
		Endpoint:        sc.GetEndpoint(),
	})
	if err != nil {
		return nil, fmt.Errorf("sms provider %q: %w", row.GetKey(), err)
	}
	return sender, nil
}
