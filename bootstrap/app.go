package bootstrap

import (
	"context"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"idea-go/helpers/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	LOG_DEBUG_LEVEL_ENABLE   = true //日志的debug是否开启
	LOG_DEBUG_LEVEL_DISABLE  = false
	ONLINE_SERVICE_HOST_PORT = ":50051"
	HTTP_TIMEOUT_HANDLER     = 10 * time.Second //TimeoutHandler默认的http超时时间，http响应时间超过后直接返回client端503(不同项目可根据接口最大超时时间调整)
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

	InitMysql()

	InitRedis()

	// InitGrpc()
}

func initEnv() {
	flag.Usage = Usage
	flag.StringVar(&DevEnv, "e", "local", "Specify env")
	flag.StringVar(&TestConfig, "t", "./config", "Specify config path for testing")
	flag.Parse()
}

func InitWeb() *gin.Engine {

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(ControlCors())
	//
	//r.Use(middleware.CheckSign())
	if DevEnv == "local" || DevEnv == "benchmark" {
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
	if DevEnv == "local" || DevEnv == "benchmark" {
		fmt.Println("local http start on " + addr)
		//router.Run(addr)
	} else {
		fmt.Println("http start on " + ONLINE_SERVICE_HOST_PORT)
		addr = ONLINE_SERVICE_HOST_PORT
		//router.Run(core.ONLINE_SERVICE_HOST_PORT)
	}
	s := &http.Server{
		Addr: addr,
		Handler: http.TimeoutHandler(
			r,
			HTTP_TIMEOUT_HANDLER,
			"server has gone away",
		),
		ReadTimeout:  HTTP_TIMEOUT_HANDLER,
		WriteTimeout: HTTP_TIMEOUT_HANDLER,
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
		if DevEnv == "local" {
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
