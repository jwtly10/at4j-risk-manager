package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"
)

// Global logger instance
var logger *slog.Logger

type customHandler struct {
	slog.Handler
	l *slog.Logger
}

func newCustomHandler(opts *slog.HandlerOptions) slog.Handler {
	return &customHandler{
		Handler: slog.NewTextHandler(os.Stdout, opts),
	}
}

func (h *customHandler) Handle(_ context.Context, r slog.Record) error {
	timeStr := r.Time.Format("2006-01-02 15:04:05.000")
	level := strings.ToUpper(r.Level.String())
	caller := ""
	msg := r.Message

	// Extract and remove caller from attrs if present
	attrs := make([]slog.Attr, 0, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "caller" {
			caller = a.Value.String()
		} else {
			attrs = append(attrs, a)
		}
		return true
	})

	fmt.Printf("%s [%s] %s - %s", timeStr, level, caller, msg)

	if len(attrs) > 0 {
		fmt.Print(" - ")
		for i, attr := range attrs {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Printf("%s=%v", attr.Key, attr.Value)
		}
	}
	fmt.Println()

	return nil
}

func getCallerInfo() string {
	_, file, line, ok := runtime.Caller(2) // Skip 2 frames to get the actual caller
	if !ok {
		return "unknown:0"
	}
	parts := strings.Split(file, "/")
	file = parts[len(parts)-1]
	return fmt.Sprintf("%s:%d", file, line)
}

func Debugf(format string, args ...any) {
	logger.Debug(fmt.Sprintf(format, args...), "caller", getCallerInfo())
}

func Infof(format string, args ...any) {
	logger.Info(fmt.Sprintf(format, args...), "caller", getCallerInfo())
}

func Warnf(format string, args ...any) {
	logger.Warn(fmt.Sprintf(format, args...), "caller", getCallerInfo())
}

func Errorf(format string, args ...any) {
	logger.Error(fmt.Sprintf(format, args...), "caller", getCallerInfo())
}

func Fatalf(format string, args ...any) {
	logger.Error(fmt.Sprintf(format, args...), "caller", getCallerInfo())
	os.Exit(1)
}

func InitLogger() {
	level := getLogLevel()

	opts := &slog.HandlerOptions{
		Level: level,
	}

	handler := newCustomHandler(opts)
	logger = slog.New(handler)

	env := os.Getenv("Env")
	if env == "" {
		env = "development"
	}

	logger.Info("logger initialized",
		"level", level.String(),
		"environment", env,
	)
}

func getLogLevel() slog.Level {
	levelStr := strings.ToUpper(os.Getenv("LOG_LEVEL"))
	if levelStr == "" {
		levelStr = "INFO"
	}

	switch levelStr {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
