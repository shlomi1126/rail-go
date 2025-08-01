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
		// In case of error, return a no-op logger
		logger = zap.NewNop()
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
		// In case of error, return a no-op logger
		logger = zap.NewNop()
	}

	return &Logger{Logger: logger}
}

// Info logs an info message with structured fields
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.Logger.Info(msg, fields...)
}

// Error logs an error message with structured fields
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.Logger.Error(msg, fields...)
}

// Warn logs a warning message with structured fields
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.Logger.Warn(msg, fields...)
}

// Debug logs a debug message with structured fields
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.Logger.Debug(msg, fields...)
}

// Fatal logs a fatal message with structured fields and exits
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.Logger.Fatal(msg, fields...)
}
