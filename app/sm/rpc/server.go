package rpc

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"log"
	"net"
	"time"

	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/prometheus/client_golang/prometheus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"

	"github.com/crazyfrankie/cloudstorage/app/sm/biz/service"
	"github.com/crazyfrankie/cloudstorage/app/sm/config"
	"github.com/crazyfrankie/cloudstorage/rpc_gen/sm"
)

var (
	serviceName = "service/sm"
	SmReg       = prometheus.NewRegistry()
	smMetrics   = grpcprom.NewServerMetrics()
)

type Server struct {
	*grpc.Server
	Addr   string
	client *clientv3.Client
}

func NewServer(sms *service.SmServer, client *clientv3.Client) *Server {
	// 设置日志
	//logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{}))
	//rpcLogger := logger.With("service", "gRPC/server", "module", "sm")
	rpcLogger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	logTraceID := func(ctx context.Context) logging.Fields {
		if span := oteltrace.SpanContextFromContext(ctx); span.IsSampled() {
			return logging.Fields{"traceID", span.TraceID().String()}
		}
		return nil
	}

	SmReg.MustRegister(smMetrics)
	// 设置 Prometheus metrics
	labelsFromContext := func(ctx context.Context) prometheus.Labels {
		if span := oteltrace.SpanContextFromContext(ctx); span.IsSampled() {
			return prometheus.Labels{"traceID": span.TraceID().String()}
		}
		return nil
	}

	// 设置 OpenTelemetry
	tp := initTracerProvider()
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	s := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			smMetrics.UnaryServerInterceptor(grpcprom.WithExemplarFromContext(labelsFromContext)),
			logging.UnaryServerInterceptor(interceptorLogger(rpcLogger), logging.WithFieldsFromContext(logTraceID)),
		),
		grpc.ChainStreamInterceptor(
			smMetrics.StreamServerInterceptor(grpcprom.WithExemplarFromContext(labelsFromContext)),
			logging.StreamServerInterceptor(interceptorLogger(rpcLogger), logging.WithFieldsFromContext(logTraceID)),
		))
	sm.RegisterShortMsgServiceServer(s, sms)
	smMetrics.InitializeMetrics(s)

	return &Server{
		Server: s,
		Addr:   config.GetConf().Server.Addr,
		client: client,
	}
}

func (s *Server) Serve() error {
	conn, err := net.Listen("tcp", s.Addr)
	if err != nil {
		panic(err)
	}

	err = registerService(s.client, s.Addr)
	if err != nil {
		panic(err)
	}

	return s.Server.Serve(conn)
}

func registerService(cli *clientv3.Client, port string) error {
	em, err := endpoints.NewManager(cli, serviceName)
	if err != nil {
		return err
	}

	addr := "127.0.0.1" + port
	serviceKey := serviceName + "/" + addr
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	leaseResp, err := cli.Grant(ctx, 180)
	if err != nil {
		return err
	}

	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err = em.AddEndpoint(ctx, serviceKey, endpoints.Endpoint{Addr: addr}, clientv3.WithLease(leaseResp.ID))

	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		ch, err := cli.KeepAlive(ctx, leaseResp.ID)
		if err != nil {
			log.Printf("keep alive failed lease id:%d", leaseResp.ID)
			return
		}
		for {
			select {
			case _, ok := <-ch:
				if !ok {
					log.Println("KeepAlive channel closed")
					return
				}
				fmt.Println("Lease renewed")
			case <-ctx.Done():
				return
			}
		}
	}()

	return err
}

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

func initTracerProvider() *trace.TracerProvider {
	res, err := newResource("cloud-storage/sm", "v0.0.1")
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
