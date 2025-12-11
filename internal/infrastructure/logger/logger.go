package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/seldomhappy/vibe_architecture/internal/infrastructure/config"
)

// ILogger defines logger interface
type ILogger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
	Fatal(format string, args ...interface{})
}

// Logger implements ILogger
type Logger struct {
	logger *log.Logger
	level  string
	format string
}

// NewLogger creates a new logger instance
func NewLogger(cfg config.LoggerConfig) ILogger {
	logger := log.New(os.Stdout, "", log.LstdFlags)

	return &Logger{
		logger: logger,
		level:  cfg.Level,
		format: cfg.Format,
	}
}

// Debug logs debug messages
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.shouldLog("debug") {
		l.log("DEBUG", format, args...)
	}
}

// Info logs info messages
func (l *Logger) Info(format string, args ...interface{}) {
	if l.shouldLog("info") {
		l.log("INFO", format, args...)
	}
}

// Warn logs warning messages
func (l *Logger) Warn(format string, args ...interface{}) {
	if l.shouldLog("warn") {
		l.log("WARN", format, args...)
	}
}

// Error logs error messages
func (l *Logger) Error(format string, args ...interface{}) {
	if l.shouldLog("error") {
		l.log("ERROR", format, args...)
	}
}

// Fatal logs fatal messages and exits
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log("FATAL", format, args...)
	os.Exit(1)
}

func (l *Logger) log(level string, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if l.format == "json" {
		// Use json.Marshal to properly escape the message
		msgBytes, _ := json.Marshal(msg)
		l.logger.Printf(`{"level":"%s","message":%s}`, level, string(msgBytes))
	} else {
		l.logger.Printf("[%s] %s", level, msg)
	}
}

func (l *Logger) shouldLog(level string) bool {
	levels := map[string]int{
		"debug": 0,
		"info":  1,
		"warn":  2,
		"error": 3,
		"fatal": 4,
	}

	configLevel := levels[l.level]
	messageLevel := levels[level]

	return messageLevel >= configLevel
}
