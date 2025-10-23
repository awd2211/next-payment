package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Log is global logger
	Log *zap.Logger
)

// NewLogger creates a new logger
func NewLogger(env string) (*zap.Logger, error) {
	var config zap.Config

	if env == "production" {
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	logger, err := config.Build(
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return nil, err
	}

	Log = logger
	return logger, nil
}

// InitLogger initializes the global logger
func InitLogger(env string) error {
	logger, err := NewLogger(env)
	if err != nil {
		return err
	}
	Log = logger
	return nil
}

// Info logs info level message
func Info(msg string, fields ...zap.Field) {
	Log.Info(msg, fields...)
}

// Debug logs debug level message
func Debug(msg string, fields ...zap.Field) {
	Log.Debug(msg, fields...)
}

// Warn logs warn level message
func Warn(msg string, fields ...zap.Field) {
	Log.Warn(msg, fields...)
}

// Error logs error level message
func Error(msg string, fields ...zap.Field) {
	Log.Error(msg, fields...)
}

// Fatal logs fatal level message and exits
func Fatal(msg string, fields ...zap.Field) {
	Log.Fatal(msg, fields...)
	os.Exit(1)
}

// Sync flushes buffer
func Sync() {
	_ = Log.Sync()
}
