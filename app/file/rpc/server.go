package rpc

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"

	"log"
	"log/slog"
	"net"
	"os"
	"time"

	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
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

	"github.com/crazyfrankie/cloudstorage/app/file/biz/service"
	"github.com/crazyfrankie/cloudstorage/app/file/config"
	"github.com/crazyfrankie/cloudstorage/rpc_gen/file"
)

var (
	serviceName = "service/file"
	FileReg     = prometheus.NewRegistry()
	fileMetrics = grpcprom.NewServerMetrics()
)

type Server struct {
	*grpc.Server
	Addr   string
	client *clientv3.Client
}

func init() {
	FileReg.MustRegister(fileMetrics)
}

func NewServer(f *service.FileServer, client *clientv3.Client) *Server {
	// 设置日志
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{}))
	rpcLogger := logger.With("service", "gRPC/server", "module", "file")
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
	tp := initTracerProvider()
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	s := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			fileMetrics.UnaryServerInterceptor(grpcprom.WithExemplarFromContext(labelsFromContext)),
			logging.UnaryServerInterceptor(interceptorLogger(rpcLogger), logging.WithFieldsFromContext(logTraceID)),
		),
		grpc.ChainStreamInterceptor(
			fileMetrics.StreamServerInterceptor(grpcprom.WithExemplarFromContext(labelsFromContext)),
			logging.StreamServerInterceptor(interceptorLogger(rpcLogger), logging.WithFieldsFromContext(logTraceID)),
		),
		grpc.MaxRecvMsgSize(20*1024*1024),
		grpc.MaxSendMsgSize(20*1024*1024))

	file.RegisterFileServiceServer(s, f)
	fileMetrics.InitializeMetrics(s)

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

// interceptorLogger adapts slog logger to interceptor logger.
// This code is simple enough to be copied and not imported.
func interceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

func initTracerProvider() *trace.TracerProvider {
	res, err := newResource("cloud-storage/file", "v0.0.1")
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
