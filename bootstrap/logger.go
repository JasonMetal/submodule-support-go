package bootstrap

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"idea-go/helper/config"
	"time"
)

var zLog *zap.Logger

func InitLogger() {
	writeSyncer := getLogWriter()
	encoder := getEncoderCfg()
	zCore := zapcore.NewCore(encoder, writeSyncer, zap.InfoLevel)

	zLog = zap.New(zCore, zap.AddCaller())
	zap.ReplaceGlobals(zLog) // 替换zap包中全局的logger实例，后续在其他包中只需使用zap.L()调用即可

	return
}

func getEncoderCfg() zapcore.Encoder {
	encoderCfg := zap.NewDevelopmentEncoderConfig()

	if DevEnv == EnvProduct {
		encoderCfg = zap.NewProductionEncoderConfig()
	}
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.TimeKey = "time"
	encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderCfg.EncodeDuration = zapcore.SecondsDurationEncoder
	encoderCfg.EncodeCaller = zapcore.ShortCallerEncoder

	return zapcore.NewJSONEncoder(encoderCfg)

}

// getLogWriter 设置zap writer
func getLogWriter() zapcore.WriteSyncer {
	path := fmt.Sprintf("./config/%s/logger.yml", DevEnv)

	cfg, err := config.GetConfig(path)
	if err != nil {
		fmt.Printf("init logger writer err")
		return nil
	}
	filename, _ := cfg.String("logger.filename")
	maxSize, _ := cfg.Int("logger.maxSize")
	maxBackup, _ := cfg.Int("logger.maxBackup")
	maxAge, _ := cfg.Int("logger.maxAge")
	compress, _ := cfg.Bool("logger.compress")

	lumberJackLogger := &lumberjack.Logger{
		Filename:   filename,  // 日志文件路径
		MaxSize:    maxSize,   // megabytes
		MaxBackups: maxBackup, // 最多保留3个备份
		MaxAge:     maxAge,    //days
		Compress:   compress,  // 是否压缩 disabled by default
	}

	return zapcore.AddSync(lumberJackLogger)
}

// Logger 接收gin框架默认的日志
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()

		cost := time.Since(start)
		zLog.Info(path,
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
			zap.Duration("cost", cost),
		)
	}
}
