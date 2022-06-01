package bootstrap

import (
	"context"
	"gitee.com/DXTeam/idea-go.git/helper/number"
	"github.com/gin-gonic/gin"
	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcRecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpcOpentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"time"
)

func NewGrpcServer() *grpc.Server {
	// TODO 定义grpc监控指标
	// grpc_prometheus.GrpcServerMonitorMetricsRegist()
	// grpc_prometheus.GetGrpcServerMetrics().EnableHandlingTimeSummary()

	s := grpc.NewServer(
		grpc.MaxConcurrentStreams(100),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			PermitWithoutStream: true,
		}),

		grpc.KeepaliveParams(
			keepalive.ServerParameters{
				MaxConnectionIdle: time.Duration(30) * time.Second,
				Time:              time.Duration(15) * time.Second,
				Timeout:           time.Duration(3) * time.Second,
			},
		),
		grpc.UnaryInterceptor(
			grpcMiddleware.ChainUnaryServer(
				//grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
				grpcRecovery.UnaryServerInterceptor(),
				grpcOpentracing.UnaryServerInterceptor(),
				//grpc_prometheus.GetGrpcServerMetrics().UnaryServerInterceptor(),
			),
		),
	)

	return s
}

func GetGrpcConn(ctx *gin.Context, name string) (*IdleConn, context.Context) {
	newCtx := setTraceCtx(ctx)
	conn, err := grpcConnPool[name].Get()
	if err == nil {
		defer grpcConnPool[name].Put(conn)
	}

	return conn, newCtx
}

func PutGrpcConn(name string, conn *IdleConn) {
	grpcConnPool[name].RetrieveConcurrentStream(conn)
}

// setTraceCtx 设置trace_id, span_id等信息
func setTraceCtx(ctx *gin.Context) context.Context {
	var ctxMD context.Context

	requestId := ctx.GetHeader("x-request-id")

	traceId := ctx.GetHeader("X-B3-TraceId")
	spanId := ctx.GetHeader("X-B3-SpanId")
	//parentSpanId := ctx.GetHeader("x-b3-parentspanid")
	sampled := ctx.GetHeader("X-B3-Sampled")

	if sampled == "1" && traceId != "" {
		number.GenerateTraceId()
		md := metadata.Pairs("X-B3-Sampled", sampled,
			"x-request-id", requestId,
			"X-B3-TraceId", traceId,
			"X-B3-SpanId", number.GenerateSpanId(8),
			"X-B3-ParentSpanId", spanId,
			"x-b3-sampled", sampled,
			"x-b3-flags", ctx.GetHeader("x-b3-flags"),
			"x-ot-span-context", ctx.GetHeader("x-ot-span-context"),
			//"X-B3-ProjectId", ctx.GetHeader("X-B3-ProjectId"),
		)
		ctxMD = metadata.NewOutgoingContext(context.Background(), md)
	} else {
		ctxMD = context.Background()
	}

	return ctxMD
}

func panicLog(err error) {
	CoreCtx.Logger.Error(err.Error(), zap.String("type", "grpc"))
}
