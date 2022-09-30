package logging

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ ContextLogger = (*ZapLogger)(nil)

type ZapLogger struct {
	*zap.SugaredLogger
}

func (z *ZapLogger) With(keysAndValues ...interface{}) ContextLogger {
	return &ZapLogger{SugaredLogger: z.SugaredLogger.With(keysAndValues...)}
}

func (z *ZapLogger) FromContext(ctx context.Context, keysAndValues ...interface{}) (context.Context, ContextLogger) {
	logger := FromContextOrFallback(ctx, z)

	if len(keysAndValues) > 0 {
		logger = logger.With(keysAndValues...)
		ctx = context.WithValue(ctx, LoggerContextKey, logger)
	}

	return ctx, logger
}

func (z *ZapLogger) Close() error {
	return z.SugaredLogger.Sync()
}

func loggingLevelToZapLevel(level loggerLevel) (zapcore.Level, bool) {
	switch level {
	case DebugLevel:
		return zapcore.DebugLevel, true
	case InfoLevel:
		return zapcore.InfoLevel, true
	case WarnLevel:
		return zapcore.WarnLevel, true
	case ErrorLevel:
		return zapcore.ErrorLevel, true
	default:
		return -1, false
	}
}

func NewZapLogger(cfg LoggerConfig) (*ZapLogger, error) {
	var zapConfig *zap.Config

	switch cfg.Mode { //nolint:exhaustive
	case DevelopmentMode:
		c := zap.NewDevelopmentConfig()
		zapConfig = &c
	default:
		c := zap.NewProductionConfig()
		zapConfig = &c
	}

	zapLevel, ok := loggingLevelToZapLevel(cfg.Level)
	if !ok {
		return nil, fmt.Errorf("unsupported logger level: %v", cfg.Level)
	}

	zapConfig.Level = zap.NewAtomicLevelAt(zapLevel)

	logger, err := zapConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("logging: can't build zap logger: %w", err)
	}

	return &ZapLogger{SugaredLogger: logger.Sugar()}, nil
}
