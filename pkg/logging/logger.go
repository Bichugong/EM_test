package logging

import (
	"go.uber.org/zap"
)

// Logger интерфейс для логгера с поддержкой уровней логирования
type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
	Fatalf(template string, args ...interface{})
	With(fields ...zap.Field) Logger
}

type ZapLogger struct {
	logger *zap.Logger
}

// Error implements Logger.
func (l *ZapLogger) Error(msg string, fields ...zap.Field) {
	panic("unimplemented")
}

// Errorf implements Logger.
func (l *ZapLogger) Errorf(template string, args ...interface{}) {
	panic("unimplemented")
}

// Fatal implements Logger.
func (l *ZapLogger) Fatal(msg string, fields ...zap.Field) {
	panic("unimplemented")
}

// Fatalf implements Logger.
func (l *ZapLogger) Fatalf(template string, args ...interface{}) {
	panic("unimplemented")
}

// Infof implements Logger.
func (l *ZapLogger) Infof(template string, args ...interface{}) {
	panic("unimplemented")
}

// Warn implements Logger.
func (l *ZapLogger) Warn(msg string, fields ...zap.Field) {
	panic("unimplemented")
}

// Warnf implements Logger.
func (l *ZapLogger) Warnf(template string, args ...interface{}) {
	panic("unimplemented")
}

func New(level string) Logger {
	var zapLevel zap.AtomicLevel
	switch level {
	case "debug":
		zapLevel = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		zapLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		zapLevel = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		zapLevel = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		zapLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	config := zap.Config{
		Level:            zapLevel,
		Development:      false,
		Encoding:         "console",
		EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := config.Build()
	if err != nil {
		panic(err)
	}

	return &ZapLogger{logger: logger}
}

// Реализация методов интерфейса Logger
func (l *ZapLogger) Debug(msg string, fields ...zap.Field) {
	l.logger.Debug(msg, fields...)
}

func (l *ZapLogger) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}

// ... (аналогично для остальных методов)

func (l *ZapLogger) Debugf(template string, args ...interface{}) {
	l.logger.Sugar().Debugf(template, args...)
}

func (l *ZapLogger) With(fields ...zap.Field) Logger {
	return &ZapLogger{logger: l.logger.With(fields...)}
}
