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

	if path != "" {
		v.SetConfigFile(path)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		v.AddConfigPath(home)
		v.SetConfigName(".dbvault/config")
	}

	_ = v.ReadInConfig()
	var cfg models.AppConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
