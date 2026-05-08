package db

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"

	"github.com/dbvault/dbvault/internal/models"
)

type PostgresConnector struct {
	Config *models.DBConfig
}

func (p *PostgresConnector) TestConnection() error {
	if p.Config == nil {
		return fmt.Errorf("postgres configuration is missing")
	}
	args := []string{"-c", "SELECT 1"}
	if p.Config.User != "" {
		args = append([]string{"-U", p.Config.User}, args...)
	}
	if p.Config.Host != "" {
		args = append([]string{"-h", p.Config.Host}, args...)
	}
	if p.Config.Port != 0 {
		args = append([]string{"-p", strconv.Itoa(p.Config.Port)}, args...)
	}
	if p.Config.Name != "" {
		args = append(args, p.Config.Name)
	}

	cmd := exec.Command("psql", args...)
	cmd.Env = environmentWithPassword(os.Environ(), "PGPASSWORD", p.Config.Password)
	return cmd.Run()
}

func (p *PostgresConnector) Backup() (io.Reader, error) {
	if p.Config == nil {
		return nil, fmt.Errorf("postgres configuration is missing")
	}
	args := []string{"--format=plain"}
	if p.Config.User != "" {
		args = append(args, "-U", p.Config.User)
	}
	if p.Config.Host != "" {
		args = append(args, "-h", p.Config.Host)
	}
	if p.Config.Port != 0 {
		args = append(args, "-p", strconv.Itoa(p.Config.Port))
	}
	if p.Config.Name != "" {
		args = append(args, p.Config.Name)
	}

	cmd := exec.Command("pg_dump", args...)
	cmd.Env = environmentWithPassword(os.Environ(), "PGPASSWORD", p.Config.Password)
	return execCommandReader(cmd)
}

func (p *PostgresConnector) Restore(src io.Reader) error {
	if p.Config == nil {
		return fmt.Errorf("postgres configuration is missing")
	}
	args := []string{}
	if p.Config.User != "" {
		args = append(args, "-U", p.Config.User)
	}
	if p.Config.Host != "" {
		args = append(args, "-h", p.Config.Host)
	}
	if p.Config.Port != 0 {
		args = append(args, "-p", strconv.Itoa(p.Config.Port))
	}
	if p.Config.Name != "" {
		args = append(args, p.Config.Name)
	}

	cmd := exec.Command("psql", args...)
	cmd.Env = environmentWithPassword(os.Environ(), "PGPASSWORD", p.Config.Password)
	return execCommandWithInput(cmd, src)
}
