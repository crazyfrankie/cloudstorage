package ioc

import (
	"context"
	"fmt"

	"time"

	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/timeout"
	"github.com/prometheus/client_golang/prometheus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/crazyfrankie/cloudstorage/rpc_gen/file"
	"github.com/crazyfrankie/cloudstorage/rpc_gen/user"
)

var (
	FileReg = prometheus.NewRegistry()
	UserReg = prometheus.NewRegistry()
	// 为不同服务创建不同的 metrics
	fileMetrics = grpcprom.NewClientMetrics(
		grpcprom.WithClientHandlingTimeHistogram(
			grpcprom.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120}),
		),
		grpcprom.WithClientCounterOptions(grpcprom.WithConstLabels(prometheus.Labels{
			"client":   "gateway",
			"target":   "file",
			"instance": "gateway-instance",
		})),
	)
	userMetrics = grpcprom.NewClientMetrics(
		grpcprom.WithClientHandlingTimeHistogram(
			grpcprom.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120}),
		),
		grpcprom.WithClientCounterOptions(grpcprom.WithConstLabels(prometheus.Labels{
			"client":   "gateway",
			"target":   "user",
			"instance": "gateway-instance",
		})),
	)
)

// interceptorLogger adapts zap logger to interceptor logger.
// This code is simple enough to be copied and not imported.
func interceptorLogger(l *zap.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		f := make([]zap.Field, 0, len(fields)/2)

		for i := 0; i < len(fields); i += 2 {
			key := fields[i]
			value := fields[i+1]

			switch v := value.(type) {
			case string:
				f = append(f, zap.String(key.(string), v))
			case int:
				f = append(f, zap.Int(key.(string), v))
			case bool:
				f = append(f, zap.Bool(key.(string), v))
			default:
				f = append(f, zap.Any(key.(string), v))
			}
		}

		logger := l.WithOptions(zap.AddCallerSkip(1)).With(f...)

		switch lvl {
		case logging.LevelDebug:
			logger.Debug(msg)
		case logging.LevelInfo:
			logger.Info(msg)
		case logging.LevelWarn:
			logger.Warn(msg)
		case logging.LevelError:
			logger.Error(msg)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}

func InitFileClient(cli *clientv3.Client) file.FileServiceClient {
	//logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{}))
	//rpcLogger := logger.With("service", "gRPC/client", "module", "file")
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	rpcLogger, err := config.Build()
	if err != nil {
		panic(err)
	}
	logTraceID := func(ctx context.Context) logging.Fields {
		if span := oteltrace.SpanContextFromContext(ctx); span.IsSampled() {
			return logging.Fields{"traceID", span.TraceID().String()}
		}
		return nil
	}
	labelsFromContext := func(ctx context.Context) prometheus.Labels {
		if span := oteltrace.SpanContextFromContext(ctx); span.IsSampled() {
			return prometheus.Labels{"traceID": span.TraceID().String()}
		}
		return nil
	}

	// 设置 prometheus
	FileReg.MustRegister(fileMetrics)
	UserReg.MustRegister(userMetrics)

	// 设置 OpenTelemetry
	tp := initTracerProvider("cloud-storage/client/file")
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	builder, err := resolver.NewBuilder(cli)
	conn, err := grpc.Dial("etcd:///service/file",
		grpc.WithDefaultCallOptions(
			grpc.MaxCallSendMsgSize(20*1024*1024), // 10MB
			grpc.MaxCallRecvMsgSize(20*1024*1024),
		),
		grpc.WithResolvers(builder),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithChainUnaryInterceptor(
			timeout.UnaryClientInterceptor(500*time.Millisecond),
			fileMetrics.UnaryClientInterceptor(grpcprom.WithExemplarFromContext(labelsFromContext)),
			logging.UnaryClientInterceptor(interceptorLogger(rpcLogger), logging.WithFieldsFromContext(logTraceID))),
		grpc.WithChainStreamInterceptor(
			fileMetrics.StreamClientInterceptor(grpcprom.WithExemplarFromContext(labelsFromContext)),
			logging.StreamClientInterceptor(interceptorLogger(rpcLogger), logging.WithFieldsFromContext(logTraceID))),
	)
	if err != nil {
		rpcLogger.Error("failed to init gRPC client", zap.Error(err))
	}

	return file.NewFileServiceClient(conn)
}

func InitUserClient(cli *clientv3.Client) user.UserServiceClient {
	//logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{}))
	//rpcLogger := logger.With("service", "gRPC/client", "module", "user")
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	rpcLogger, err := config.Build()
	if err != nil {
		panic(err)
	}
	logTraceID := func(ctx context.Context) logging.Fields {
		if span := oteltrace.SpanContextFromContext(ctx); span.IsSampled() {
			return logging.Fields{"traceID", span.TraceID().String()}
		}
		return nil
	}
	labelsFromContext := func(ctx context.Context) prometheus.Labels {
		if span := oteltrace.SpanContextFromContext(ctx); span.IsSampled() {
			return prometheus.Labels{"traceID": span.TraceID().String()}
		}
		return nil
	}

	// 设置 OpenTelemetry
	tp := initTracerProvider("cloud-storage/client/user")
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	builder, err := resolver.NewBuilder(cli)
	conn, err := grpc.Dial("etcd:///service/user",
		grpc.WithDefaultCallOptions(
			grpc.MaxCallSendMsgSize(20*1024*1024), // 10MB
			grpc.MaxCallRecvMsgSize(20*1024*1024),
		),
		grpc.WithResolvers(builder),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithChainUnaryInterceptor(
			timeout.UnaryClientInterceptor(500*time.Millisecond),
			userMetrics.UnaryClientInterceptor(grpcprom.WithExemplarFromContext(labelsFromContext)),
			logging.UnaryClientInterceptor(interceptorLogger(rpcLogger), logging.WithFieldsFromContext(logTraceID))),
		grpc.WithChainStreamInterceptor(
			userMetrics.StreamClientInterceptor(grpcprom.WithExemplarFromContext(labelsFromContext)),
			logging.StreamClientInterceptor(interceptorLogger(rpcLogger), logging.WithFieldsFromContext(logTraceID))),
	)
	if err != nil {
		rpcLogger.Error("failed to init gRPC client", zap.Error(err))
	}

	return user.NewUserServiceClient(conn)
}

func initTracerProvider(servicename string) *trace.TracerProvider {
	res, err := newResource(servicename, "v0.0.1")
	if err != nil {
		fmt.Printf("failed create resource, %s", err)
	}

	tp, err := newTraceProvider(res)
	if err != nil {
		panic(err)
	}

	return tp
}

func newResource(servicename, serviceVersion string) (*resource.Resource, error) {
	return resource.Merge(resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceNameKey.String(servicename),
			semconv.ServiceVersionKey.String(serviceVersion)))
}

func newTraceProvider(res *resource.Resource) (*trace.TracerProvider, error) {
	exporter, err := zipkin.New("http://localhost:9411/api/v2/spans")
	if err != nil {
		return nil, err
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(exporter, trace.WithBatchTimeout(time.Second)), trace.WithResource(res))

	return traceProvider, nil
}
