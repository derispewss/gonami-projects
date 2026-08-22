package storage

import (
	"context"
	"io"
	"log/slog"

	"github.com/derispewss/gonami-projects/internal/config"
)

type Storage interface {
	Save(ctx context.Context, key string, r io.Reader, contentType string) error
}

func New(cfg *config.Config) Storage {
	switch cfg.StorageDriver {
	case "minio":
		m, err := newMinio(cfg)
		if err != nil {
			slog.Warn("fallback ke local storage", "error", err)
			return newLocal(cfg.StorageLocalDir)
		}
		return m
	default:
		return newLocal(cfg.StorageLocalDir)
	}
}
