package storage

import "io"

type S3Storage struct {
	Bucket string
	Region string
	Prefix string
}

func (s *S3Storage) Save(name string, src io.Reader) (string, error) {
	return "", nil
}

func (s *S3Storage) Load(path string) (io.ReadCloser, error) {
	return nil, nil
}
