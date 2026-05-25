// Package logging provides a levelled, structured logger for the bootstrap
// simulation study.  It wraps the standard library's log/slog package so that
// the rest of the codebase only imports this thin adapter.
package logging

import (
	"io"
	"log/slog"
	"os"
)

// Level mirrors slog levels for callers that do not import log/slog directly.
type Level = slog.Level

const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
)

// Logger is the application-wide structured logger.
var Logger *slog.Logger

// Init configures the global Logger.
//
//   - level   : minimum level to emit (Debug, Info, Warn, Error)
//   - out     : destination writer (nil → os.Stdout)
//   - addJSON : true → JSON format; false → plain text
func Init(level Level, out io.Writer, addJSON bool) {
	if out == nil {
		out = os.Stdout
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: level == LevelDebug, // include file:line only at debug
	}

	var h slog.Handler
	if addJSON {
		h = slog.NewJSONHandler(out, opts)
	} else {
		h = slog.NewTextHandler(out, opts)
	}

	Logger = slog.New(h)
	slog.SetDefault(Logger)
}

// Debug emits a debug-level message with optional key-value pairs.
func Debug(msg string, args ...any) { Logger.Debug(msg, args...) }

// Info emits an info-level message with optional key-value pairs.
func Info(msg string, args ...any) { Logger.Info(msg, args...) }

// Warn emits a warning-level message with optional key-value pairs.
func Warn(msg string, args ...any) { Logger.Warn(msg, args...) }

// Error emits an error-level message with optional key-value pairs.
func Error(msg string, args ...any) { Logger.Error(msg, args...) }
