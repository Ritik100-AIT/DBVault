package storage

import "io"

type StorageBackend interface {
	Save(name string, src io.Reader) (string, error)
	Load(path string) (io.ReadCloser, error)
}
