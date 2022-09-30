package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcRecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpcOpentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"github.com/qulaz/artforintrovert-test/gen/api/v1"
	"github.com/qulaz/artforintrovert-test/internal/common/commonerr"
	"github.com/qulaz/artforintrovert-test/internal/config"
	grpcController "github.com/qulaz/artforintrovert-test/internal/controller/grpc"
	"github.com/qulaz/artforintrovert-test/internal/entity"
	"github.com/qulaz/artforintrovert-test/internal/usecase"
	"github.com/qulaz/artforintrovert-test/internal/usecase/repo"
	"github.com/qulaz/artforintrovert-test/pkg/cache"
	"github.com/qulaz/artforintrovert-test/pkg/interceptors/grpc_sentry"
	"github.com/qulaz/artforintrovert-test/pkg/interceptors/requestid"
	"github.com/qulaz/artforintrovert-test/pkg/logging"
	"github.com/qulaz/artforintrovert-test/pkg/mongodb"
	"github.com/qulaz/artforintrovert-test/pkg/shutdown"
	"github.com/qulaz/artforintrovert-test/pkg/tracing"
)

const (
	serviceName = "artforintrovert-test"
)

func runGrpcGateway(ctx context.Context, grpcEndpoint string, host string, port string) {
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	err := mux.HandlePath(
		"GET",
		"/metrics",
		func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
			promhttp.Handler().ServeHTTP(w, r)
		},
	)
	if err != nil {
		panic(err)
	}

	if err := api.RegisterProductServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts); err != nil {
		panic(err)
	}

	httpServer := http.Server{ //nolint: exhaustruct
		Addr:           fmt.Sprintf("%s:%s", host, port),
		Handler:        mux,
		ReadTimeout:    time.Second * 5,
		WriteTimeout:   time.Second * 10,
		MaxHeaderBytes: 1 << 20,
	}

	panic(httpServer.ListenAndServe())
}

func main() { //nolint: cyclop
	cfg, err := config.GetConfig()
	if err != nil {
		panic(err)
	}

	loggerConfig := logging.LoggerConfig{Mode: logging.ProductionMode, Level: logging.InfoLevel}
	if cfg.API.Debug {
		loggerConfig = logging.LoggerConfig{
			Mode: logging.DevelopmentMode, Level: logging.DebugLevel,
		}
	}

	logger, err := logging.NewZapLogger(loggerConfig)
	if err != nil {
		panic(err)
	}

	if cfg.Sentry != nil && cfg.Sentry.DSN != "" {
		err := sentry.Init(
			sentry.ClientOptions{ //nolint: exhaustruct
				Dsn:         cfg.Sentry.DSN,
				Environment: cfg.Sentry.Env,
				BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
					if hint.Context == nil {
						return event
					}

					if requestId := requestid.FromContext(hint.Context); requestId != "" {
						if event.Extra == nil {
							event.Extra = make(map[string]interface{}, 1)
						}
						event.Extra["requestId"] = requestId
					}

					return event
				},
				AttachStacktrace: true,
			},
		)
		if err != nil {
			logger.Fatalw("Can't initialize Sentry", "err", err)
		}
	}

	var tracer *tracing.Traning

	if cfg.Tracing != nil && cfg.Tracing.ExporterAddress != "" && cfg.Tracing.ExporterPort != "" {
		tracer, err = tracing.New(serviceName, cfg.Tracing.ExporterAddress, cfg.Tracing.ExporterPort)
		if err != nil {
			logger.Fatalw(err.Error())
		}
	} else {
		logger.Infow("Tracing configuration is not set")
	}

	mongo, err := mongodb.New(cfg.Database.DSN)
	if err != nil {
		logger.Fatalw(err.Error())
	}

	mongoDatabase := mongo.Client().Database(cfg.Database.Name)
	productCache := cache.NewMemoryEntityCache[*entity.Product]()

	productRepo := repo.NewMongoRepository(mongoDatabase, logger)

	productUseCase := usecase.NewProductUseCase(
		productRepo,
		productCache,
		logger,
		cfg.API.ProductsCacheTtl,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go productUseCase.SyncCache(ctx)

	productGrpcServer := grpcController.NewProductGrpcServer(productUseCase, logger)

	server := grpc.NewServer(
		grpc.UnaryInterceptor(
			grpcMiddleware.ChainUnaryServer(
				requestid.UnaryServerInterceptor(),
				grpcRecovery.UnaryServerInterceptor(grpcRecovery.WithRecoveryHandler(func(p interface{}) error {
					return commonerr.ErrInternalServerError
				})),
				grpc_sentry.UnaryServerInterceptor(true),
				logging.UnaryServerInterceptor(logger, logging.ExtractRequestId, true),
				grpcOpentracing.UnaryServerInterceptor(),
			),
		),
		grpc.StreamInterceptor(
			grpcMiddleware.ChainStreamServer(
				requestid.StreamServerInterceptor(),
				grpcRecovery.StreamServerInterceptor(grpcRecovery.WithRecoveryHandler(func(p interface{}) error {
					return commonerr.ErrInternalServerError
				})),
				grpc_sentry.StreamServerInterceptor(true),
				logging.StreamServerInterceptor(logger, logging.ExtractRequestId, true),
				grpcOpentracing.StreamServerInterceptor(),
			),
		),
	)
	api.RegisterProductServiceServer(server, productGrpcServer)

	if cfg.API.Debug {
		reflection.Register(server)
	}

	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%s", cfg.API.Host, cfg.API.GrpcPort))
	if err != nil {
		logger.Fatalw(err.Error())
	}

	go shutdown.Graceful(
		[]os.Signal{syscall.SIGABRT, syscall.SIGQUIT, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM},
		mongo, tracer, listen, // add io.Closer items here
	)

	logger.Infow(fmt.Sprintf("🚀 Starting GRPC server at http://%s:%s", cfg.API.Host, cfg.API.GrpcPort))

	go runGrpcGateway(ctx, listen.Addr().String(), cfg.API.Host, cfg.API.RestPort)

	logger.Infow(fmt.Sprintf("🚀 Starting REST server at http://%s:%s", cfg.API.Host, cfg.API.RestPort))

	if err := server.Serve(listen); err != nil {
		switch {
		case errors.Is(err, net.ErrClosed):
			logger.Infow("✅ Server shutdown successfully")
		default:
			logger.Errorw("🔥 Server stopped due error", "error", err.Error())
		}
	}

	server.GracefulStop()
}
