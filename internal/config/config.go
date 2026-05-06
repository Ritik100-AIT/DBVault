package config

import (
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

	viper.BindEnv("database.username", "DBVAULT_DB_USERNAME")
	viper.BindEnv("database.password", "DBVAULT_DB_PASSWORD")
	viper.BindEnv("database.host", "DBVAULT_DB_HOST")
	viper.BindEnv("database.port", "DBVAULT_DB_PORT")
	viper.BindEnv("database.name", "DBVAULT_DB_NAME")
	viper.BindEnv("notifications.slack.webhook_url", "DBVAULT_SLACK_WEBHOOK")
	viper.BindEnv("storage.s3.access_key", "DBVAULT_STORAGE_S3_ACCESS_KEY")
	viper.BindEnv("storage.s3.secret_key", "DBVAULT_STORAGE_S3_SECRET_KEY")
	viper.BindEnv("storage.s3.bucket", "DBVAULT_STORAGE_S3_BUCKET")
	viper.BindEnv("storage.s3.region", "DBVAULT_STORAGE_S3_REGION")
	viper.BindEnv("storage.s3.prefix", "DBVAULT_STORAGE_S3_PREFIX")
	viper.BindEnv("storage.type", "DBVAULT_STORAGE_TYPE")

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
