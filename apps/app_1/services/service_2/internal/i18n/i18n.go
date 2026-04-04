package i18n

import (
	"embed"

	"github.com/google/wire"

	i18nmiddleware "cyber-ecosystem/shared-go/kratos/middleware/i18n"
)

//go:embed translations/*.yaml
var translationsFS embed.FS

type Bundle = i18nmiddleware.Bundle

func NewBundle() (*Bundle, error) {
	return i18nmiddleware.NewBundleFS(translationsFS, "translations", "v1")
}

var ProviderSet = wire.NewSet(
	NewBundle,
)
