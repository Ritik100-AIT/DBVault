package db

import "io"

type PostgresConnector struct{}

func (p *PostgresConnector) TestConnection() error {
	return nil
}

func (p *PostgresConnector) Backup() (io.Reader, error) {
	return nil, nil
}

func (p *PostgresConnector) Restore(src io.Reader) error {
	return nil
}
