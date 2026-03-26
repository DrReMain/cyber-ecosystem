package server

import (
	"context"
	stderrors "errors"
	"fmt"

	"github.com/DrReMain/cyber-ecosystem/examples/template2/internal/locales"

	"github.com/DrReMain/cyber-ecosystem/shared-go/kratos/i18n"
	"github.com/DrReMain/cyber-ecosystem/shared-go/kratos/i18n/goi18n"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"

	"buf.build/go/protovalidate"
	"golang.org/x/text/language"
)

var languages = []string{"zh-Hans", "en"}

var i18nTranslator i18n.Translator = newDefaultI18nTranslator()

func SetI18nTranslator(translator i18n.Translator) {
	if translator == nil {
		return
	}
	i18nTranslator = translator
}

func newDefaultI18nTranslator() i18n.Translator {
	provider := goi18n.NewProvider(goi18n.NewBundle(language.English))
	for _, lang := range languages {
		_ = provider.LoadMessageFileFS(locales.FS, fmt.Sprintf("active.%s.json", lang))
	}
	return i18n.NewService(i18n.NewHeaderLanguageProvider("", languages...), provider)
}

func resolveErrorMessage(ctx context.Context, se *errors.Error) string {
	if se == nil {
		return ""
	}
	message := se.Message
	if shouldIgnoreFrameworkMessage(se, message) {
		message = ""
	}
	if message == "" {
		message = i18nTranslator.Translate(ctx, se.Reason, se.Reason, nil)
	}
	return message
}

func shouldIgnoreFrameworkMessage(se *errors.Error, message string) bool {
	if se == nil || message == "" {
		return false
	}
	var valErr *protovalidate.ValidationError
	if stderrors.As(se.Unwrap(), &valErr) && message == valErr.Error() {
		return true
	}
	return false
}

func localizeErrorMiddleware(defaultReason string) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			reply, err := handler(ctx, req)
			if err == nil {
				return reply, nil
			}
			origin := errors.FromError(err)
			reason := origin.Reason
			if reason == "" {
				reason = defaultReason
			}
			se := errors.New(int(origin.Code), reason, origin.Message).WithCause(origin.Unwrap()).WithMetadata(origin.Metadata)
			message := resolveErrorMessage(ctx, se)
			if message == se.Message && se.Reason == origin.Reason {
				return reply, err
			}
			rewritten := errors.New(int(se.Code), se.Reason, message).WithCause(se.Unwrap()).WithMetadata(se.Metadata)
			return reply, rewritten
		}
	}
}
