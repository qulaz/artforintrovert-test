package commonerr

import (
	"context"

	"github.com/getsentry/sentry-go"
)

type SentryInfo struct {
	// Tags key-value теги ошибки. Отображаются практически в самом начале обзорной страницы ошибки
	Tags map[string]string

	// Contexts Контекст — озаглавленный блок key-value значений на странице с ошибкой.
	// Здесь key - название блока контекста, значение — какой-то key-value объект
	Contexts map[string]interface{}

	// Extras Экстра — инстанс контекста Contexts, который уже имеет название `Additional Data`.
	// Принимает уже key-value значение для отображения на странице
	Extras map[string]interface{}
}

func SendToSentry(ctx context.Context, err error, info *SentryInfo) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		if info != nil {
			if info.Tags != nil {
				hub.Scope().SetTags(info.Tags)
			}
			if info.Contexts != nil {
				hub.Scope().SetContexts(info.Contexts)
			}
			if info.Extras != nil {
				hub.Scope().SetExtras(info.Extras)
			}
		}

		sentryClient := hub.Client()
		if sentryClient != nil {
			hub.Client().CaptureException(
				err,
				&sentry.EventHint{Context: ctx, OriginalException: err}, //nolint: exhaustruct
				hub.Scope(),
			)
		}
	}
}
