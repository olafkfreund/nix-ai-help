package logger

import "testing"

func TestLoggerInfo(t *testing.T) {
	l := NewLogger()
	l.Info("test info message")
}
