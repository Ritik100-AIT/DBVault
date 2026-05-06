package storage

import (
	"bytes"
	"io"
	"testing"

	"github.com/dbvault/dbvault/internal/models"
)

func TestNewStorageBackendLocal(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &models.AppConfig{
		Storage: models.StorageConfig{
			Type: "local",
			Local: models.LocalStorage{
				Path: tmpDir,
			},
		},
	}

	backend, err := NewStorageBackend(cfg)
	if err != nil {
		t.Fatalf("expected local backend, got error: %v", err)
	}

	path, err := backend.Save("test.txt", bytes.NewBufferString("hello"))
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}

	file, err := backend.Load(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}

	if string(data) != "hello" {
		t.Fatalf("expected hello, got %q", string(data))
	}
}

func TestNewStorageBackendS3Validation(t *testing.T) {
	cfg := &models.AppConfig{
		Storage: models.StorageConfig{
			Type: "s3",
			S3:   models.S3Storage{},
		},
	}

	_, err := NewStorageBackend(cfg)
	if err == nil {
		t.Fatal("expected error for missing s3 bucket")
	}
}
