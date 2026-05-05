package db

import "io"

type DBConnector interface {
	TestConnection() error
	Backup() (io.Reader, error)
	Restore(src io.Reader) error
}

func NewConnector(dbType string) DBConnector {
	switch dbType {
	case "mysql":
		return &MySQLConnector{}
	case "postgres":
		return &PostgresConnector{}
	case "mongodb":
		return &MongoDBConnector{}
	case "sqlite":
		return &SQLiteConnector{}
	default:
		return nil
	}
}
