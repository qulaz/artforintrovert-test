package logging

import (
	"context"
	"io"
)

type loggerContextKeyType string

const LoggerContextKey loggerContextKeyType = "logger"

type ContextLogger interface {
	StructureLogger

	// FromContext возвращает обогащенный тегами логгер из контекста или самого себя,
	// если передан пустой контекст/в контексте нет логгера.
	// Также обогащает этот логгер переданными тегами в формате ключ-значение
	FromContext(ctx context.Context, keysAndValues ...interface{}) (context.Context, ContextLogger)

	io.Closer
}

func FromContext(ctx context.Context) (ContextLogger, bool) {
	if ctx == nil {
		return nil, false
	}

	if ctxLogger, ok := ctx.Value(LoggerContextKey).(ContextLogger); ok {
		return ctxLogger, true
	}

	return nil, false
}

func FromContextOrFallback(ctx context.Context, fallback ContextLogger) ContextLogger {
	if ctxLogger, ok := FromContext(ctx); ok {
		return ctxLogger
	}

	return fallback
}

// FromContextOrDummy возвращает логгер из контекста. Если передан нулевой контекст
// или логгера в нем нет — возвращается логгер-пустышка.
func FromContextOrDummy(ctx context.Context) ContextLogger {
	return FromContextOrFallback(ctx, NewDummyLogger())
}
