package db

import "io"

// DBConnector defines operations for a database connector.
type DBConnector interface {
	TestConnection() error
	Backup() (io.Reader, error)
	Restore(src io.Reader) error
}

// NewConnector returns a DBConnector implementation for the given database type.
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
