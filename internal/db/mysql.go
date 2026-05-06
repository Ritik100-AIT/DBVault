package db

import (
	"bytes"
	"io"
)

type MySQLConnector struct{}

func (m *MySQLConnector) TestConnection() error {
	return nil
}

func (m *MySQLConnector) Backup() (io.Reader, error) {
	return bytes.NewReader([]byte("-- mysql backup data --\n")), nil
}

func (m *MySQLConnector) Restore(src io.Reader) error {
	_, _ = io.ReadAll(src)
	return nil
}
