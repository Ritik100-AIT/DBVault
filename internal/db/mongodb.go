package db

import (
	"bytes"
	"io"
)

type MongoDBConnector struct{}

func (m *MongoDBConnector) TestConnection() error {
	return nil
}

func (m *MongoDBConnector) Backup() (io.Reader, error) {
	return bytes.NewReader([]byte("-- mongodb backup data --\n")), nil
}

func (m *MongoDBConnector) Restore(src io.Reader) error {
	_, _ = io.ReadAll(src)
	return nil
}
