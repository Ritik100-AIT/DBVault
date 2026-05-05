package cmd

import (
	"fmt"
	"log"

	"github.com/dbvault/dbvault/internal/backup"
	"github.com/dbvault/dbvault/internal/config"
	"github.com/dbvault/dbvault/internal/db"
	"github.com/dbvault/dbvault/internal/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create a backup for a supported database",
	Run: func(cmd *cobra.Command, args []string) {
		// Load configuration
		cfgPath, _ := cmd.Flags().GetString("config")
		cfg, err := config.LoadConfig(cfgPath)
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		// Override config with CLI flags
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
		if storage, _ := cmd.Flags().GetString("storage"); storage != "" {
			cfg.Storage.Type = storage
		}
		if compress, _ := cmd.Flags().GetString("compress"); compress != "" {
			cfg.Backup.Compression = compress
		}

		// Create DB connector
		connector := db.NewConnector(string(cfg.Database.Type))
		if connector == nil {
			log.Fatalf("Unsupported database type: %s", cfg.Database.Type)
		}

		// Create backup manager
		manager := backup.NewBackupManager(connector, cfg)

		// Run backup
		if err := manager.Run(); err != nil {
			log.Fatalf("Backup failed: %v", err)
		}

		fmt.Println("Backup completed successfully")
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

	// Bind flags to viper for config merging
	viper.BindPFlag("database.type", backupCmd.Flags().Lookup("db"))
	viper.BindPFlag("database.host", backupCmd.Flags().Lookup("host"))
	viper.BindPFlag("database.username", backupCmd.Flags().Lookup("user"))
	viper.BindPFlag("database.password", backupCmd.Flags().Lookup("password"))
	viper.BindPFlag("database.name", backupCmd.Flags().Lookup("name"))
	viper.BindPFlag("backup.type", backupCmd.Flags().Lookup("type"))
	viper.BindPFlag("storage.type", backupCmd.Flags().Lookup("storage"))
	viper.BindPFlag("backup.compression", backupCmd.Flags().Lookup("compress"))
}
