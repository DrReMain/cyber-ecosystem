package goi18n

import (
	"encoding/json"
	"fmt"
	"io/fs"

	libi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

type Provider struct {
	bundle *libi18n.Bundle
}

func NewProvider(bundle *libi18n.Bundle) *Provider {
	return &Provider{bundle: bundle}
}

func NewBundle(defaultLang language.Tag) *libi18n.Bundle {
	bundle := libi18n.NewBundle(defaultLang)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	return bundle
}

func (p *Provider) LoadMessageFile(path string) error {
	if p.bundle == nil {
		return fmt.Errorf("go-i18n bundle is nil")
	}
	_, err := p.bundle.LoadMessageFile(path)
	return err
}

func (p *Provider) LoadMessageFileFS(fsys fs.FS, path string) error {
	if p.bundle == nil {
		return fmt.Errorf("go-i18n bundle is nil")
	}
	_, err := p.bundle.LoadMessageFileFS(fsys, path)
	return err
}

func (p *Provider) GetMessage(lang, key string, args map[string]any) (string, bool) {
	if p.bundle == nil || key == "" {
		return "", false
	}
	localizer := libi18n.NewLocalizer(p.bundle, lang)
	msg, err := localizer.Localize(&libi18n.LocalizeConfig{
		MessageID:    key,
		TemplateData: args,
	})
	if err != nil || msg == "" {
		return "", false
	}
	return msg, true
}
