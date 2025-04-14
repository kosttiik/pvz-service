package logger

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Log  *zap.Logger
	once sync.Once
)

func Init() error {
	var err error
	once.Do(func() {
		config := zap.NewProductionConfig()

		levelStr := os.Getenv("LOG_LEVEL")
		if levelStr != "" {
			level, parseErr := zapcore.ParseLevel(levelStr)
			if parseErr == nil {
				config.Level = zap.NewAtomicLevelAt(level)
			}
		}

		config.OutputPaths = []string{"stdout"}
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

		Log, err = config.Build()
		if err == nil {
			zap.ReplaceGlobals(Log)
		}
	})
	return err
}

func Close() {
	if Log != nil {
		Log.Sync()
	}
}
