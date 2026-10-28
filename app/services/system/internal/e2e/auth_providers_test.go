package e2e

// The e2e identity-provider assembly: the registry built exactly as the
// wire graph does, with the sms family's mock backend standing in for the
// channel — the mock chain the sms login e2e drives end to end. The
// controller the mock backend answers is the harness's window into what
// the channel would have delivered. Tests that never touch a provider
// still pay the same construction so the shape under test is the shipped
// one.

import (
	"log/slog"
	"testing"

	"cyber-ecosystem/shared-go/capability/cache"
	smsmock "cyber-ecosystem/shared-go/capability/sms/mock"

	"cyber-ecosystem/app/services/system/internal/module/auth"
	"cyber-ecosystem/app/services/system/internal/platform"
)

func e2eProviders(t *testing.T, logger *slog.Logger, cacheClient *cache.Cache, tokenRP auth.TokenRP) (map[string]auth.IdentityProvider, *smsmock.Controller) {
	t.Helper()
	handleCache, err := platform.NewCacheErrorHandler()
	if err != nil {
		t.Fatalf("cache error handler: %v", err)
	}
	handleSms, err := platform.NewSmsErrorHandler()
	if err != nil {
		t.Fatalf("sms error handler: %v", err)
	}
	sender, ctrl := smsmock.Open()
	smsP := auth.NewSmsProvider(logger, sender, handleSms, handleCache, cacheClient, tokenRP)
	return auth.NewIdentityProviders(auth.NewWechatProvider(), smsP), ctrl
}
