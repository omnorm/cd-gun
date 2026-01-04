package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// LogLevel represents the severity level of a log message
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

// Logger provides logging functionality for cd-gun
type Logger struct {
	level    LogLevel
	debugLog *log.Logger
	infoLog  *log.Logger
	warnLog  *log.Logger
	errorLog *log.Logger
	out      io.Writer
}

// NewLogger creates a new logger with the specified level
func NewLogger(level string, out io.Writer) *Logger {
	if out == nil {
		out = os.Stdout
	}

	lvl := parseLevel(level)

	return &Logger{
		level:    lvl,
		debugLog: log.New(out, "[DEBUG] ", log.LstdFlags|log.Lshortfile),
		infoLog:  log.New(out, "[INFO] ", log.LstdFlags),
		warnLog:  log.New(out, "[WARN] ", log.LstdFlags),
		errorLog: log.New(out, "[ERROR] ", log.LstdFlags|log.Lshortfile),
		out:      out,
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...interface{}) {
	if l.level <= DebugLevel {
		l.debugLog.Printf(msg, args...)
	}
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...interface{}) {
	if l.level <= InfoLevel {
		l.infoLog.Printf(msg, args...)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...interface{}) {
	if l.level <= WarnLevel {
		l.warnLog.Printf(msg, args...)
	}
}

// Error logs an error message
func (l *Logger) Error(msg string, args ...interface{}) {
	if l.level <= ErrorLevel {
		l.errorLog.Printf(msg, args...)
	}
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Debug(format, args...)
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.Info(format, args...)
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.Warn(format, args...)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Error(format, args...)
}

// SetLevel changes the log level
func (l *Logger) SetLevel(level string) {
	l.level = parseLevel(level)
}

// parseLevel converts a string to LogLevel
func parseLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	default:
		return InfoLevel
	}
}

// Printf is a convenience method that always prints (ignores level)
func (l *Logger) Printf(format string, args ...interface{}) {
	fmt.Fprintf(l.out, format, args...)
}
