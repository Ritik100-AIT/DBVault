package storage

import (
	"errors"
	"io"
)

// S3Storage is a placeholder S3 implementation for future cloud storage support.
type S3Storage struct {
	Bucket string
	Region string
	Prefix string
}

// Save is not yet implemented but validates S3 configuration.
func (s *S3Storage) Save(name string, src io.Reader) (string, error) {
	if s.Bucket == "" {
		return "", errors.New("s3 bucket is not configured")
	}
	return "", errors.New("s3 storage is not yet implemented")
}

// Load returns an error until S3 backend support is implemented.
func (s *S3Storage) Load(path string) (io.ReadCloser, error) {
	return nil, errors.New("s3 storage is not yet implemented")
}
