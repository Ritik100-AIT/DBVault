package storage

import (
	"io"
	"os"
	"path/filepath"
)

// LocalStorage implements the StorageBackend interface on the local filesystem.
type LocalStorage struct {
	BasePath string
}

// Save writes the stream to a local file under BasePath.
func (l *LocalStorage) Save(name string, src io.Reader) (string, error) {
	if err := os.MkdirAll(l.BasePath, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(l.BasePath, name)
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(f, src); err != nil {
		return "", err
	}
	return path, nil
}

// Load opens a locally stored backup file for reading.
func (l *LocalStorage) Load(path string) (io.ReadCloser, error) {
	return os.Open(path)
}
