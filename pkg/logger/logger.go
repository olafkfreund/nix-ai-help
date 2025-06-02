package logger

import (
	"io"
	"log"
	"os"
	"strings"
)

// LogLevel represents the logging level
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

// Logger is a custom logger that provides structured logging capabilities.
type Logger struct {
	*log.Logger
	level LogLevel
}

// NewLogger creates a new instance of Logger with default info level.
func NewLogger() *Logger {
	return &Logger{
		Logger: log.New(os.Stdout, "", log.LstdFlags),
		level:  InfoLevel,
	}
}

// NewLoggerWithWriter creates a new Logger with a custom writer and default info level.
func NewLoggerWithWriter(w io.Writer) *Logger {
	return &Logger{
		Logger: log.New(w, "", log.LstdFlags),
		level:  InfoLevel,
	}
}

// NewLoggerWithLevel creates a new instance of Logger with specified level.
func NewLoggerWithLevel(levelStr string) *Logger {
	level := InfoLevel
	switch strings.ToLower(levelStr) {
	case "debug":
		level = DebugLevel
	case "info":
		level = InfoLevel
	case "warn", "warning":
		level = WarnLevel
	case "error":
		level = ErrorLevel
	}

	return &Logger{
		Logger: log.New(os.Stderr, "", log.LstdFlags),
		level:  level,
	}
}

// NewLoggerWithLevelAndWriter creates a new Logger with a custom writer and specified level.
func NewLoggerWithLevelAndWriter(levelStr string, w io.Writer) *Logger {
	level := InfoLevel
	switch strings.ToLower(levelStr) {
	case "debug":
		level = DebugLevel
	case "info":
		level = InfoLevel
	case "warn", "warning":
		level = WarnLevel
	case "error":
		level = ErrorLevel
	}
	return &Logger{
		Logger: log.New(w, "", log.LstdFlags),
		level:  level,
	}
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(levelStr string) {
	switch strings.ToLower(levelStr) {
	case "debug":
		l.level = DebugLevel
	case "info":
		l.level = InfoLevel
	case "warn", "warning":
		l.level = WarnLevel
	case "error":
		l.level = ErrorLevel
	}
}

// Info logs an informational message.
func (l *Logger) Info(msg string) {
	if l.level <= InfoLevel {
		l.Println("INFO: " + msg)
	}
}

// Warn logs a warning message.
func (l *Logger) Warn(msg string) {
	if l.level <= WarnLevel {
		l.Println("WARN: " + msg)
	}
}

// Error logs an error message.
func (l *Logger) Error(msg string) {
	if l.level <= ErrorLevel {
		l.Println("ERROR: " + msg)
	}
}

// Debug logs a debug message.
func (l *Logger) Debug(msg string) {
	if l.level <= DebugLevel {
		l.Println("DEBUG: " + msg)
	}
}
