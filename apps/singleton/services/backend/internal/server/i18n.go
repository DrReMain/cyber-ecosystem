package server

import (
	"embed"

	"golang.org/x/text/language"

	i18n "cyber-ecosystem/shared-go/kratos/middleware/i18n"
)

//go:embed locales/*.yaml
var locales embed.FS

func NewI18nBundle() (*i18n.Bundle, error) {
	return i18n.NewBundleFS(locales, "locales", "v1", language.Make("zh-CN"))
}
