package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dbvault/dbvault/internal/models"
	"github.com/spf13/viper"
)

// SetupViper configures Viper to load defaults, environment variables,
// and an optional YAML config file for DBVault.
func SetupViper(configPath string) error {
	viper.SetConfigType("yaml")
	viper.SetEnvPrefix("DBVAULT")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetDefault("backup.type", "full")
	viper.SetDefault("backup.compression", "gzip")
	viper.SetDefault("storage.type", "local")
	viper.SetDefault("storage.local.path", "./backups")
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("notifications.slack.enabled", false)

	viper.BindEnv("database.username", "DBVAULT_DB_USERNAME")
	viper.BindEnv("database.password", "DBVAULT_DB_PASSWORD")
	viper.BindEnv("database.host", "DBVAULT_DB_HOST")
	viper.BindEnv("database.port", "DBVAULT_DB_PORT")
	viper.BindEnv("database.name", "DBVAULT_DB_NAME")
	viper.BindEnv("notifications.slack.webhook_url", "DBVAULT_SLACK_WEBHOOK")
	viper.BindEnv("notifications.slack.enabled", "DBVAULT_SLACK_ENABLED")
	viper.BindEnv("storage.local.path", "DBVAULT_STORAGE_LOCAL_PATH")
	viper.BindEnv("storage.s3.access_key", "DBVAULT_STORAGE_S3_ACCESS_KEY")
	viper.BindEnv("storage.s3.secret_key", "DBVAULT_STORAGE_S3_SECRET_KEY")
	viper.BindEnv("storage.s3.bucket", "DBVAULT_STORAGE_S3_BUCKET")
	viper.BindEnv("storage.s3.region", "DBVAULT_STORAGE_S3_REGION")
	viper.BindEnv("storage.s3.endpoint", "DBVAULT_STORAGE_S3_ENDPOINT")
	viper.BindEnv("storage.s3.force_path_style", "DBVAULT_STORAGE_S3_FORCE_PATH_STYLE")
	viper.BindEnv("storage.s3.prefix", "DBVAULT_STORAGE_S3_PREFIX")
	viper.BindEnv("storage.type", "DBVAULT_STORAGE_TYPE")
	viper.BindEnv("logging.level", "DBVAULT_LOG_LEVEL")
	viper.BindEnv("logging.format", "DBVAULT_LOG_FORMAT")
	viper.BindEnv("logging.file", "DBVAULT_LOG_FILE")
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		viper.AddConfigPath(filepath.Join(home, ".dbvault"))
		viper.SetConfigName("config")
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok || os.IsNotExist(err) {
			return nil
		}
		return err
	}

	return nil
}

// LoadConfig unmarshals the merged Viper configuration into AppConfig.
func LoadConfig() (*models.AppConfig, error) {
	var cfg models.AppConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// ValidateConfig ensures the provided application config is complete and sane.
func ValidateConfig(cfg *models.AppConfig) error {
	if cfg == nil {
		return fmt.Errorf("config is required")
	}

	switch cfg.Database.Type {
	case models.MySQL, models.Postgres, models.MongoDB:
		// Database host/name may be optional if the CLI tool can infer defaults.
	case models.SQLite:
		if cfg.Database.Name == "" && cfg.Database.Host == "" {
			return fmt.Errorf("sqlite database path is required in database.name or database.host")
		}
	default:
		return fmt.Errorf("unsupported database type: %s", cfg.Database.Type)
	}

	if cfg.Storage.Type == "s3" {
		if cfg.Storage.S3.Bucket == "" {
			return fmt.Errorf("storage.s3.bucket is required for s3 storage")
		}
	}

	switch cfg.Backup.Type {
	case models.BackupFull:
		// always supported
	case models.BackupIncremental:
		if cfg.Database.Type != models.MySQL && cfg.Database.Type != models.MongoDB {
			return fmt.Errorf("incremental backups are only supported for mysql and mongodb")
		}
	case models.BackupDifferential:
		if cfg.Database.Type != models.Postgres {
			return fmt.Errorf("differential backups are only supported for postgres")
		}
	default:
		return fmt.Errorf("unsupported backup type: %s", cfg.Backup.Type)
	}

	if cfg.Notifications.Slack.Enabled && cfg.Notifications.Slack.WebhookURL == "" {
		return fmt.Errorf("notifications.slack.webhook_url is required when slack notifications are enabled")
	}

	return nil
}
