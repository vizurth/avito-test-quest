package logger

import (
	"context"

	"go.uber.org/zap"
)

type key string

const (
	KeyForLogger    key = "logger"
	KeyForRequestID key = "requestID"
)

// Logger обёртка над zap.Logger
type Logger struct {
	l *zap.Logger
}

// New создаёт новый логгер и добавляет его в контекст
func New(ctx context.Context) (context.Context, *Logger, error) {
	l, err := zap.NewProduction()
	if err != nil {
		return ctx, nil, err
	}
	wrapped := &Logger{l: l}
	ctx = context.WithValue(ctx, KeyForLogger, wrapped)
	return ctx, wrapped, nil
}

// GetLoggerFromCtx безопасно достаёт логгер из контекста
func GetLoggerFromCtx(ctx context.Context) *Logger {
	if l, ok := ctx.Value(KeyForLogger).(*Logger); ok {
		return l
	}
	return nil
}

// GetOrCreateLoggerFromCtx возвращает логгер из контекста или создаёт новый
func GetOrCreateLoggerFromCtx(ctx context.Context) *Logger {
	if l := GetLoggerFromCtx(ctx); l != nil {
		return l
	}
	l, _ := zap.NewProduction()
	return &Logger{l: l}
}

// Debug логирует сообщение уровня Debug с полями
func (l *Logger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	if ctx.Value(KeyForRequestID) != nil {
		fields = append(fields, zap.String(string(KeyForRequestID), ctx.Value(KeyForRequestID).(string)))
	}
	l.l.Debug(msg, fields...)
}

// Info логирует сообщение уровня Info с полями
func (l *Logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	if ctx.Value(KeyForRequestID) != nil {
		fields = append(fields, zap.String(string(KeyForRequestID), ctx.Value(KeyForRequestID).(string)))
	}
	l.l.Info(msg, fields...)
}

// Warn логирует сообщение уровня Warn с полями
func (l *Logger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	if ctx.Value(KeyForRequestID) != nil {
		fields = append(fields, zap.String(string(KeyForRequestID), ctx.Value(KeyForRequestID).(string)))
	}
	l.l.Warn(msg, fields...)
}

// Error логирует сообщение уровня Error с полями
func (l *Logger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	if ctx.Value(KeyForRequestID) != nil {
		fields = append(fields, zap.String(string(KeyForRequestID), ctx.Value(KeyForRequestID).(string)))
	}
	l.l.Error(msg, fields...)
}

// Fatal логирует сообщение уровня Fatal с полями и завершает программу
func (l *Logger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	if ctx.Value(KeyForRequestID) != nil {
		fields = append(fields, zap.String(string(KeyForRequestID), ctx.Value(KeyForRequestID).(string)))
	}
	l.l.Fatal(msg, fields...)
}
