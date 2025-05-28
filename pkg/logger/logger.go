package logger

import (
	"log"
	"os"
)

// Logger is a custom logger that provides structured logging capabilities.
type Logger struct {
	*log.Logger
}

// NewLogger creates a new instance of Logger.
func NewLogger() *Logger {
	return &Logger{
		Logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

// Info logs an informational message.
func (l *Logger) Info(msg string) {
	l.Println("INFO: " + msg)
}

// Warn logs a warning message.
func (l *Logger) Warn(msg string) {
	l.Println("WARN: " + msg)
}

// Error logs an error message.
func (l *Logger) Error(msg string) {
	l.Println("ERROR: " + msg)
}

// Debug logs a debug message.
func (l *Logger) Debug(msg string) {
	l.Println("DEBUG: " + msg)
}
