package db

import "io"

type MySQLConnector struct{}

func (m *MySQLConnector) TestConnection() error {
	return nil
}

func (m *MySQLConnector) Backup() (io.Reader, error) {
	return nil, nil
}

func (m *MySQLConnector) Restore(src io.Reader) error {
	return nil
}
