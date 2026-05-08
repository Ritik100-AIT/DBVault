package db

import (
	"fmt"
	"io"
	"net/url"
	"os/exec"

	"github.com/dbvault/dbvault/internal/models"
)

type MongoDBConnector struct {
	Config *models.DBConfig
}

func (m *MongoDBConnector) TestConnection() error {
	if m.Config == nil {
		return fmt.Errorf("mongodb configuration is missing")
	}

	shell := detectMongoShell()
	args := []string{"--quiet", "--eval", "db.runCommand({ping:1})"}
	if m.Config.Host != "" {
		args = append([]string{"--host", m.Config.Host}, args...)
	}
	if m.Config.Port != 0 {
		args = append([]string{"--port", fmt.Sprint(m.Config.Port)}, args...)
	}
	if m.Config.User != "" {
		args = append([]string{"--username", m.Config.User}, args...)
	}
	if m.Config.Password != "" {
		args = append([]string{"--password", m.Config.Password}, args...)
	}

	cmd := exec.Command(shell, args...)
	return cmd.Run()
}

func (m *MongoDBConnector) Backup() (io.Reader, error) {
	if m.Config == nil {
		return nil, fmt.Errorf("mongodb configuration is missing")
	}

	args := []string{"--archive"}
	uri := buildMongoURI(m.Config)
	if uri != "" {
		args = append([]string{"--uri", uri}, args...)
	}
	if m.Config.Name != "" {
		args = append(args, "--db", m.Config.Name)
	}

	cmd := exec.Command("mongodump", args...)
	return execCommandReader(cmd)
}

func (m *MongoDBConnector) Restore(src io.Reader) error {
	if m.Config == nil {
		return fmt.Errorf("mongodb configuration is missing")
	}

	args := []string{"--archive"}
	uri := buildMongoURI(m.Config)
	if uri != "" {
		args = append([]string{"--uri", uri}, args...)
	}
	if m.Config.Name != "" {
		args = append(args, "--db", m.Config.Name)
	}

	cmd := exec.Command("mongorestore", args...)
	return execCommandWithInput(cmd, src)
}

func detectMongoShell() string {
	if _, err := exec.LookPath("mongo"); err == nil {
		return "mongo"
	}
	return "mongosh"
}

func buildMongoURI(cfg *models.DBConfig) string {
	if cfg.Host == "" {
		return ""
	}

	hostPort := cfg.Host
	if cfg.Port != 0 {
		hostPort = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	}

	u := url.URL{
		Scheme: "mongodb",
		Host:   hostPort,
	}
	if cfg.User != "" {
		if cfg.Password != "" {
			u.User = url.UserPassword(cfg.User, cfg.Password)
		} else {
			u.User = url.User(cfg.User)
		}
	}
	if cfg.Name != "" {
		u.Path = "/" + cfg.Name
	}
	return u.String()
}
