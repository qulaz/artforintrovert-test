//nolint:lll
package logging_test

import (
	"context"

	"github.com/qulaz/artforintrovert-test/pkg/logging"
)

type Service struct {
	logger logging.ContextLogger
}

func (s Service) DoSmth(ctx context.Context) {
	ctx, logger := s.logger.FromContext(ctx, "service", "service name")
	s.logger.Debugw("Debug message")                      // DEBUG	logging/logger_test.go:15	Debug message
	logger.Infow("Start doing something", "key", "value") // INFO	logging/logger_test.go:16	Start doing something	{"service": "service name", "key": "value"}
	helperFunc(ctx)
}

func helperFunc(ctx context.Context) {
	logger := logging.FromContextOrDummy(ctx)
	logger.Debugw("Doing helper stuff") // DEBUG	logging/logger_test.go:22	Doing helper stuff	{"service": "service name"}
	//  																							^^^^^^^^^^^^^^^^^^^^^^^^^^ - context is still here
} //nolint:wsl

func Example() {
	logger, err := logging.NewZapLogger(logging.LoggerConfig{
		Level: logging.DebugLevel,
		Mode:  logging.DevelopmentMode,
	})
	if err != nil {
		panic(err)
	}

	s := Service{
		logger: logger,
	}

	logger.Infow("Successfully app initialization") // INFO	logging/logger_test.go:39	Successfully app initialization

	s.DoSmth(context.Background())

	productionLogger, err := logging.NewZapLogger(logging.LoggerConfig{
		Level: logging.InfoLevel,
		Mode:  logging.ProductionMode,
	})
	if err != nil {
		panic(err)
	}

	productionLogger.Infow("Production logger message", "key", "value") // {"level":"info","ts":1627133796.76541,"caller":"logging/logger_test.go:51","msg":"Production logger message","key":"value"}

	_ = productionLogger.Close()
	_ = logger.Close()
}
