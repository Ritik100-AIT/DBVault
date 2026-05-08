package db

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/dbvault/dbvault/internal/models"
)

type DBConnector interface {
	TestConnection() error
	Backup() (io.Reader, error)
	Restore(src io.Reader) error
}

// NewConnector returns a DBConnector implementation for the given database config.
func NewConnector(cfg *models.DBConfig) DBConnector {
	switch cfg.Type {
	case models.MySQL:
		return &MySQLConnector{Config: cfg}
	case models.Postgres:
		return &PostgresConnector{Config: cfg}
	case models.MongoDB:
		return &MongoDBConnector{Config: cfg}
	case models.SQLite:
		return &SQLiteConnector{Config: cfg}
	default:
		return nil
	}
}

func execCommandReader(cmd *exec.Cmd) (io.ReadCloser, error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		if _, copyErr := io.Copy(pw, stdout); copyErr != nil {
			pw.CloseWithError(copyErr)
			return
		}
		if waitErr := cmd.Wait(); waitErr != nil {
			pw.CloseWithError(fmt.Errorf("%w: %s", waitErr, strings.TrimSpace(stderr.String())))
		}
	}()

	return pr, nil
}

func execCommandWithInput(cmd *exec.Cmd, input io.Reader) error {
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return err
	}

	_, copyErr := io.Copy(stdin, input)
	_ = stdin.Close()
	if copyErr != nil {
		cmd.Wait()
		return copyErr
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
	}

	return nil
}

func environmentWithPassword(env []string, key, password string) []string {
	if password == "" {
		return env
	}
	return append(env, fmt.Sprintf("%s=%s", key, password))
}
