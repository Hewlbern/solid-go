package logging

import (
	"fmt"
	"io"
	"os"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	// Debug represents debug level logging
	Debug LogLevel = iota
	// Info represents info level logging
	Info
	// Warn represents warning level logging
	Warn
	// Error represents error level logging
	Error
)

// Logger is the interface for logging
type Logger interface {
	// Debug logs a debug message
	Debug(message string, meta ...interface{})
	// Info logs an info message
	Info(message string, meta ...interface{})
	// Warn logs a warning message
	Warn(message string, meta ...interface{})
	// Error logs an error message
	Error(message string, meta ...interface{})
	// Fatal logs a fatal message
	Fatal(message string, meta ...interface{})
}

// VoidLogger is a logger that does nothing
type VoidLogger struct{}

// Debug implements Logger.Debug
func (l *VoidLogger) Debug(message string, meta ...interface{}) {}

// Info implements Logger.Info
func (l *VoidLogger) Info(message string, meta ...interface{}) {}

// Warn implements Logger.Warn
func (l *VoidLogger) Warn(message string, meta ...interface{}) {}

// Error implements Logger.Error
func (l *VoidLogger) Error(message string, meta ...interface{}) {}

// Fatal implements Logger.Fatal
func (l *VoidLogger) Fatal(message string, meta ...interface{}) {}

// BasicLogger is a basic implementation of Logger
type BasicLogger struct {
	level LogLevel
	out   io.Writer
}

// NewBasicLogger creates a new BasicLogger
func NewBasicLogger(level LogLevel) *BasicLogger {
	return &BasicLogger{
		level: level,
		out:   os.Stdout,
	}
}

// Debug implements Logger.Debug
func (l *BasicLogger) Debug(message string, meta ...interface{}) {
	if l.level <= Debug {
		fmt.Fprintf(l.out, "[DEBUG] %s %v\n", message, meta)
	}
}

// Info implements Logger.Info
func (l *BasicLogger) Info(format string, args ...interface{}) {
	if l.level <= Info {
		fmt.Fprintf(l.out, "[INFO] "+format+"\n", args...)
	}
}

// Warn implements Logger.Warn
func (l *BasicLogger) Warn(format string, args ...interface{}) {
	if l.level <= Warn {
		fmt.Fprintf(l.out, "[WARN] "+format+"\n", args...)
	}
}

// Error implements Logger.Error
func (l *BasicLogger) Error(format string, args ...interface{}) {
	if l.level <= Error {
		fmt.Fprintf(l.out, "[ERROR] "+format+"\n", args...)
	}
}
