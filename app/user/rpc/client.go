package rpc

import (
	"context"
	"time"

	"github.com/crazyfrankie/cloudstorage/rpc_gen/file"
	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/timeout"
	"github.com/prometheus/client_golang/prometheus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/crazyfrankie/cloudstorage/rpc_gen/sm"
)

var (
	FileReg = prometheus.NewRegistry()
	SmReg   = prometheus.NewRegistry()
	// 为不同服务创建不同的 metrics
	fileMetrics = grpcprom.NewClientMetrics(
		grpcprom.WithClientHandlingTimeHistogram(
			grpcprom.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120}),
		),
		grpcprom.WithClientCounterOptions(grpcprom.WithConstLabels(prometheus.Labels{
			"client":   "user",
			"target":   "file",
			"instance": "user-instance",
		})),
	)
	smMetrics = grpcprom.NewClientMetrics(
		grpcprom.WithClientHandlingTimeHistogram(
			grpcprom.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120}),
		),
		grpcprom.WithClientCounterOptions(grpcprom.WithConstLabels(prometheus.Labels{
			"client":   "user",
			"target":   "sm",
			"instance": "user-instance",
		})),
	)
)

func InitSmClient(cli *clientv3.Client) sm.ShortMsgServiceClient {
	//logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{}))
	//rpcLogger := logger.With("service", "gRPC/client", "module", "sm")
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

	FileReg.MustRegister(fileMetrics)
	SmReg.MustRegister(smMetrics)
	// 设置 OpenTelemetry
	tp := initTracerProvider("cloud-storage/sm")
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	builder, err := resolver.NewBuilder(cli)
	conn, err := grpc.Dial("etcd:///service/sm",
		grpc.WithDefaultCallOptions(
			grpc.MaxCallSendMsgSize(20*1024*1024), // 10MB
			grpc.MaxCallRecvMsgSize(20*1024*1024),
		),
		grpc.WithResolvers(builder),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithChainUnaryInterceptor(
			smMetrics.UnaryClientInterceptor(grpcprom.WithExemplarFromContext(labelsFromContext)),
			logging.UnaryClientInterceptor(interceptorLogger(rpcLogger), logging.WithFieldsFromContext(logTraceID)),
			timeout.UnaryClientInterceptor(500*time.Millisecond),
		),
		grpc.WithChainStreamInterceptor(
			smMetrics.StreamClientInterceptor(grpcprom.WithExemplarFromContext(labelsFromContext)),
			logging.StreamClientInterceptor(interceptorLogger(rpcLogger), logging.WithFieldsFromContext(logTraceID)),
		),
	)
	if err != nil {
		rpcLogger.Error("failed to init gRPC client", zap.Error(err))
	}

	return sm.NewShortMsgServiceClient(conn)
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
	// 设置 OpenTelemetry
	tp := initTracerProvider("cloud-storage/file")
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
			fileMetrics.UnaryClientInterceptor(grpcprom.WithExemplarFromContext(labelsFromContext)),
			logging.UnaryClientInterceptor(interceptorLogger(rpcLogger), logging.WithFieldsFromContext(logTraceID)),
			timeout.UnaryClientInterceptor(500*time.Millisecond),
		),
		grpc.WithChainStreamInterceptor(
			fileMetrics.StreamClientInterceptor(grpcprom.WithExemplarFromContext(labelsFromContext)),
			logging.StreamClientInterceptor(interceptorLogger(rpcLogger), logging.WithFieldsFromContext(logTraceID)),
		),
	)
	if err != nil {
		rpcLogger.Error("failed to init gRPC client", zap.Error(err))
	}

	return file.NewFileServiceClient(conn)
}
