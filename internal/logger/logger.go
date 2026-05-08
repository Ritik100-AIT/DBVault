package logger

import (
	"io"
	"log/slog"
	"os"

	"github.com/dbvault/dbvault/internal/models"
)

type Logger struct {
	logger *slog.Logger
}

func NewLogger(cfg *models.LoggingConfig) (*Logger, error) {
	var handle io.Writer = os.Stdout
	if cfg != nil && cfg.File != "" {
		f, err := os.OpenFile(cfg.File, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return nil, err
		}
		handle = f
	}

	options := &slog.HandlerOptions{}
	if cfg != nil && cfg.Format == "json" {
		handler := slog.NewJSONHandler(handle, options)
		return &Logger{logger: slog.New(handler)}, nil
	}

	handler := slog.NewTextHandler(handle, options)
	return &Logger{logger: slog.New(handler)}, nil
}

func (l *Logger) Info(msg string) {
	if l == nil || l.logger == nil {
		return
	}
	l.logger.Info(msg)
}

func (l *Logger) Error(msg string) {
	if l == nil || l.logger == nil {
		return
	}
	l.logger.Error(msg)
}

func (l *Logger) Debug(msg string) {
	if l == nil || l.logger == nil {
		return
	}
	l.logger.Debug(msg)
}
