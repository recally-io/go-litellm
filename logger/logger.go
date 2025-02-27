package logger

import (
	"log/slog"
	"os"
	"strings"
)

var DefaultLogger = New()

// Logger is a wrapper around slog.Logger
type Logger struct {
	*slog.Logger
}

// Debug logs a message at level Fatal on the standard logger.
// it will exit the program after logging
func (l Logger) Fatal(msg string, args ...interface{}) {
	l.Error(msg, args...)
	os.Exit(1)
}

// New creates a new logger
func New() Logger {

	// load logLevel from env, if not set, use info
	level := slog.LevelInfo
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel != "" && strings.ToLower(logLevel) == "debug" {
		level = slog.LevelDebug
	}

	var logger *slog.Logger

	handlers := make([]slog.Handler, 0)

	if level >= slog.LevelDebug {
		handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level:     slog.LevelDebug,
			AddSource: true,
		})
		handlers = append(handlers, handler)
	} else {
		handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     level,
		})
		handlers = append(handlers, handler)
	}

	logger = slog.New(NewMultiHandler(handlers...))
	return Logger{
		Logger: logger,
	}
}
