package config

import (
	"os"

	"github.com/dbvault/dbvault/internal/models"
	"github.com/spf13/viper"
)

func LoadConfig(path string) (*models.AppConfig, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetDefault("backup.type", "full")
	v.SetDefault("backup.compression", "gzip")
	v.SetDefault("storage.type", "local")
	v.SetDefault("logging.level", "info")

	// Bind environment variables
	v.BindEnv("database.password", "DBVAULT_DB_PASSWORD")
	v.BindEnv("notifications.slack.webhook_url", "DBVAULT_SLACK_WEBHOOK")
	v.BindEnv("storage.s3.access_key", "AWS_ACCESS_KEY_ID")
	v.BindEnv("storage.s3.secret_key", "AWS_SECRET_ACCESS_KEY")

	if path != "" {
		v.SetConfigFile(path)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		v.AddConfigPath(home + "/.dbvault")
		v.SetConfigName("config")
	}

	_ = v.ReadInConfig() // Ignore error if file doesn't exist
	var cfg models.AppConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
