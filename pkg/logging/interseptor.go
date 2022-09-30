package logging

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"github.com/qulaz/artforintrovert-test/pkg/interceptors/requestid"
)

type ExtraLogDataFn = func(ctx context.Context) []any

func UnaryServerInterceptor(logger ContextLogger, extraFn ExtraLogDataFn, repanic bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		startTime := time.Now()
		defer logRecover(
			ctx,
			startTime,
			info.FullMethod,
			logger,
			extraFn,
			repanic,
		)

		resp, err := handler(ctx, req)

		logRequest(
			ctx,
			startTime,
			err,
			info.FullMethod,
			logger,
			extraFn,
		)

		return resp, err
	}
}

func StreamServerInterceptor(logger ContextLogger, extraFn ExtraLogDataFn, repanic bool) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		startTime := time.Now()
		defer logRecover(
			stream.Context(),
			startTime,
			info.FullMethod,
			logger,
			extraFn,
			repanic,
		)

		err := handler(srv, stream)

		logRequest(
			stream.Context(),
			startTime,
			err,
			info.FullMethod,
			logger,
			extraFn,
		)

		return err
	}
}

func ExtractRequestId(ctx context.Context) []any {
	requestId := requestid.FromContext(ctx)
	if requestId == "" {
		return []any{}
	}

	return []any{"grpc.requestId", requestId}
}

func logRecover(
	ctx context.Context,
	startTime time.Time,
	method string,
	logger ContextLogger,
	extraFn ExtraLogDataFn,
	repanic bool,
) {
	if err := recover(); err != nil {
		extraData := extraFn(ctx)
		logData := getLogKeyValues(
			extraData,
			"grpc.duration", time.Since(startTime),
			"grpc.method", method,
			"panic", err,
		)

		logger.Errorw("⚠️ Panic", logData...)

		if repanic {
			panic(err)
		}
	}
}

func logRequest(
	ctx context.Context,
	startTime time.Time,
	handlerErr error,
	method string,
	logger ContextLogger,
	extraFn ExtraLogDataFn,
) {
	extraData := extraFn(ctx)
	logData := getLogKeyValues(
		extraData,
		"grpc.code", status.Code(handlerErr).String(),
		"grpc.duration", time.Since(startTime),
		"grpc.method", method,
	)

	logger.Infow("✅ Served", logData...)
}

func getLogKeyValues(extraData []interface{}, keyValues ...interface{}) []interface{} {
	logKeyValues := make([]interface{}, 0, len(extraData)+len(keyValues))

	logKeyValues = append(logKeyValues, keyValues...)
	logKeyValues = append(logKeyValues, extraData...)

	return logKeyValues
}
