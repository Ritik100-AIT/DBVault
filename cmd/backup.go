package cmd

import (
	"fmt"

	"github.com/dbvault/dbvault/internal/backup"
	"github.com/dbvault/dbvault/internal/config"
	"github.com/dbvault/dbvault/internal/db"
	"github.com/dbvault/dbvault/internal/logger"
	"github.com/dbvault/dbvault/internal/models"
	"github.com/dbvault/dbvault/internal/notify"
	"github.com/dbvault/dbvault/internal/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var newDBConnector = func(cfg *models.DBConfig) db.DBConnector {
	return db.NewConnector(cfg)
}

// backupCmd runs a backup using merged config values from file, env, and flags.
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create a backup for a supported database",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if dbType, _ := cmd.Flags().GetString("db"); dbType != "" {
			cfg.Database.Type = models.DBType(dbType)
		}
		if host, _ := cmd.Flags().GetString("host"); host != "" {
			cfg.Database.Host = host
		}
		if user, _ := cmd.Flags().GetString("user"); user != "" {
			cfg.Database.User = user
		}
		if password, _ := cmd.Flags().GetString("password"); password != "" {
			cfg.Database.Password = password
		}
		if name, _ := cmd.Flags().GetString("name"); name != "" {
			cfg.Database.Name = name
		}
		if backupType, _ := cmd.Flags().GetString("type"); backupType != "" {
			cfg.Backup.Type = models.BackupType(backupType)
		}
		if storageType, _ := cmd.Flags().GetString("storage"); storageType != "" {
			cfg.Storage.Type = storageType
		}
		if compress, _ := cmd.Flags().GetString("compress"); compress != "" {
			cfg.Backup.Compression = compress
		}

		if err := config.ValidateConfig(cfg); err != nil {
			return fmt.Errorf("invalid configuration: %w", err)
		}

		connector := newDBConnector(&cfg.Database)
		if connector == nil {
			return fmt.Errorf("unsupported database type: %s", cfg.Database.Type)
		}

		backend, err := storage.NewStorageBackend(cfg)
		if err != nil {
			return fmt.Errorf("failed to initialize storage backend: %w", err)
		}

		var notifier *notify.SlackNotifier
		if cfg.Notifications.Slack.Enabled {
			notifier = notify.NewSlackNotifier(cfg.Notifications.Slack.WebhookURL)
		}

		loggerInstance, err := logger.NewLogger(&cfg.Logging)
		if err != nil {
			return fmt.Errorf("failed to initialize logger: %w", err)
		}

		manager := backup.NewBackupManager(connector, backend, cfg, notifier, loggerInstance)
		if err := manager.Run(); err != nil {
			return fmt.Errorf("backup failed: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
	backupCmd.Flags().String("db", "", "Database type: mysql | postgres | mongodb | sqlite")
	backupCmd.Flags().String("host", "", "Database host")
	backupCmd.Flags().String("user", "", "Database user")
	backupCmd.Flags().String("password", "", "Database password")
	backupCmd.Flags().String("name", "", "Database name")
	backupCmd.Flags().String("type", "full", "Backup type: full | incremental | differential")
	backupCmd.Flags().String("storage", "local", "Storage backend: local | s3")
	backupCmd.Flags().String("compress", "gzip", "Compression method: gzip | none")
	backupCmd.Flags().String("local-path", "", "Local backup storage path")
	backupCmd.Flags().String("s3-bucket", "", "S3 bucket name")
	backupCmd.Flags().String("s3-region", "", "S3 region")
	backupCmd.Flags().String("s3-access-key", "", "S3 access key")
	backupCmd.Flags().String("s3-secret-key", "", "S3 secret key")
	backupCmd.Flags().String("s3-endpoint", "", "S3 or MinIO endpoint URL")
	backupCmd.Flags().Bool("s3-force-path-style", false, "Use path style for S3 endpoints")
	backupCmd.Flags().String("s3-prefix", "", "S3 object key prefix")
	backupCmd.Flags().Bool("notify", false, "Send Slack notification on backup completion")

	viper.BindPFlag("database.type", backupCmd.Flags().Lookup("db"))
	viper.BindPFlag("database.host", backupCmd.Flags().Lookup("host"))
	viper.BindPFlag("database.username", backupCmd.Flags().Lookup("user"))
	viper.BindPFlag("database.password", backupCmd.Flags().Lookup("password"))
	viper.BindPFlag("database.name", backupCmd.Flags().Lookup("name"))
	viper.BindPFlag("backup.type", backupCmd.Flags().Lookup("type"))
	viper.BindPFlag("storage.type", backupCmd.Flags().Lookup("storage"))
	viper.BindPFlag("storage.local.path", backupCmd.Flags().Lookup("local-path"))
	viper.BindPFlag("storage.s3.bucket", backupCmd.Flags().Lookup("s3-bucket"))
	viper.BindPFlag("storage.s3.region", backupCmd.Flags().Lookup("s3-region"))
	viper.BindPFlag("storage.s3.access_key", backupCmd.Flags().Lookup("s3-access-key"))
	viper.BindPFlag("storage.s3.secret_key", backupCmd.Flags().Lookup("s3-secret-key"))
	viper.BindPFlag("storage.s3.endpoint", backupCmd.Flags().Lookup("s3-endpoint"))
	viper.BindPFlag("storage.s3.force_path_style", backupCmd.Flags().Lookup("s3-force-path-style"))
	viper.BindPFlag("storage.s3.prefix", backupCmd.Flags().Lookup("s3-prefix"))
	viper.BindPFlag("backup.compression", backupCmd.Flags().Lookup("compress"))
	viper.BindPFlag("notifications.slack.enabled", backupCmd.Flags().Lookup("notify"))
}
