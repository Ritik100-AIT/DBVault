package db

import "io"

type SQLiteConnector struct{}

func (s *SQLiteConnector) TestConnection() error {
	return nil
}

func (s *SQLiteConnector) Backup() (io.Reader, error) {
	return nil, nil
}

func (s *SQLiteConnector) Restore(src io.Reader) error {
	return nil
}
