package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestSetupViperLoadsEnvAndDefaults(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	t.Setenv("DBVAULT_DB_PASSWORD", "secret")

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")
	configData := `database:
  type: mysql
  host: localhost
  username: root
  name: prod_db
backup:
  type: incremental
  compression: gzip
storage:
  type: local
  local:
    path: ./backups
`
	if err := os.WriteFile(cfgPath, []byte(configData), 0o600); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	if err := SetupViper(cfgPath); err != nil {
		t.Fatalf("SetupViper returned error: %v", err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	if cfg.Database.Password != "secret" {
		t.Fatalf("expected password from env, got %q", cfg.Database.Password)
	}

	if cfg.Backup.Type != "incremental" {
		t.Fatalf("expected backup.type incremental, got %q", cfg.Backup.Type)
	}

	if cfg.Storage.Type != "local" {
		t.Fatalf("expected storage.type local, got %q", cfg.Storage.Type)
	}

	if cfg.Logging.Level != "info" {
		t.Fatalf("expected default logging.level info, got %q", cfg.Logging.Level)
	}

	if !strings.Contains(viper.ConfigFileUsed(), "config.yaml") {
		t.Fatalf("expected config file to be loaded, got %q", viper.ConfigFileUsed())
	}
}
