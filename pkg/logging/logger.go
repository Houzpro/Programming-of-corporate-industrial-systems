package logging

import (
	"io"
	"log/slog"
	"os"
)

type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

var Logger *slog.Logger

func Init(level LogLevel, output io.Writer) {
	var logLevel slog.Level
	switch level {
	case LogLevelDebug:
		logLevel = slog.LevelDebug
	case LogLevelWarn:
		logLevel = slog.LevelWarn
	case LogLevelError:
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	if output == nil {
		output = os.Stdout
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	handler := slog.NewJSONHandler(output, opts)
	Logger = slog.New(handler)
	slog.SetDefault(Logger)
}
