package storage

import (
	"io"
	"os"
	"path/filepath"
)

type LocalStorage struct {
	BasePath string
}

func (l *LocalStorage) Save(name string, src io.Reader) (string, error) {
	path := filepath.Join(l.BasePath, name)
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	_, err = io.Copy(f, src)
	return path, err
}

func (l *LocalStorage) Load(path string) (io.ReadCloser, error) {
	return os.Open(path)
}
