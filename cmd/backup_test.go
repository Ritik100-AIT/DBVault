package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/dbvault/dbvault/internal/db"
	"github.com/dbvault/dbvault/internal/models"
	"github.com/spf13/viper"
)

type testDBConnector struct{}

func (t *testDBConnector) TestConnection() error {
	return nil
}

func (t *testDBConnector) Backup() (io.Reader, error) {
	return bytes.NewReader([]byte("-- backup data --\n")), nil
}

func (t *testDBConnector) Restore(src io.Reader) error {
	_, _ = io.ReadAll(src)
	return nil
}

func TestBackupCommandRunsWithLocalStorage(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	oldConnector := newDBConnector
	newDBConnector = func(cfg *models.DBConfig) db.DBConnector {
		return &testDBConnector{}
	}
	defer func() { newDBConnector = oldConnector }()

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
