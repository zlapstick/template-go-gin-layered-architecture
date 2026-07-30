package helper

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

func Logger(module string) *slog.Logger {
	dir := "logs"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		slog.Error("gagal membuat folder logs, fallback ke stdout", "error", err)
		return slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}

	path := filepath.Join(dir, module+".log")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		slog.Error("gagal membuka file log, fallback ke stdout", "module", module, "error", err)
		return slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}

	writer := io.MultiWriter(f, os.Stdout)

	handler := slog.NewJSONHandler(writer, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	return slog.New(handler).With("module", module)
}
