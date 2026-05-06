package cmd

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestBackupCommandRunsWithLocalStorage(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	t.Setenv("DBVAULT_STORAGE_LOCAL_PATH", t.TempDir())

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

	defer func() {
		backupCmd.Flags().Set("db", "")
		backupCmd.Flags().Set("storage", "local")
		backupCmd.Flags().Set("compress", "gzip")
		backupCmd.Flags().Set("type", "full")
	}()

	RootCmd.SetArgs([]string{"backup", "--db", "mysql", "--storage", "local", "--type", "full", "--compress", "gzip"})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("expected backup command to succeed, got %v", err)
	}

	w.Close()
	var output strings.Builder
	if _, err := io.Copy(&output, r); err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	if !strings.Contains(output.String(), "Backup completed successfully") {
		t.Fatalf("expected success output, got: %s", output.String())
	}
}
