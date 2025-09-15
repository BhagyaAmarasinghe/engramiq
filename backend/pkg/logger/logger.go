package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap logger for structured logging
// Structured logging is crucial for debugging distributed systems
type Logger struct {
	*zap.SugaredLogger
}

// New creates a new logger instance
func New(environment string) *Logger {
	var logger *zap.Logger
	var err error

	if environment == "production" {
		// Production config: JSON format, no debug logs
		config := zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		
		logger, err = config.Build()
	} else {
		// Development config: Console format, colored output
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		
		logger, err = config.Build()
	}

	if err != nil {
		panic(err)
	}

	return &Logger{logger.Sugar()}
}

// WithContext adds context fields to the logger
func (l *Logger) WithContext(fields ...interface{}) *Logger {
	return &Logger{l.With(fields...)}
}

// WithRequestID adds request ID to all log entries
func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{l.With("request_id", requestID)}
}

// WithUserID adds user context to logs
func (l *Logger) WithUserID(userID string) *Logger {
	return &Logger{l.With("user_id", userID)}
}

// WithError adds error details to logs
func (l *Logger) WithError(err error) *Logger {
	return &Logger{l.With("error", err.Error())}
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() {
	l.SugaredLogger.Sync()
}