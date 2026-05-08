package models

import "time"

type DBType string

type BackupType string

const (
	MySQL    DBType = "mysql"
	Postgres DBType = "postgres"
	MongoDB  DBType = "mongodb"
	SQLite   DBType = "sqlite"

	BackupFull         BackupType = "full"
	BackupIncremental  BackupType = "incremental"
	BackupDifferential BackupType = "differential"
)

type DBConfig struct {
	Type     DBType `yaml:"type"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"username"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
}

type StorageConfig struct {
	Type  string       `yaml:"type"`
	Local LocalStorage `yaml:"local"`
	S3    S3Storage    `yaml:"s3"`
}

type LocalStorage struct {
	Path string `yaml:"path"`
}

type S3Storage struct {
	Bucket         string `yaml:"bucket"`
	Region         string `yaml:"region"`
	AccessKey      string `yaml:"access_key"`
	SecretKey      string `yaml:"secret_key"`
	Endpoint       string `yaml:"endpoint"`
	ForcePathStyle bool   `yaml:"force_path_style"`
	Prefix         string `yaml:"prefix"`
}

type NotificationConfig struct {
	Slack SlackConfig `yaml:"slack"`
}

type SlackConfig struct {
	Enabled    bool   `yaml:"enabled"`
	WebhookURL string `yaml:"webhook_url"`
}

type LoggingConfig struct {
	Level      string `yaml:"level"`
	Format     string `yaml:"format"`
	File       string `yaml:"file"`
	MaxSizeMB  int    `yaml:"max_size_mb"`
	MaxBackups int    `yaml:"max_backups"`
}

type AppConfig struct {
	Database DBConfig `yaml:"database"`
	Backup   struct {
		Type        BackupType `yaml:"type"`
		Compression string     `yaml:"compression"`
		OutputDir   string     `yaml:"output_dir"`
	} `yaml:"backup"`
	Storage       StorageConfig      `yaml:"storage"`
	Notifications NotificationConfig `yaml:"notifications"`
	Logging       LoggingConfig      `yaml:"logging"`
}

type BackupRecord struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Database    string    `json:"database"`
	Path        string    `json:"path"`
	Checksum    string    `json:"checksum"`
	SizeBytes   int64     `json:"size_bytes"`
	Compression string    `json:"compression,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
}
