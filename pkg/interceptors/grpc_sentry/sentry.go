package grpc_sentry

import (
	"context"

	"github.com/getsentry/sentry-go"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
)

func UnaryServerInterceptor(repanic bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		hub := sentry.GetHubFromContext(ctx)
		if hub == nil {
			hub = sentry.CurrentHub().Clone()
			ctx = sentry.SetHubOnContext(ctx, hub)
		}

		hub.Scope().SetExtra("requestBody", req)
		hub.Scope().SetTransaction(info.FullMethod)
		defer recoverWithSentry(ctx, hub, repanic)

		resp, err := handler(ctx, req)

		return resp, err
	}
}

func StreamServerInterceptor(repanic bool) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := stream.Context()
		hub := sentry.GetHubFromContext(ctx)
		if hub == nil {
			hub = sentry.CurrentHub().Clone()
			ctx = sentry.SetHubOnContext(ctx, hub)
		}

		hub.Scope().SetTransaction(info.FullMethod)
		defer recoverWithSentry(ctx, hub, repanic)

		wrapped := grpc_middleware.WrapServerStream(stream)
		wrapped.WrappedContext = ctx

		err := handler(srv, wrapped)

		return err
	}
}

func recoverWithSentry(ctx context.Context, hub *sentry.Hub, repanic bool) {
	if err := recover(); err != nil {
		hub.RecoverWithContext(ctx, err)

		if repanic {
			panic(err)
		}
	}
}
