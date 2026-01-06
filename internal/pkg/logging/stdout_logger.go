package logging

import (
	"log/slog"
	"os"
)

type Logger interface {
	Info(message string, args ...any)
	Warn(message string, args ...any)
	Error(message string, args ...any)
}

var StdoutLogger = slog.New(slog.NewTextHandler(os.Stdout, nil))
