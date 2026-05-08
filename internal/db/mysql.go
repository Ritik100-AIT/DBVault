package db

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"

	"github.com/dbvault/dbvault/internal/models"
)

type MySQLConnector struct {
	Config *models.DBConfig
}

func (m *MySQLConnector) TestConnection() error {
	if m.Config == nil {
		return fmt.Errorf("mysql configuration is missing")
	}
	args := []string{"-e", "SELECT 1"}
	if m.Config.User != "" {
		args = append([]string{"-u", m.Config.User}, args...)
	}
	if m.Config.Host != "" {
		args = append([]string{"-h", m.Config.Host}, args...)
	}
	if m.Config.Port != 0 {
		args = append([]string{"-P", strconv.Itoa(m.Config.Port)}, args...)
	}

	cmd := exec.Command("mysql", args...)
	cmd.Env = environmentWithPassword(os.Environ(), "MYSQL_PWD", m.Config.Password)
	return cmd.Run()
}

func (m *MySQLConnector) Backup() (io.Reader, error) {
	if m.Config == nil {
		return nil, fmt.Errorf("mysql configuration is missing")
	}
	args := []string{"--single-transaction", "--skip-lock-tables"}
	if m.Config.User != "" {
		args = append(args, "-u", m.Config.User)
	}
	if m.Config.Host != "" {
		args = append(args, "-h", m.Config.Host)
	}
	if m.Config.Port != 0 {
		args = append(args, "-P", strconv.Itoa(m.Config.Port))
	}
	if m.Config.Name != "" {
		args = append(args, m.Config.Name)
	}

	cmd := exec.Command("mysqldump", args...)
	cmd.Env = environmentWithPassword(os.Environ(), "MYSQL_PWD", m.Config.Password)
	return execCommandReader(cmd)
}

func (m *MySQLConnector) Restore(src io.Reader) error {
	if m.Config == nil {
		return fmt.Errorf("mysql configuration is missing")
	}
	args := []string{}
	if m.Config.User != "" {
		args = append(args, "-u", m.Config.User)
	}
	if m.Config.Host != "" {
		args = append(args, "-h", m.Config.Host)
	}
	if m.Config.Port != 0 {
		args = append(args, "-P", strconv.Itoa(m.Config.Port))
	}
	if m.Config.Name != "" {
		args = append(args, m.Config.Name)
	}

	cmd := exec.Command("mysql", args...)
	cmd.Env = environmentWithPassword(os.Environ(), "MYSQL_PWD", m.Config.Password)
	return execCommandWithInput(cmd, src)
}
