package commonerr

import (
	"context"
	"errors"

	pkgErrors "github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/qulaz/artforintrovert-test/pkg/logging"
)

var ErrInternalServerError = status.Error(codes.Internal, "Internal server error")

func GrpcErrHandler(ctx context.Context, err error, sentryInfo *SentryInfo) error {
	var appError AppError

	if errors.As(err, &appError) {
		switch appError.errorType {
		case ErrorTypeIncorrectInput:
			return status.Error(codes.InvalidArgument, err.Error())
		case ErrorTypeNotFound:
			return status.Error(codes.NotFound, err.Error())
		case ErrorTypeUnknown:
			return internalServerErrorHandler(ctx, err, sentryInfo)
		}
	}

	return internalServerErrorHandler(ctx, err, sentryInfo)
}

func internalServerErrorHandler(ctx context.Context, err error, sentryInfo *SentryInfo) error {
	logger := logging.FromContextOrDummy(ctx)
	logger.Errorw("⚠️ Unexpected Error", "err", err)

	if sentryInfo == nil {
		sentryInfo = &SentryInfo{}
	}
	if sentryInfo.Tags == nil {
		sentryInfo.Tags = make(map[string]string, 1)
	}
	SendToSentry(ctx, pkgErrors.WithStack(err), sentryInfo)

	return ErrInternalServerError
}
