package logger

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)
var (
	once     sync.Once
	Logger *zap.Logger
)
func Init(level string) error {
	var l zapcore.Level
	if err := l.UnmarshalText([]byte(level)); err != nil {
		return err
	}

	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	
	var err error
	Logger, err = config.Build()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	
	zap.ReplaceGlobals(Logger)

	return nil
}

func Get() *zap.Logger {
	once.Do(func() {
		config := zap.NewProductionConfig()
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		
		var err error
		Logger, err = config.Build()
		if err != nil {
			panic("failed to initialize logger: " + err.Error())
		}
	})
	return Logger
}

func Sync() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}