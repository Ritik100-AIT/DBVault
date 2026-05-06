package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestConfigViewCommand(t *testing.T) {
	t.Setenv("DBVAULT_DB_TYPE", "")
	t.Setenv("DBVAULT_DATABASE_TYPE", "")
	viper.Reset()
	defer viper.Reset()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w
	defer func() {
		w.Close()
		os.Stdout = oldStdout
	}()

	RootCmd.SetArgs([]string{"config", "view"})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("expected config view to execute cleanly, got %v", err)
	}

	w.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Current configuration:") {
		t.Fatalf("unexpected config view output: %s", output)
	}
	if !strings.Contains(output, "backup") {
		t.Fatalf("expected config fields in output, got: %s", output)
	}
}

func TestBackupCommandFailsWithoutDatabaseType(t *testing.T) {
	t.Setenv("DBVAULT_DB_TYPE", "")
	t.Setenv("DBVAULT_DATABASE_TYPE", "")
	viper.Reset()
	defer viper.Reset()

	backupCmd.Flags().Set("db", "")
	tmpConfig := filepath.Join(t.TempDir(), "config.yaml")
	RootCmd.SetArgs([]string{"-c", tmpConfig, "backup"})

	err := RootCmd.Execute()
	if err == nil {
		t.Fatal("expected backup command to return an error for missing database type")
	}
	if !strings.Contains(err.Error(), "unsupported database type") {
		t.Fatalf("unexpected error: %v", err)
	}
}
