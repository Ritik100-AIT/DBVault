package storage

import (
	"fmt"
	"io"

	"github.com/dbvault/dbvault/internal/models"
)

// StorageBackend defines methods used by DBVault to save and load backup data.
type StorageBackend interface {
	Save(name string, src io.Reader) (string, error)
	Load(path string) (io.ReadCloser, error)
}

// NewStorageBackend creates a storage backend instance from the config.
func NewStorageBackend(cfg *models.AppConfig) (StorageBackend, error) {
	switch cfg.Storage.Type {
	case "local", "":
		return &LocalStorage{BasePath: cfg.Storage.Local.Path}, nil
	case "s3":
		if cfg.Storage.S3.Bucket == "" {
			return nil, fmt.Errorf("s3 bucket is required")
		}
		return &S3Storage{
			Bucket: cfg.Storage.S3.Bucket,
			Region: cfg.Storage.S3.Region,
			Prefix: cfg.Storage.S3.Prefix,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", cfg.Storage.Type)
	}
}
