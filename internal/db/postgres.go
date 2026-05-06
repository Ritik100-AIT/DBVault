package db

import (
	"bytes"
	"io"
)

type PostgresConnector struct{}

func (p *PostgresConnector) TestConnection() error {
	return nil
}

func (p *PostgresConnector) Backup() (io.Reader, error) {
	return bytes.NewReader([]byte("-- postgres backup data --\n")), nil
}

func (p *PostgresConnector) Restore(src io.Reader) error {
	_, _ = io.ReadAll(src)
	return nil
}
