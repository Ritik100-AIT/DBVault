package db

import "io"

type MongoDBConnector struct{}

func (m *MongoDBConnector) TestConnection() error {
	return nil
}

func (m *MongoDBConnector) Backup() (io.Reader, error) {
	return nil, nil
}

func (m *MongoDBConnector) Restore(src io.Reader) error {
	return nil
}
