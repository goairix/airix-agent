package logger

import (
	"context"
	"log"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type logger struct {
	_logger *zap.Logger
}

type Field struct {
	Key   string
	Value interface{}
}

// ErrorField 错误堆栈
func ErrorField(err error) Field {
	return Field{Key: "error_field", Value: err}
}

func AddField(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

func ZapLogger() *zap.Logger {
	return _logger._logger
}

var _logger *logger

// InitLogger 初始化日志
func InitLogger(debug bool) {
	newZapLogger(debug)
	_logger = &logger{_logger: _zapLogger}
}

func (l *logger) log(ctx context.Context, level zapcore.Level, message string, fields ...Field) {
	// logger 未初始化时降级到标准库日志
	if l == nil || l._logger == nil {
		if level == zapcore.FatalLevel {
			log.Fatalf("[FATAL] %s %v", message, fieldsToArgs(fields))
		}
		log.Printf("[%s] %s %v", level.CapitalString(), message, fieldsToArgs(fields))
		return
	}

	fields = append(
		fields,
		l.trace(ctx)...,
	)

	zipFields := make([]zap.Field, 0, len(fields))
	for _, field := range fields {
		if field.Key == "error_field" {
			zipFields = append(zipFields, zap.Error(field.Value.(error)))
		} else {
			zipFields = append(zipFields, zap.Any(field.Key, field.Value))
		}
	}

	check := l._logger.Check(level, message)
	check.Write(zipFields...)
}

func (l *logger) trace(ctx context.Context) []Field {
	span := trace.SpanFromContext(ctx)
	var fields []Field
	if span.SpanContext().HasTraceID() {
		fields = append(fields, Field{Key: "trace_id", Value: span.SpanContext().TraceID().String()})
	}
	if span.SpanContext().HasSpanID() {
		fields = append(fields, Field{Key: "span_id", Value: span.SpanContext().SpanID().String()})
	}
	if spanIns, ok := span.(sdktrace.ReadWriteSpan); ok {
		fields = append(fields, Field{Key: "span_name", Value: spanIns.Name()})
		if spanIns.Parent().HasSpanID() {
			fields = append(fields, Field{Key: "parent_span_id", Value: spanIns.Parent().SpanID().String()})
		}
	}
	return fields
}

// fieldsToArgs 将 Field 切片转为适合标准库日志输出的参数
func fieldsToArgs(fields []Field) []interface{} {
	args := make([]interface{}, 0, len(fields)*2)
	for _, f := range fields {
		args = append(args, f.Key, f.Value)
	}
	return args
}

func Debug(ctx context.Context, message string, fields ...Field) {
	_logger.log(ctx, zapcore.DebugLevel, message, fields...)
}

func Info(ctx context.Context, message string, fields ...Field) {
	_logger.log(ctx, zapcore.InfoLevel, message, fields...)
}

func Warn(ctx context.Context, message string, fields ...Field) {
	_logger.log(ctx, zapcore.WarnLevel, message, fields...)
}

func Error(ctx context.Context, message string, fields ...Field) {
	_logger.log(ctx, zapcore.ErrorLevel, message, fields...)
}

func Fatal(ctx context.Context, message string, fields ...Field) {
	_logger.log(ctx, zapcore.FatalLevel, message, fields...)
}

func Panic(ctx context.Context, message string, fields ...Field) {
	_logger.log(ctx, zapcore.PanicLevel, message, fields...)
}

func WithOptions(opts ...zap.Option) *zap.Logger {
	return _logger._logger.WithOptions(opts...)
}
