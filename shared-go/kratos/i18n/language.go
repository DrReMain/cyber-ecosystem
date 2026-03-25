package i18n

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v2/transport"

	"golang.org/x/text/language"
)

const HeaderAcceptLanguage = "Accept-Language"

type HeaderLanguageProvider struct {
	defaultLang string
	supported   []language.Tag
	matcher     language.Matcher
}

func NewHeaderLanguageProvider(defaultLang string, supported ...string) *HeaderLanguageProvider {
	p := &HeaderLanguageProvider{defaultLang: defaultLang}
	if len(supported) == 0 {
		return p
	}
	tags := make([]language.Tag, 0, len(supported))
	for _, item := range supported {
		tag, err := language.Parse(strings.TrimSpace(item))
		if err != nil {
			continue
		}
		tags = append(tags, tag)
	}
	if len(tags) == 0 {
		return p
	}
	p.supported = tags
	p.matcher = language.NewMatcher(tags)
	return p
}

func (p *HeaderLanguageProvider) GetLanguage(ctx context.Context) string {
	tr, ok := transport.FromServerContext(ctx)
	if !ok {
		return p.defaultLang
	}
	raw := strings.TrimSpace(tr.RequestHeader().Get(HeaderAcceptLanguage))
	if raw == "" {
		return p.defaultLang
	}
	tags, _, err := language.ParseAcceptLanguage(raw)
	if err != nil || len(tags) == 0 {
		return p.defaultLang
	}
	if p.matcher != nil {
		matchedTag, _, _ := p.matcher.Match(tags...)
		if matched := strings.TrimSpace(matchedTag.String()); matched != "" {
			return matched
		}
	}
	lang := strings.TrimSpace(tags[0].String())
	if lang == "" {
		return p.defaultLang
	}
	return lang
}
