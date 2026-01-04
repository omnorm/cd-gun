package logger

import (
	"bytes"
	"testing"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name      string
		level     string
		wantPanic bool
	}{
		{"debug level", "debug", false},
		{"info level", "info", false},
		{"warn level", "warn", false},
		{"error level", "error", false},
		{"uppercase level", "DEBUG", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			log := NewLogger(tt.level, &buf)
			if log == nil {
				t.Error("NewLogger() returned nil")
			}
		})
	}
}

func TestLogMethods(t *testing.T) {
	var buf bytes.Buffer
	log := NewLogger("debug", &buf)

	// Test that these don't panic
	log.Debug("debug message")
	log.Info("info message")
	log.Warn("warn message")
	log.Error("error message")

	// Check that something was logged
	if buf.Len() == 0 {
		t.Error("Expected log output, but buffer is empty")
	}
}

func TestLogLevels(t *testing.T) {
	tests := []struct {
		name  string
		level string
		// These would be debug, info, warn, error messages respectively
		// We just verify they don't panic
	}{
		{"debug", "debug"},
		{"info", "info"},
		{"warn", "warn"},
		{"error", "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			log := NewLogger(tt.level, &buf)

			log.Debug("test debug")
			log.Info("test info")
			log.Warn("test warn")
			log.Error("test error")
		})
	}
}

func TestNewLoggerWithNilWriter(t *testing.T) {
	log := NewLogger("info", nil)
	if log == nil {
		t.Error("NewLogger() should not return nil when writer is nil")
	}

	// Should use os.Stdout as fallback
	log.Info("test message")
}
