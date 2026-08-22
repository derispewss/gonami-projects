package storage

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/derispewss/finwa-projects/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type minioStorage struct {
	client *minio.Client
	bucket string
}

func newMinio(cfg *config.Config) (*minioStorage, error) {
	cli, err := minio.New(cfg.StorageEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.StorageAccessKey, cfg.StorageSecretKey, ""),
		Secure: cfg.StorageUseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("gagal inisialisasi minio client: %w", err)
	}
	m := &minioStorage{client: cli, bucket: cfg.StorageBucket}

	ctx := context.Background()
	exists, err := cli.BucketExists(ctx, m.bucket)
	if err != nil {
		slog.Warn("minio tidak dapat dijangkau saat startup", "endpoint", cfg.StorageEndpoint, "error", err)
	} else if !exists {
		if err := cli.MakeBucket(ctx, m.bucket, minio.MakeBucketOptions{}); err != nil {
			slog.Warn("gagal membuat bucket minio", "bucket", m.bucket, "error", err)
		} else {
			slog.Info("bucket minio dibuat", "bucket", m.bucket)
		}
	}
	return m, nil
}

func (m *minioStorage) Save(ctx context.Context, key string, r io.Reader, contentType string) error {
	_, err := m.client.PutObject(ctx, m.bucket, key, r, -1,
		minio.PutObjectOptions{ContentType: contentType})
	return err
}
