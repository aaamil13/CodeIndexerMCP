package utils

import (
	"log"
	"os"
)

// Logger levels
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

var currentLevel = INFO

// Logger is a simple logger
type Logger struct {
	prefix string
	level  LogLevel
	logger *log.Logger
}

// NewLogger creates a new logger
func NewLogger(prefix string) *Logger {
	return &Logger{
		prefix: prefix,
		level:  currentLevel,
		logger: log.New(os.Stdout, prefix+" ", log.LstdFlags),
	}
}

// SetLevel sets the log level
func SetLevel(level LogLevel) {
	currentLevel = level
}

// Debug logs debug messages
func (l *Logger) Debug(v ...interface{}) {
	if l.level <= DEBUG {
		l.logger.Println("[DEBUG]", v)
	}
}

// Debugf logs formatted debug messages
func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.level <= DEBUG {
		l.logger.Printf("[DEBUG] "+format, v...)
	}
}

// Info logs info messages
func (l *Logger) Info(v ...interface{}) {
	if l.level <= INFO {
		l.logger.Println("[INFO]", v)
	}
}

// Infof logs formatted info messages
func (l *Logger) Infof(format string, v ...interface{}) {
	if l.level <= INFO {
		l.logger.Printf("[INFO] "+format, v...)
	}
}

// Warn logs warning messages
func (l *Logger) Warn(v ...interface{}) {
	if l.level <= WARN {
		l.logger.Println("[WARN]", v)
	}
}

// Warnf logs formatted warning messages
func (l *Logger) Warnf(format string, v ...interface{}) {
	if l.level <= WARN {
		l.logger.Printf("[WARN] "+format, v...)
	}
}

// Error logs error messages
func (l *Logger) Error(v ...interface{}) {
	if l.level <= ERROR {
		l.logger.Println("[ERROR]", v)
	}
}

// Errorf logs formatted error messages
func (l *Logger) Errorf(format string, v ...interface{}) {
	if l.level <= ERROR {
		l.logger.Printf("[ERROR] "+format, v...)
	}
}

// Global logger instance
var defaultLogger = NewLogger("[CodeIndexer]")

// Debug logs debug message using default logger
func Debug(v ...interface{}) {
	defaultLogger.Debug(v...)
}

// Debugf logs formatted debug message using default logger
func Debugf(format string, v ...interface{}) {
	defaultLogger.Debugf(format, v...)
}

// Info logs info message using default logger
func Info(v ...interface{}) {
	defaultLogger.Info(v...)
}

// Infof logs formatted info message using default logger
func Infof(format string, v ...interface{}) {
	defaultLogger.Infof(format, v...)
}

// Warn logs warning message using default logger
func Warn(v ...interface{}) {
	defaultLogger.Warn(v...)
}

// Warnf logs formatted warning message using default logger
func Warnf(format string, v ...interface{}) {
	defaultLogger.Warnf(format, v...)
}

// Error logs error message using default logger
func Error(v ...interface{}) {
	defaultLogger.Error(v...)
}

// Errorf logs formatted error message using default logger
func Errorf(format string, v ...interface{}) {
	defaultLogger.Errorf(format, v...)
}
