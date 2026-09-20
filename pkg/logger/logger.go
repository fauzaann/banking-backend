package logger

import (
	"log/slog"
	"os"
)

var logger *slog.Logger

// Init menyiapkan logger global: text di development, JSON di production.
func Init(env string) {
	level := slog.LevelDebug
	if env == "production" {
		level = slog.LevelInfo
	}
	opts := &slog.HandlerOptions{Level: level}

	if env == "production" {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, opts))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stdout, opts))
	}
	slog.SetDefault(logger)
}

// L mengembalikan logger global, dan menginisialisasinya jika belum ada.
func L() *slog.Logger {
	if logger == nil {
		Init("development")
	}
	return logger
}

// Info, Warn, Error, Debug adalah fungsi utilitas untuk logging dengan level yang sesuai.
func Info(msg string, args ...any)  { L().Info(msg, args...) }
func Warn(msg string, args ...any)  { L().Warn(msg, args...) }
func Error(msg string, args ...any) { L().Error(msg, args...) }
func Debug(msg string, args ...any) { L().Debug(msg, args...) }
