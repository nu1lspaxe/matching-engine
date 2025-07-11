package logger

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	logger *zap.Logger
	once   sync.Once
)

func Init() {
	once.Do(func() {
		config := LogConfig{
			Filename:   "./logs/app.log",
			MaxSize:    10,
			MaxBackups: 5,
			MaxAge:     30,
			Level:      "DEBUG",
		}
		logger = createLogger(config)
	})
}

func GetLogger() *zap.Logger {
	if logger == nil {
		Init()
	}
	return logger
}

type LogConfig struct {
	Level      string
	Filename   string
	MaxSize    int
	MaxBackups int
	MaxAge     int
}

func createLogger(config LogConfig) *zap.Logger {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   config.Filename,
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
		Compress:   true,
	}

	writeSyncer := zapcore.AddSync(lumberJackLogger)

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	var level zapcore.Level
	switch config.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	core := zapcore.NewCore(encoder, writeSyncer, level)
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}

func Info(msg string, args ...interface{}) {
	GetLogger().Sugar().Infof(msg, args...)
}

func Debug(msg string, args ...interface{}) {
	GetLogger().Sugar().Debugf(msg, args...)
}

func Warn(msg string, args ...interface{}) {
	GetLogger().Sugar().Warnf(msg, args...)
}

func Error(msg string, args ...interface{}) {
	GetLogger().Sugar().Errorf(msg, args...)
}

func Fatal(msg string, args ...interface{}) {
	GetLogger().Sugar().Fatalf(msg, args...)
}

func InfoWith(msg string, fields ...zap.Field) {
	GetLogger().Info(msg, fields...)
}

func DebugWith(msg string, fields ...zap.Field) {
	GetLogger().Debug(msg, fields...)
}

func WarnWith(msg string, fields ...zap.Field) {
	GetLogger().Warn(msg, fields...)
}

func ErrorWith(msg string, fields ...zap.Field) {
	GetLogger().Error(msg, fields...)
}
