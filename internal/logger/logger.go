package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Logger struct {
	logger zerolog.Logger
}

func New() *Logger {
	// Configure zerolog
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Create a logger with pretty console output
	logger := log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	})

	return &Logger{
		logger: logger,
	}
}

func (l *Logger) Info(msg string, fields ...interface{}) {
	event := l.logger.Info()
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			key, ok := fields[i].(string)
			if !ok {
				continue
			}
			event.Interface(key, fields[i+1])
		}
	}
	event.Msg(msg)
}

func (l *Logger) Error(msg string, fields ...interface{}) {
	event := l.logger.Error()
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			key, ok := fields[i].(string)
			if !ok {
				continue
			}
			event.Interface(key, fields[i+1])
		}
	}
	event.Msg(msg)
}

func (l *Logger) Debug(msg string, fields ...interface{}) {
	event := l.logger.Debug()
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			key, ok := fields[i].(string)
			if !ok {
				continue
			}
			event.Interface(key, fields[i+1])
		}
	}
	event.Msg(msg)
}

func (l *Logger) Fatal(msg string, fields ...interface{}) {
	event := l.logger.Fatal()
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			key, ok := fields[i].(string)
			if !ok {
				continue
			}
			event.Interface(key, fields[i+1])
		}
	}
	event.Msg(msg)
}
