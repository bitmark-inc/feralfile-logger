package log

import (
	"context"
	"fmt"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

type cloudflareLogger struct {
}

func (l *cloudflareLogger) Printf(format string, args ...any) {
	Logger().Info(fmt.Sprintf(format, args...), zap.Any("args", args))
}

func CloudflareLogger() *cloudflareLogger {
	return &cloudflareLogger{}
}

type cadenceWorkflowLogger struct {
	ctx    workflow.Context
	logger *zap.Logger
}

func CadenceWorkflowLogger(ctx workflow.Context) *cadenceWorkflowLogger {
	return &cadenceWorkflowLogger{
		ctx:    ctx,
		logger: workflow.GetLogger(ctx),
	}
}

func (l *cadenceWorkflowLogger) attachInfo(fields ...zap.Field) []zap.Field {
	fields = append(fields, zap.String("workflow_id", workflow.GetInfo(l.ctx).WorkflowExecution.ID))
	fields = append(fields, zap.String("workflow_type", workflow.GetInfo(l.ctx).WorkflowType.Name))
	fields = append(fields, zap.String("workflow_run_id", workflow.GetInfo(l.ctx).WorkflowExecution.RunID))
	return fields
}

func (l *cadenceWorkflowLogger) Debug(msg string, fields ...zap.Field) {
	ldebug(l.logger, msg, hub(), l.attachInfo(fields...)...)
}

func (l *cadenceWorkflowLogger) Info(msg string, fields ...zap.Field) {
	linfo(l.logger, msg, hub(), l.attachInfo(fields...)...)
}

func (l *cadenceWorkflowLogger) Warn(msg string, fields ...zap.Field) {
	lwarn(l.logger, msg, hub(), l.attachInfo(fields...)...)
}

func (l *cadenceWorkflowLogger) Error(msg string, fields ...zap.Field) {
	lerror(l.logger, msg, hub(), l.attachInfo(fields...)...)
}

type cadenceActivityLogger struct {
	ctx    context.Context
	logger *zap.Logger
}

func CadenceActivityLogger(ctx context.Context) *cadenceActivityLogger {
	return &cadenceActivityLogger{
		ctx:    ctx,
		logger: activity.GetLogger(ctx),
	}
}

func (l *cadenceActivityLogger) attachInfo(fields ...zap.Field) []zap.Field {
	fields = append(fields, zap.String("workflow_id", activity.GetInfo(l.ctx).WorkflowExecution.ID))
	fields = append(fields, zap.String("workflow_run_id", activity.GetInfo(l.ctx).WorkflowExecution.RunID))
	fields = append(fields, zap.String("activity_id", activity.GetInfo(l.ctx).ActivityID))
	fields = append(fields, zap.String("activity_type", activity.GetInfo(l.ctx).ActivityType.Name))
	wft := activity.GetInfo(l.ctx).WorkflowType
	if wft != nil {
		fields = append(fields, zap.String("workflow_type", wft.Name))
	}

	return fields
}

func (l *cadenceActivityLogger) Debug(msg string, fields ...zap.Field) {
	ldebug(l.logger, msg, hubOnContext(l.ctx), l.attachInfo(fields...)...)
}

func (l *cadenceActivityLogger) Info(msg string, fields ...zap.Field) {
	linfo(l.logger, msg, hubOnContext(l.ctx), l.attachInfo(fields...)...)
}

func (l *cadenceActivityLogger) Warn(msg string, fields ...zap.Field) {
	lwarn(l.logger, msg, hubOnContext(l.ctx), l.attachInfo(fields...)...)
}

func (l *cadenceActivityLogger) Error(msg string, fields ...zap.Field) {
	lerror(l.logger, msg, hubOnContext(l.ctx), l.attachInfo(fields...)...)
}
