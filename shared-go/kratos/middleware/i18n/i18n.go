package i18n

import (
	"context"

	"golang.org/x/text/language"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

const defaultHeaderLang = "Accept-Language"

var internalKeyMap = map[string]string{
	"UNKNOWN":        "ERROR_REASON_UNSPECIFIED",
	"UNAUTHORIZED":   "ERROR_REASON_UNAUTHORIZED",
	"RATELIMIT":      "ERROR_REASON_RATELIMIT",
	"CIRCUITBREAKER": "ERROR_REASON_CIRCUITBREAKER",
}

type contextKey struct{}

type options struct {
	headerLang string
}

type Option func(*options)

func WithHeaderLang(header string) Option {
	return func(o *options) {
		o.headerLang = header
	}
}

func resolveLang(tr transport.Transporter, headerLang string) string {
	accept := tr.RequestHeader().Get(headerLang)
	if accept == "" {
		return ""
	}
	tags, _, _ := language.ParseAcceptLanguage(accept)
	if len(tags) == 0 {
		return ""
	}
	return tags[0].String()
}

func LangFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(contextKey{}).(string); ok {
		return v
	}
	return ""
}

func Server(bundle *Bundle, opts ...Option) middleware.Middleware {
	o := &options{headerLang: defaultHeaderLang}
	for _, opt := range opts {
		opt(o)
	}

	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			reply, err := handler(ctx, req)
			if err != nil {
				ke := errors.FromError(err)

				if ke.Metadata != nil && ke.Metadata[MetadataI18nTranslated] == "true" {
					return reply, err
				}

				tr, ok := transport.FromServerContext(ctx)
				if !ok {
					return reply, err
				}

				lang := resolveLang(tr, o.headerLang)
				if lang == "" {
					return reply, err
				}

				if ke.Reason == "" {
					ke = errors.New(500, internalKeyMap["UNKNOWN"], "").WithCause(ke.Unwrap())
				}

				if transKey, ok := internalKeyMap[ke.Reason]; ok {
					ke = errors.Clone(ke)
					ke.Reason = transKey
				}

				if ke.Metadata == nil {
					ke.Metadata = make(map[string]string)
				}

				if msg := ke.Metadata[MetadataI18nMessage]; msg == "" {
					ke.Message = bundle.Localize(ke.Reason, ke.Message, ke.Metadata, lang)
				}

				ke.Metadata[MetadataI18nTranslated] = "true"
				ke.Metadata[MetadataI18nLocale] = lang

				ctx = context.WithValue(ctx, contextKey{}, lang)
				return reply, ke
			}
			return reply, err
		}
	}
}
