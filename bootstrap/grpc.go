package bootstrap

import (
	"errors"
	"fmt"
	"gitee.com/DXTeam/idea-go.git/helper/logger"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/keepalive"
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
			grpc_middleware.ChainUnaryServer(
				//grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
				grpc_recovery.UnaryServerInterceptor(),
				grpc_opentracing.UnaryServerInterceptor(),
				//grpc_prometheus.GetGrpcServerMetrics().UnaryServerInterceptor(),
			),
		),
	)

	return s
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

		time.Sleep(2 * time.Minute)
	}
	s.GracefulStop()
	s.Stop()
	os.Exit(0)
}
