package i18n

import "context"

type LanguageProvider interface {
	GetLanguage(ctx context.Context) string
}

type MessageProvider interface {
	GetMessage(lang, key string, args map[string]any) (string, bool)
}

type Translator interface {
	Translate(ctx context.Context, key, fallback string, args map[string]any) string
}

type Service struct {
	langProvider    LanguageProvider
	messageProvider MessageProvider
}

func NewService(langProvider LanguageProvider, messageProvider MessageProvider) *Service {
	return &Service{
		langProvider:    langProvider,
		messageProvider: messageProvider,
	}
}

func (s *Service) Translate(ctx context.Context, key, fallback string, args map[string]any) string {
	if key == "" {
		return fallback
	}
	if s.langProvider != nil && s.messageProvider != nil {
		lang := s.langProvider.GetLanguage(ctx)
		if lang != "" {
			if message, ok := s.messageProvider.GetMessage(lang, key, args); ok && message != "" {
				return message
			}
		}
	}
	if fallback != "" {
		return fallback
	}
	return key
}
