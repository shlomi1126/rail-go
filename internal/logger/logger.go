package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
}

// New creates a new logger instance with production configuration
func New() *Logger {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.MessageKey = "message"
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.CallerKey = "" // Remove caller information
	config.DisableCaller = true         // Disable caller logging
	config.DisableStacktrace = true     // Disable stacktrace logging

	logger, err := config.Build()
	if err != nil {
		panic("failed to create logger: " + err.Error())
	}

	return &Logger{Logger: logger}
}

// NewDevelopment creates a new logger instance with development configuration
func NewDevelopment() *Logger {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.CallerKey = "" // Remove caller information
	config.DisableCaller = true         // Disable caller logging
	config.DisableStacktrace = true     // Disable stacktrace logging

	logger, err := config.Build()
	if err != nil {
		panic("failed to create logger: " + err.Error())
	}

	return &Logger{Logger: logger}
}

// Info logs an info message with structured fields
func (l *Logger) Info(msg string, fields ...interface{}) {
	zapFields := convertToZapFields(fields...)
	l.Logger.Info(msg, zapFields...)
}

// Error logs an error message with structured fields
func (l *Logger) Error(msg string, fields ...interface{}) {
	zapFields := convertToZapFields(fields...)
	l.Logger.Error(msg, zapFields...)
}

// Debug logs a debug message with structured fields
func (l *Logger) Debug(msg string, fields ...interface{}) {
	zapFields := convertToZapFields(fields...)
	l.Logger.Debug(msg, zapFields...)
}

// Fatal logs a fatal message with structured fields and exits
func (l *Logger) Fatal(msg string, fields ...interface{}) {
	zapFields := convertToZapFields(fields...)
	l.Logger.Fatal(msg, zapFields...)
}

// convertToZapFields converts variadic interface{} to zap.Field slice
func convertToZapFields(fields ...interface{}) []zapcore.Field {
	if len(fields) == 0 {
		return nil
	}

	zapFields := make([]zapcore.Field, 0, len(fields)/2)
	for i := 0; i < len(fields); i += 2 {
		if i+1 >= len(fields) {
			break
		}
		key, ok := fields[i].(string)
		if !ok {
			continue
		}
		zapFields = append(zapFields, zap.Any(key, fields[i+1]))
	}
	return zapFields
}
