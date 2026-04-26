package logs

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	instanceLogger *zap.Logger
	loggerConfig   LoggerConfig
	logOnce        sync.Once
	logInitErr     error
)

type LoggerConfig struct {
	Filename   string `json:"filename" yaml:"filename"`
	MaxSize    int    `json:"maxsize" yaml:"maxsize"`
	MaxAge     int    `json:"maxage" yaml:"maxage"`
	MaxBackups int    `json:"maxbackups" yaml:"maxbackups"`
	LocalTime  bool   `json:"localtime" yaml:"localtime"`
	Compress   bool   `json:"compress" yaml:"compress"`
	Level      int    `json:"level" yaml:"level"`
	size       int64
	file       *os.File
}

func Init(logConfigJson []byte, option ...zap.Option) error {
	logOnce.Do(func() {
		var err error
		err = json.Unmarshal(logConfigJson, &loggerConfig)
		if err != nil {
			logInitErr = fmt.Errorf("failed to parse logger configuration: %w", err)
			return
		}

		// 日志级别
		var encoderCfg zapcore.EncoderConfig
		level := zapcore.Level(loggerConfig.Level)
		if level < zapcore.DebugLevel || level > zapcore.FatalLevel {
			level = zapcore.InfoLevel
		}

		// 编码器配置
		encoderCfg = zapcore.EncoderConfig{
			TimeKey:    "time",
			LevelKey:   "level",
			NameKey:    "logger",
			CallerKey:  "caller",
			MessageKey: "msg",
			//StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}

		var zapCore []zapcore.Core
		encoder := zapcore.NewJSONEncoder(encoderCfg)
		// 控制台输出
		consoleWriter := zapcore.Lock(os.Stdout)
		zapCore = append(zapCore, zapcore.NewCore(
			encoder,
			consoleWriter,
			level,
		))
		// 文件输出配置
		fileConfig := zapcore.AddSync(&lumberjack.Logger{
			Filename:   loggerConfig.Filename,
			MaxSize:    loggerConfig.MaxSize,    // megabytes
			MaxBackups: loggerConfig.MaxBackups, // 日志文件保留的最大个数
			MaxAge:     loggerConfig.MaxAge,     // days
			LocalTime:  loggerConfig.LocalTime,
			Compress:   loggerConfig.Compress, // 是否压缩
		})
		fileWriter := zapcore.AddSync(fileConfig)

		zapCore = append(zapCore, zapcore.NewCore(
			encoder,
			fileWriter,
			level,
		))

		// 合并两个输出目标
		core := zapcore.NewTee(zapCore...)
		option = append(option, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
		instanceLogger = zap.New(core, option...)
	})
	return logInitErr
}

func GetLogger(m string) *zap.Logger {
	if instanceLogger == nil {
		return zap.NewNop()
	}
	return instanceLogger.With(zap.String("module", m))
}
