package log

import (
	"context"
	"strings"

	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var defaultLogger *zap.Logger

func Initialize(level string, isDebug bool, sentryOptions *sentry.ClientOptions) error {
	log, err := New(level, isDebug)
	if err != nil {
		return err
	}

	defaultLogger = log

	// Init sentry
	if sentryOptions != nil {
		err := sentry.Init(*sentryOptions)
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

func SetSentryHubOnContext(ctx context.Context) context.Context {
	hub := sentry.CurrentHub().Clone()
	return sentry.SetHubOnContext(ctx, hub)
}

func Debug(msg string, fields ...zap.Field) {
	DefaultLogger().WithOptions(zap.AddCallerSkip(1)).Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	info(msg, sentry.CurrentHub(), fields...)
}

func InfoWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	info(msg, sentry.GetHubFromContext(ctx), fields...)
}

func info(msg string, hub *sentry.Hub, fields ...zap.Field) {
	DefaultLogger().WithOptions(zap.AddCallerSkip(1)).Info(msg, fields...)

	// Add a breadcrumb
	addBreadcrumb(hub, &sentry.Breadcrumb{
		Message: msg,
		Level:   sentry.LevelInfo,
		Data:    zapFieldsToMap(fields),
	})
}

func Warn(msg string, fields ...zap.Field) {
	warn(msg, sentry.CurrentHub(), fields...)
}

func WarnWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	warn(msg, sentry.GetHubFromContext(ctx), fields...)
}

func warn(msg string, hub *sentry.Hub, fields ...zap.Field) {
	DefaultLogger().WithOptions(zap.AddCallerSkip(1)).Warn(msg, fields...)

	// Add a breadcrumb
	addBreadcrumb(hub, &sentry.Breadcrumb{
		Message: msg,
		Level:   sentry.LevelWarning,
		Data:    zapFieldsToMap(fields),
	})
}

func Error(msg string, fields ...zap.Field) {
	err(msg, sentry.CurrentHub(), fields...)
}

func ErrorWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	err(msg, sentry.GetHubFromContext(ctx), fields...)
}

func err(msg string, hub *sentry.Hub, fields ...zap.Field) {
	DefaultLogger().WithOptions(zap.AddCallerSkip(1)).Error(msg, fields...)

	// Add a breadcrumb
	addBreadcrumb(hub, &sentry.Breadcrumb{
		Message: msg,
		Level:   sentry.LevelError,
		Data:    zapFieldsToMap(fields),
	})
}

func addBreadcrumb(hub *sentry.Hub, breadcrumb *sentry.Breadcrumb) {
	if nil != hub {
		hub.AddBreadcrumb(breadcrumb, nil)
	} else {
		sentry.AddBreadcrumb(breadcrumb)
	}
}

func Panic(msg string, fields ...zap.Field) {
	DefaultLogger().WithOptions(zap.AddCallerSkip(1)).Panic(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	DefaultLogger().WithOptions(zap.AddCallerSkip(1)).Fatal(msg, fields...)
}

func Sugar() *zap.SugaredLogger {
	return DefaultLogger().Sugar()
}

func Sync() error {
	return DefaultLogger().Sync()
}

func DefaultLogger() *zap.Logger {
	if defaultLogger == nil {
		panic("use logger without initializing")
	}

	return defaultLogger
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
