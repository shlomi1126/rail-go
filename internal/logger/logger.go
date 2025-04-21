package logger

import (
	"log"
	"os"
	"sync"
)

type Logger struct {
	*log.Logger
	mu sync.Mutex
}

func New() *Logger {
	return &Logger{
		Logger: log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile),
	}
}

func (l *Logger) Info(msg string, fields ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Printf("[INFO] %s %v", msg, fields)
}

func (l *Logger) Error(msg string, fields ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Printf("[ERROR] %s %v", msg, fields)
}

func (l *Logger) Debug(msg string, fields ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Printf("[DEBUG] %s %v", msg, fields)
}

func (l *Logger) Fatal(msg string, fields ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Fatalf("[FATAL] %s %v", msg, fields)
} 