package bootstrap

import (
	"context"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"idea-go/helper/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	OnlineServiceHostPort = ":50051"
	HttpTimeoutHandler    = 10 * time.Second //TimeoutHandler默认的http超时时间，http响应时间超过后直接返回client端503(不同项目可根据接口最大超时时间调整)
	EnvLocal              = "local"
	EnvTest               = "test"
	EnvBVT                = "bvt"
	EnvProduct            = "product"
	EnvBenchmark          = "benchmark"
)

type coreCtx struct {
	//ConsoleLogger *zap.Logger
	Logger    *zap.Logger
	UDPLogger *zap.Logger
	//TracingLogKafkaCollect collect.Collector
	//NoCallerLogger *zap.Logger
}

var (
	CoreCtx    coreCtx
	err        error
	DevEnv     string
	TestConfig string
)

//var grpcConnPool map[string]http2.Pool

func Init() {
	initEnv()

	InitLogger()

	InitMysql()

	InitRedis()

	InitGrpc()
}

func initEnv() {
	flag.Usage = Usage
	flag.StringVar(&DevEnv, "e", EnvLocal, "Specify env")
	flag.StringVar(&TestConfig, "t", "./config", "Specify config path for testing")
	flag.Parse()
}

func InitWeb() *gin.Engine {

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(SetLogger())
	r.Use(gin.Recovery())
	r.Use(ControlCors())

	//
	//r.Use(middleware.CheckSign())
	if DevEnv == EnvLocal || DevEnv == EnvBenchmark {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		signal.Notify(c, syscall.SIGKILL)
		go func() {
			<-c
			os.Exit(0)
		}()
	} else {
		logger.RedirectLog()
	}

	return r
}

func Usage() {
	fmt.Fprintln(os.Stderr, "Usage of ", os.Args[0], " [-e=local]")
	flag.PrintDefaults()
	os.Exit(0)
}

func RunWeb(r *gin.Engine, addr string) {
	if DevEnv == EnvLocal || DevEnv == EnvBenchmark {
		fmt.Println("local http start on " + addr)
		//router.Run(addr)
	} else {
		fmt.Println("http start on " + OnlineServiceHostPort)
		addr = OnlineServiceHostPort
		//router.Run(core.ONLINE_SERVICE_HOST_PORT)
	}
	s := &http.Server{
		Addr: addr,
		Handler: http.TimeoutHandler(
			r,
			HttpTimeoutHandler,
			"server has gone away",
		),
		ReadTimeout:  HttpTimeoutHandler,
		WriteTimeout: HttpTimeoutHandler,
		IdleTimeout:  1 * time.Minute,
		//MaxHeaderBytes: 1 << 20,
	}

	go s.ListenAndServe()

	gracefulShutdown(s)
}

// gracefulShutdown 优雅退出
func gracefulShutdown(server *http.Server) {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT, os.Interrupt)
	<-ch
	//core.DebugLog("http shutdown")

	cxt, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	server.Shutdown(cxt)
	os.Exit(0)

}

// ControlCors 设置CORS
func ControlCors() gin.HandlerFunc {
	return func(context *gin.Context) {

		context.Next()
	}
}

func CheckError(err error) error {
	if err != nil {
		if DevEnv == EnvLocal {
			CoreCtx.Logger.Error(err.Error(), zap.String("type", "system"))
			//} else {
			//	SyncUDPLog(LogStruct{
			//		Err:       err,
			//		ErrType:   "error",
			//		NeedStack: true,
			//		ErrLevel:  "error",
			//		ErrInfo:   nil,
			//		Metrics:   "",
			//	})
		}

		return err
	}

	return nil
}
