package cmd

import (
	"fmt"
	"io"

	"github.com/dbvault/dbvault/internal/compress"
	"github.com/dbvault/dbvault/internal/config"
	"github.com/dbvault/dbvault/internal/db"
	"github.com/dbvault/dbvault/internal/models"
	"github.com/dbvault/dbvault/internal/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// restoreCmd handles restore operations from a saved backup source.
var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore a backup from local storage or remote backend",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		source, _ := cmd.Flags().GetString("source")
		if source == "" {
			return fmt.Errorf("backup source is required")
		}

		if dbType, _ := cmd.Flags().GetString("db"); dbType != "" {
			cfg.Database.Type = models.DBType(dbType)
		}
		if storageType, _ := cmd.Flags().GetString("storage"); storageType != "" {
			cfg.Storage.Type = storageType
		}

		backend, err := storage.NewStorageBackend(cfg)
		if err != nil {
			return fmt.Errorf("failed to initialize storage backend: %w", err)
		}

		reader, err := backend.Load(source)
		if err != nil {
			return fmt.Errorf("failed to load backup: %w", err)
		}
		defer reader.Close()

		var payload io.Reader = reader
		if cfg.Backup.Compression == "gzip" {
			payload, err = compress.GzipDecompress(reader)
			if err != nil {
				return fmt.Errorf("decompression failed: %w", err)
			}
		}

		connector := db.NewConnector(string(cfg.Database.Type))
		if connector == nil {
			return fmt.Errorf("unsupported database type: %s", cfg.Database.Type)
		}

		fmt.Println("Loading backup from storage...")
		fmt.Printf("Restoring to %s database...\n", cfg.Database.Type)
		if err := connector.Restore(payload); err != nil {
			return fmt.Errorf("restore failed: %w", err)
		}

		fmt.Println("Restore completed successfully!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)
	restoreCmd.Flags().String("source", "", "Backup source path or key")
	restoreCmd.Flags().String("db", "", "Database type: mysql | postgres | mongodb | sqlite")
	restoreCmd.Flags().String("storage", "local", "Storage backend: local | s3")
	restoreCmd.Flags().Bool("verify", true, "Verify checksum before restore")

	viper.BindPFlag("database.type", restoreCmd.Flags().Lookup("db"))
	viper.BindPFlag("storage.type", restoreCmd.Flags().Lookup("storage"))
}
