package log

import (
	"fmt"

	"go.uber.org/zap"
)

type cloudflareLogger struct {
}

func (l *cloudflareLogger) Printf(format string, args ...any) {
	DefaultLogger().Info(fmt.Sprintf(format, args...), zap.Any("args", args))
}

func CloudflareLogger() *cloudflareLogger {
	return &cloudflareLogger{}
}
