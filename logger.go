package log

import (
	"context"

	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

// Initialize creates a new logger
func Initialize(isDebug bool, sentryOptions *sentry.ClientOptions) error {
	// init logger
	log, err := newLogger(isDebug)
	if err != nil {
		return err
	}

	logger = log

	// init sentry
	if sentryOptions != nil {
		err := sentry.Init(*sentryOptions)
		if err != nil {
			return err
		}
	}

	return nil
}

// newLogger creates a new logger
func newLogger(isDebug bool) (*zap.Logger, error) {
	var config zap.Config
	if isDebug {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	return config.Build()
}

// Logger returns the logger
func Logger() *zap.Logger {
	if logger == nil {
		panic("logger not initialized")
	}
	return logger
}

// Debug logs a message at Debug level.
func Debug(msg string, fields ...zap.Field) {
	ldebug(Logger(), msg, hub(), fields...)
}

func ldebug(logger *zap.Logger, msg string, hub *sentry.Hub, fields ...zap.Field) {
	logger.WithOptions(zap.AddCallerSkip(1)).Debug(msg, fields...)
}

// Info logs a message at Info level.
func Info(msg string, fields ...zap.Field) {
	linfo(Logger(), msg, hub(), fields...)
}

// InfoWithContext logs a message at Info level with a context.
func InfoWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	linfo(Logger(), msg, hubOnContext(ctx), fields...)
}

func linfo(logger *zap.Logger, msg string, hub *sentry.Hub, fields ...zap.Field) {
	logger.WithOptions(zap.AddCallerSkip(1)).Info(msg, fields...)

	// add a breadcrumb
	addBreadcrumb(hub, &sentry.Breadcrumb{
		Message: msg,
		Level:   sentry.LevelInfo,
		Data:    zapFieldsToMap(fields),
	})
}

// Warn logs a message at Warning level.
func Warn(msg string, fields ...zap.Field) {
	lwarn(Logger(), msg, hub(), fields...)
}

// Warn logs a message at Warning level.
func WarnWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	lwarn(Logger(), msg, hubOnContext(ctx), fields...)
}

func lwarn(logger *zap.Logger, msg string, hub *sentry.Hub, fields ...zap.Field) {
	logger.WithOptions(zap.AddCallerSkip(1)).Warn(msg, fields...)

	// add a breadcrumb
	addBreadcrumb(hub, &sentry.Breadcrumb{
		Message: msg,
		Level:   sentry.LevelWarning,
		Data:    zapFieldsToMap(fields),
	})
}

// Error logs a message at Error level.
func Error(msg string, fields ...zap.Field) {
	lerror(Logger(), msg, hub(), fields...)
}

// ErrorWithContext logs a message at Error level with a context.
func ErrorWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	lerror(Logger(), msg, hubOnContext(ctx), fields...)
}

func lerror(logger *zap.Logger, msg string, hub *sentry.Hub, fields ...zap.Field) {
	logger.WithOptions(zap.AddCallerSkip(1)).Error(msg, fields...)

	// add a breadcrumb
	addBreadcrumb(hub, &sentry.Breadcrumb{
		Message: msg,
		Level:   sentry.LevelError,
		Data:    zapFieldsToMap(fields),
	})

	// capture message
	captureMessage(msg, hub)
}

// Panic logs a message at Panic level.
func Panic(msg string, fields ...zap.Field) {
	lpanic(Logger(), msg, hub(), fields...)
}

// PanicWithContext logs a message at Panic level with a context.
func PanicWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	lpanic(Logger(), msg, hubOnContext(ctx), fields...)
}

func lpanic(logger *zap.Logger, msg string, hub *sentry.Hub, fields ...zap.Field) {
	// add a breadcrumb
	addBreadcrumb(hub, &sentry.Breadcrumb{
		Message: msg,
		Level:   sentry.LevelFatal,
		Data:    zapFieldsToMap(fields),
	})

	// capture message
	captureMessage(msg, hub)

	// panic
	logger.WithOptions(zap.AddCallerSkip(1)).Panic(msg, fields...)
}

// Fatal logs a message at Fatal level.
func Fatal(msg string, fields ...zap.Field) {
	lfatal(Logger(), msg, hub(), fields...)
}

// FatalWithContext logs a message at Fatal level with a context.
func FatalWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	lfatal(Logger(), msg, hubOnContext(ctx), fields...)
}

func lfatal(logger *zap.Logger, msg string, hub *sentry.Hub, fields ...zap.Field) {
	// add a breadcrumb
	addBreadcrumb(hub, &sentry.Breadcrumb{
		Message: msg,
		Level:   sentry.LevelFatal,
		Data:    zapFieldsToMap(fields),
	})

	// capture message
	captureMessage(msg, hub)

	logger.WithOptions(zap.AddCallerSkip(1)).Fatal(msg, fields...)
}

// Sugar returns a sugared Logger().
func Sugar() *zap.SugaredLogger {
	return Logger().Sugar()
}

// Sync flushes any buffered log entries.
func Sync() error {
	return Logger().Sync()
}

// addBreadcrumb adds a breadcrumb to Sentry
func addBreadcrumb(hub *sentry.Hub, breadcrumb *sentry.Breadcrumb) {
	if nil != hub {
		hub.AddBreadcrumb(breadcrumb, nil)
	} else {
		sentry.AddBreadcrumb(breadcrumb)
	}
}

// captureMessage captures a message to Sentry
func captureMessage(msg string, hub *sentry.Hub) {
	if nil != hub {
		hub.CaptureMessage(msg)
	} else {
		sentry.CaptureMessage(msg)
	}
}

// hubOnContext returns a hub from the context, creating a new one if it doesn't exist
func hubOnContext(ctx context.Context) *sentry.Hub {
	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub().Clone()
		sentry.SetHubOnContext(ctx, hub)
	}
	return hub
}

// hub returns the current hub
func hub() *sentry.Hub {
	return sentry.CurrentHub()
}

// zapFieldsToMap converts zap fields to a map that Sentry can understand
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
