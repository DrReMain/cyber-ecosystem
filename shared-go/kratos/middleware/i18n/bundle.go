package i18n

import (
	"embed"
	"fmt"
	"io/fs"

	"path/filepath"
	"strconv"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

const (
	MetadataI18nPrefix     = "x-md-global-i18n"
	MetadataI18nTranslated = "x-md-global-i18n-translated"
	MetadataI18nLocale     = "x-md-global-i18n-locale"
	MetadataI18nMessage    = "x-md-global-i18n-message"
)

type Bundle struct {
	bundle *i18n.Bundle
}

func NewBundleFS(fsys embed.FS, translationsDir, serviceName string) (*Bundle, error) {
	b := i18n.NewBundle(language.AmericanEnglish)
	b.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)

	pattern := filepath.Join(translationsDir, serviceName+".*.yaml")

	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		matched, _ := filepath.Match(pattern, path)
		if !matched {
			return nil
		}

		data, err := fsys.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read translation file %s: %w", path, err)
		}

		if _, err = b.ParseMessageFileBytes(data, path); err != nil {
			return fmt.Errorf("parse translation file %s: %w", path, err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &Bundle{bundle: b}, nil
}

func NewBundleFromDir(translationsDir, serviceName string) (*Bundle, error) {
	b := i18n.NewBundle(language.AmericanEnglish)
	b.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)

	pattern := filepath.Join(translationsDir, serviceName+".*.yaml")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	for _, match := range matches {
		if _, err := b.LoadMessageFile(match); err != nil {
			return nil, fmt.Errorf("load translation file %s: %w", match, err)
		}
	}

	return &Bundle{bundle: b}, nil
}

func NewBundleFromFiles(files map[string][]byte) (*Bundle, error) {
	b := i18n.NewBundle(language.AmericanEnglish)
	b.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)

	for name, data := range files {
		if _, err := b.ParseMessageFileBytes(data, name); err != nil {
			return nil, fmt.Errorf("parse translation file %s: %w", name, err)
		}
	}

	return &Bundle{bundle: b}, nil
}

func (b *Bundle) Localize(key, fallback string, data map[string]string, lang string) string {
	localizer := i18n.NewLocalizer(b.bundle, lang)

	msg := &i18n.Message{ID: key}
	if fallback != "" {
		msg.Other = fallback
	}

	cfg := &i18n.LocalizeConfig{
		DefaultMessage: msg,
		TemplateData:   toTemplateData(data),
	}

	if count := pluralCount(data); count != nil {
		cfg.PluralCount = count
	}

	result, err := localizer.Localize(cfg)
	if err != nil {
		return fallback
	}
	return result
}

func (b *Bundle) Localizer(tag language.Tag) *i18n.Localizer {
	return i18n.NewLocalizer(b.bundle, tag.String())
}

func toTemplateData(data map[string]string) map[string]interface{} {
	if data == nil {
		return nil
	}
	m := make(map[string]interface{}, len(data))
	for k, v := range data {
		if strings.HasPrefix(k, MetadataI18nPrefix) {
			continue
		}
		m[k] = convertValue(v)
	}
	return m
}

func convertValue(s string) interface{} {
	switch s {
	case "true":
		return true
	case "false":
		return false
	default:
		if i, err := strconv.Atoi(s); err == nil {
			return i
		}
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return f
		}
		return s
	}
}

func pluralCount(data map[string]string) interface{} {
	if data == nil {
		return nil
	}
	if v, ok := data["Count"]; ok {
		return convertValue(v)
	}
	if v, ok := data["count"]; ok {
		return convertValue(v)
	}
	return nil
}
