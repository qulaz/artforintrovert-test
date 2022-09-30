package logging

import "context"

type DummyLogger struct{}

func (d DummyLogger) Debugw(_ string, _ ...interface{}) {}

func (d DummyLogger) Infow(_ string, _ ...interface{}) {}

func (d DummyLogger) Warnw(_ string, _ ...interface{}) {}

func (d DummyLogger) Errorw(_ string, _ ...interface{}) {}

func (d DummyLogger) Fatalw(_ string, _ ...interface{}) {}

func (d DummyLogger) With(_ ...interface{}) ContextLogger {
	return d
}

func (d DummyLogger) FromContext(ctx context.Context, _ ...interface{}) (context.Context, ContextLogger) {
	return ctx, d
}

func (d DummyLogger) Close() error {
	return nil
}

func NewDummyLogger() ContextLogger {
	return DummyLogger{}
}
