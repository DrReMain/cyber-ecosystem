package i18n

import (
	"context"

	"golang.org/x/text/language"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

const DefaultHeaderLang = "Accept-Language"

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
	o := &options{headerLang: DefaultHeaderLang}
	for _, opt := range opts {
		opt(o)
	}

	defaultLang := bundle.DefaultLang()

	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			reply, err := handler(ctx, req)
			if err != nil {
				ke := errors.FromError(err)

				// Business code provided a custom message, skip translation.
				if ke.Metadata != nil && ke.Metadata[MetadataI18nMessage] != "" {
					return reply, err
				}

				tr, ok := transport.FromServerContext(ctx)
				if !ok {
					return reply, err
				}

				lang := resolveLang(tr, o.headerLang)
				if lang == "" {
					if defaultLang != "" {
						lang = defaultLang
					} else {
						return reply, err
					}
				}

				translated := bundle.Localize(ke.Reason, ke.Message, ke.Metadata, lang)

				// Create a new error to avoid mutating shared/package-level error objects.
				// Metadata map is shared (not written), so no map copy needed.
				newErr := errors.New(int(ke.Code), ke.Reason, translated)
				if ke.Metadata != nil {
					newErr.Metadata = ke.Metadata
				}
				if cause := ke.Unwrap(); cause != nil {
					_ = newErr.WithCause(cause)
				}

				ctx = context.WithValue(ctx, contextKey{}, lang)
				return reply, newErr
			}
			return reply, err
		}
	}
}
