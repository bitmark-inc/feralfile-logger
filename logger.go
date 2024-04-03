package log

import (
	"strings"

	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var defaultLogger *zap.Logger

func Initialize(level string, isDebug bool, sentryDsn string) error {
	log, err := New(level, isDebug)
	if err != nil {
		return err
	}

	defaultLogger = log

	// Init sentry
	if sentryDsn != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn: sentryDsn,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func New(level string, isDebug bool) (*zap.Logger, error) {
	var config zap.Config

	if isDebug {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	// override log level by configuration
	l := zap.ErrorLevel
	switch strings.ToUpper(level) {
	case "TRACE", "DEBUG":
		l = zap.DebugLevel
	case "INFO":
		l = zap.InfoLevel
	case "WARN":
		l = zap.WarnLevel
	}

	config.Level = zap.NewAtomicLevelAt(l)

	return config.Build()
}

func MustDefaultLogger() *zap.Logger {
	if defaultLogger == nil {
		panic("use indexer logger without initializing")
	}

	return defaultLogger
}

func Debug(msg string, fields ...zap.Field) {
	MustDefaultLogger().WithOptions(zap.AddCallerSkip(1)).Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	MustDefaultLogger().WithOptions(zap.AddCallerSkip(1)).Info(msg, fields...)

	// Add a breadcrumb
	sentry.AddBreadcrumb(&sentry.Breadcrumb{
		Message: msg,
		Level:   sentry.LevelInfo,
		Data:    zapFieldsToMap(fields),
	})
}

func Warn(msg string, fields ...zap.Field) {
	MustDefaultLogger().WithOptions(zap.AddCallerSkip(1)).Warn(msg, fields...)

	// Add a breadcrumb
	sentry.AddBreadcrumb(&sentry.Breadcrumb{
		Message: msg,
		Level:   sentry.LevelWarning,
		Data:    zapFieldsToMap(fields),
	})
}

func Error(msg string, fields ...zap.Field) {
	MustDefaultLogger().WithOptions(zap.AddCallerSkip(1)).Error(msg, fields...)

	// Add a breadcrumb
	sentry.AddBreadcrumb(&sentry.Breadcrumb{
		Message: msg,
		Level:   sentry.LevelError,
		Data:    zapFieldsToMap(fields),
	})

	// Capture the error with Sentry
	sentry.CaptureMessage(msg)
}

func Panic(msg string, fields ...zap.Field) {
	MustDefaultLogger().WithOptions(zap.AddCallerSkip(1)).Panic(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	MustDefaultLogger().WithOptions(zap.AddCallerSkip(1)).Fatal(msg, fields...)
}

func DefaultLogger() *zap.Logger {
	return MustDefaultLogger()
}

func Sugar() *zap.SugaredLogger {
	return MustDefaultLogger().Sugar()
}

// Convert zap fields to a map that Sentry can understand
func zapFieldsToMap(fields []zap.Field) map[string]interface{} {
	data := make(map[string]interface{})
	for _, field := range fields {
		var value interface{}
		encoder := zapcore.NewMapObjectEncoder()
		field.AddTo(encoder)
		for _, v := range encoder.Fields {
			value = v
		}

		data[field.Key] = value
	}
	return data
}
