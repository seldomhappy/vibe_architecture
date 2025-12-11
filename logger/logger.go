package logger

import (
	"fmt"
	"log"
	"os"
)

// ILogger defines the logging interface
type ILogger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
	Fatal(format string, args ...interface{})
}

// Logger implements ILogger
type Logger struct {
	appName string
	logger  *log.Logger
}

// New creates a new logger instance
func New(appName string) ILogger {
	return &Logger{
		appName: appName,
		logger:  log.New(os.Stdout, "", log.LstdFlags),
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log("DEBUG", format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log("INFO", format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log("WARN", format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log("ERROR", format, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log("FATAL", format, args...)
	os.Exit(1)
}

func (l *Logger) log(level, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	l.logger.Printf("[%s] [%s] %s", l.appName, level, message)
}
