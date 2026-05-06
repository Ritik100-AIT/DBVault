package cmd

import (
	"fmt"

	"github.com/dbvault/dbvault/internal/backup"
	"github.com/dbvault/dbvault/internal/config"
	"github.com/dbvault/dbvault/internal/db"
	"github.com/dbvault/dbvault/internal/models"
	"github.com/dbvault/dbvault/internal/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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

		connector := db.NewConnector(string(cfg.Database.Type))
		if connector == nil {
			return fmt.Errorf("unsupported database type: %s", cfg.Database.Type)
		}

		backend, err := storage.NewStorageBackend(cfg)
		if err != nil {
			return fmt.Errorf("failed to initialize storage backend: %w", err)
		}

		// Create the backup manager that orchestrates DB extraction and storage.
		manager := backup.NewBackupManager(connector, backend, cfg)
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

	viper.BindPFlag("database.type", backupCmd.Flags().Lookup("db"))
	viper.BindPFlag("database.host", backupCmd.Flags().Lookup("host"))
	viper.BindPFlag("database.username", backupCmd.Flags().Lookup("user"))
	viper.BindPFlag("database.password", backupCmd.Flags().Lookup("password"))
	viper.BindPFlag("database.name", backupCmd.Flags().Lookup("name"))
	viper.BindPFlag("backup.type", backupCmd.Flags().Lookup("type"))
	viper.BindPFlag("storage.type", backupCmd.Flags().Lookup("storage"))
	viper.BindPFlag("backup.compression", backupCmd.Flags().Lookup("compress"))
}
