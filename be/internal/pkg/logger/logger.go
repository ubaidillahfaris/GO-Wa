package logger

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// Fields is a type alias for map[string]interface{} for convenience
type Fields map[string]interface{}

var levelNames = map[LogLevel]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

var levelEmojis = map[LogLevel]string{
	DEBUG: "🔍",
	INFO:  "ℹ️",
	WARN:  "⚠️",
	ERROR: "❌",
	FATAL: "💀",
}

// Logger represents a structured logger
type Logger struct {
	prefix   string
	level    LogLevel
	logger   *log.Logger
	fields   map[string]interface{}
	useEmoji bool
}

// New creates a new logger instance
func New(prefix string) *Logger {
	return &Logger{
		prefix:   prefix,
		level:    INFO,
		logger:   log.New(os.Stdout, "", 0),
		fields:   make(map[string]interface{}),
		useEmoji: true,
	}
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// WithPrefix returns a new logger with the specified prefix
func (l *Logger) WithPrefix(prefix string) *Logger {
	newLogger := &Logger{
		prefix:   prefix,
		level:    l.level,
		logger:   l.logger,
		fields:   make(map[string]interface{}),
		useEmoji: l.useEmoji,
	}
	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}
	return newLogger
}

// WithField adds a field to the logger
func (l *Logger) WithField(key string, value interface{}) *Logger {
	newLogger := &Logger{
		prefix:   l.prefix,
		level:    l.level,
		logger:   l.logger,
		fields:   make(map[string]interface{}),
		useEmoji: l.useEmoji,
	}
	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}
	// Add new field
	newLogger.fields[key] = value
	return newLogger
}

// WithFields adds multiple fields to the logger
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	newLogger := &Logger{
		prefix:   l.prefix,
		level:    l.level,
		logger:   l.logger,
		fields:   make(map[string]interface{}),
		useEmoji: l.useEmoji,
	}
	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}
	// Add new fields
	for k, v := range fields {
		newLogger.fields[k] = v
	}
	return newLogger
}

// WithContext extracts fields from context
func (l *Logger) WithContext(ctx context.Context) *Logger {
	// You can extract request ID, user ID, etc. from context here
	return l
}

// log is the internal logging function
func (l *Logger) log(level LogLevel, msg string, args ...interface{}) {
	if level < l.level {
		return
	}

	// Format message
	message := fmt.Sprintf(msg, args...)

	// Build log line
	var parts []string

	// Timestamp
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	parts = append(parts, timestamp)

	// Level with emoji
	levelStr := levelNames[level]
	if l.useEmoji {
		emoji := levelEmojis[level]
		parts = append(parts, fmt.Sprintf("[%s %s]", emoji, levelStr))
	} else {
		parts = append(parts, fmt.Sprintf("[%s]", levelStr))
	}

	// Prefix
	if l.prefix != "" {
		parts = append(parts, fmt.Sprintf("[%s]", l.prefix))
	}

	// Message
	parts = append(parts, message)

	// Fields
	if len(l.fields) > 0 {
		var fieldParts []string
		for k, v := range l.fields {
			fieldParts = append(fieldParts, fmt.Sprintf("%s=%v", k, v))
		}
		parts = append(parts, fmt.Sprintf("{%s}", strings.Join(fieldParts, ", ")))
	}

	// Print log
	logLine := strings.Join(parts, " ")
	l.logger.Println(logLine)

	// Fatal exits the program
	if level == FATAL {
		os.Exit(1)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.log(DEBUG, msg, args...)
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...interface{}) {
	l.log(INFO, msg, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.log(WARN, msg, args...)
}

// Error logs an error message
// Supports multiple signatures:
// - Error(msg string, args ...interface{}) - formatted message
// - Error(msg string, err error) - message with error
// - Error(msg string, err error, fields Fields) - message with error and fields
func (l *Logger) Error(msg string, args ...interface{}) {
	// Check if first arg is an error and second is Fields
	if len(args) >= 2 {
		if err, ok := args[0].(error); ok {
			if fields, ok := args[1].(Fields); ok {
				// Add error to fields
				fields["error"] = err.Error()
				l.WithFields(fields).log(ERROR, msg)
				return
			}
		}
	}

	// Check if first arg is an error
	if len(args) == 1 {
		if err, ok := args[0].(error); ok {
			l.WithField("error", err.Error()).log(ERROR, msg)
			return
		}
	}

	// Default formatted message
	l.log(ERROR, msg, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, args ...interface{}) {
	l.log(FATAL, msg, args...)
}

// ErrorWithStack logs an error with stack trace
func (l *Logger) ErrorWithStack(err error, msg string) {
	// Get stack trace
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	stack := string(buf[:n])

	l.WithField("error", err.Error()).
		WithField("stack", stack).
		Error(msg)
}

// Success logs a success message (INFO level with ✅ emoji)
func (l *Logger) Success(msg string, args ...interface{}) {
	message := fmt.Sprintf(msg, args...)
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	var parts []string
	parts = append(parts, timestamp)
	parts = append(parts, "[✅ SUCCESS]")

	if l.prefix != "" {
		parts = append(parts, fmt.Sprintf("[%s]", l.prefix))
	}

	parts = append(parts, message)

	if len(l.fields) > 0 {
		var fieldParts []string
		for k, v := range l.fields {
			fieldParts = append(fieldParts, fmt.Sprintf("%s=%v", k, v))
		}
		parts = append(parts, fmt.Sprintf("{%s}", strings.Join(fieldParts, ", ")))
	}

	l.logger.Println(strings.Join(parts, " "))
}

// Default logger instance
var defaultLogger = New("")

// SetDefaultLevel sets the level for the default logger
func SetDefaultLevel(level LogLevel) {
	defaultLogger.SetLevel(level)
}

// Package-level functions using default logger
func Debug(msg string, args ...interface{}) {
	defaultLogger.Debug(msg, args...)
}

func Info(msg string, args ...interface{}) {
	defaultLogger.Info(msg, args...)
}

func Warn(msg string, args ...interface{}) {
	defaultLogger.Warn(msg, args...)
}

func Error(msg string, args ...interface{}) {
	defaultLogger.Error(msg, args...)
}

func Fatal(msg string, args ...interface{}) {
	defaultLogger.Fatal(msg, args...)
}

func Success(msg string, args ...interface{}) {
	defaultLogger.Success(msg, args...)
}
