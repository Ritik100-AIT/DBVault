package db

import (
	"bytes"
	"io"
)

type SQLiteConnector struct{}

func (s *SQLiteConnector) TestConnection() error {
	return nil
}

func (s *SQLiteConnector) Backup() (io.Reader, error) {
	return bytes.NewReader([]byte("-- sqlite backup data --\n")), nil
}

func (s *SQLiteConnector) Restore(src io.Reader) error {
	_, _ = io.ReadAll(src)
	return nil
}
