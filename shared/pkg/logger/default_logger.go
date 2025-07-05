package logger

import (
	"log/slog"
	"os"
)

func NewDefaultLogger() *slog.Logger {
	jsonLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	return jsonLogger
}
