package logging

import (
	"sync"
)

// LoggerFactory creates loggers
type LoggerFactory interface {
	// CreateLogger creates a new logger
	CreateLogger(name string) Logger
}

// VoidLoggerFactory creates void loggers
type VoidLoggerFactory struct{}

// NewVoidLoggerFactory creates a new VoidLoggerFactory
func NewVoidLoggerFactory() *VoidLoggerFactory {
	return &VoidLoggerFactory{}
}

// CreateLogger implements LoggerFactory.CreateLogger
func (f *VoidLoggerFactory) CreateLogger(name string) Logger {
	return &VoidLogger{}
}

// LazyLoggerFactory creates loggers lazily
type LazyLoggerFactory struct {
	factory LoggerFactory
	loggers map[string]Logger
	mu      sync.RWMutex
}

// NewLazyLoggerFactory creates a new LazyLoggerFactory
func NewLazyLoggerFactory(factory LoggerFactory) *LazyLoggerFactory {
	return &LazyLoggerFactory{
		factory: factory,
		loggers: make(map[string]Logger),
	}
}

// CreateLogger implements LoggerFactory.CreateLogger
func (f *LazyLoggerFactory) CreateLogger(name string) Logger {
	f.mu.RLock()
	if logger, ok := f.loggers[name]; ok {
		f.mu.RUnlock()
		return logger
	}
	f.mu.RUnlock()

	f.mu.Lock()
	defer f.mu.Unlock()

	// Double-check after acquiring write lock
	if logger, ok := f.loggers[name]; ok {
		return logger
	}

	logger := f.factory.CreateLogger(name)
	f.loggers[name] = logger
	return logger
}

// LogUtil provides utility functions for logging
type LogUtil struct{}

// NewLogUtil creates a new LogUtil
func NewLogUtil() *LogUtil {
	return &LogUtil{}
}

// GetLogLevel gets the log level from a string
func (u *LogUtil) GetLogLevel(level string) LogLevel {
	switch level {
	case "debug":
		return Debug
	case "info":
		return Info
	case "warn":
		return Warn
	case "error":
		return Error
	default:
		return Info
	}
}

// GetLogLevelString gets the string representation of a log level
func (u *LogUtil) GetLogLevelString(level LogLevel) string {
	switch level {
	case Debug:
		return "debug"
	case Info:
		return "info"
	case Warn:
		return "warn"
	case Error:
		return "error"
	default:
		return "info"
	}
}
