package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"gitee.com/DXTeam/idea-go.git/helper/logger"
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
	"google.golang.org/grpc/reflection"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"
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

func RunServer(s *grpc.Server, addr string) {
	// register reflection
	registerGrpcReflection(s)

	if runtime.GOOS == "windows" {
		l, err := net.Listen("tcp", addr)
		if err != nil {
			logger.Error("grpc", zap.NamedError("grpc", errors.New("grpc error"+err.Error())))
		}

		if err != nil {
			os.Exit(1)
		}

		go s.Serve(l)
		gracefulGrpcShutdown(s)
	} else {
		// todo 后面加入jpillora/oversee平滑重启

		//overseer.Run(overseer.Config{
		//	//Debug:            true,
		//	TerminateTimeout: 120 * time.Second,
		//	Program: func(state overseer.State) {
		//		s.Serve(state.Listener)
		//	},
		//	Address: addr,
		//})

		l, err := net.Listen("tcp", addr)
		if err != nil {
			logger.Error("grpc", zap.NamedError("grpc", errors.New("grpc error"+err.Error())))
		}

		if err != nil {
			os.Exit(1)
		}
		fmt.Println("local grpc start on " + addr)

		go s.Serve(l)
		gracefulGrpcShutdown(s)

	}
}

// registerGrpcReflection 用户本地调试
func registerGrpcReflection(s *grpc.Server) {
	if DevEnv == EnvLocal {
		reflection.Register(s)
	}
}

func gracefulGrpcShutdown(s *grpc.Server) {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT, os.Interrupt)
	getSignal := <-ch
	if getSignal == syscall.SIGTERM || getSignal == syscall.SIGQUIT || getSignal == syscall.SIGINT {

		logger.Debug("grpc", zap.NamedError("shutdown", errors.New("grpc server shutdown, handle signal "+fmt.Sprint(getSignal))))
		time.Sleep(3 * time.Second)
	}
	s.GracefulStop()
	s.Stop()
	os.Exit(0)
}
