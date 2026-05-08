package db

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/dbvault/dbvault/internal/models"
)

type SQLiteConnector struct {
	Config *models.DBConfig
}

func (s *SQLiteConnector) databasePath() string {
	if s.Config == nil {
		return ""
	}
	if s.Config.Name != "" {
		return s.Config.Name
	}
	return s.Config.Host
}

func (s *SQLiteConnector) TestConnection() error {
	path := s.databasePath()
	if path == "" {
		return fmt.Errorf("sqlite database path is required")
	}
	cmd := exec.Command("sqlite3", path, "SELECT 1;")
	return cmd.Run()
}

func (s *SQLiteConnector) Backup() (io.Reader, error) {
	path := s.databasePath()
	if path == "" {
		return nil, fmt.Errorf("sqlite database path is required")
	}
	cmd := exec.Command("sqlite3", path, ".dump")
	return execCommandReader(cmd)
}

func (s *SQLiteConnector) Restore(src io.Reader) error {
	path := s.databasePath()
	if path == "" {
		return fmt.Errorf("sqlite database path is required")
	}
	_ = os.Remove(path)
	cmd := exec.Command("sqlite3", path)
	return execCommandWithInput(cmd, src)
}
