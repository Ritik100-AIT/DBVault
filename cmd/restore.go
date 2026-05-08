package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/dbvault/dbvault/internal/compress"
	"github.com/dbvault/dbvault/internal/config"
	"github.com/dbvault/dbvault/internal/logger"
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
		if compress, _ := cmd.Flags().GetString("compress"); compress != "" {
			cfg.Backup.Compression = compress
		}
		if cfg.Backup.Compression == "" && strings.HasSuffix(source, ".gz") {
			cfg.Backup.Compression = "gzip"
		}
		verify, _ := cmd.Flags().GetBool("verify")
		if err := config.ValidateConfig(cfg); err != nil {
			return fmt.Errorf("invalid configuration: %w", err)
		}

		loggerInstance, err := logger.NewLogger(&cfg.Logging)
		if err != nil {
			return fmt.Errorf("failed to initialize logger: %w", err)
		}

		backend, err := storage.NewStorageBackend(cfg)
		if err != nil {
			return fmt.Errorf("failed to initialize storage backend: %w", err)
		}

		if verify {
			if err := verifyBackupChecksum(backend, source); err != nil {
				return fmt.Errorf("backup verification failed: %w", err)
			}
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

		connector := newDBConnector(&cfg.Database)
		if connector == nil {
			return fmt.Errorf("unsupported database type: %s", cfg.Database.Type)
		}

		if loggerInstance != nil {
			loggerInstance.Info(fmt.Sprintf("Restoring backup %s to %s", source, cfg.Database.Type))
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

func verifyBackupChecksum(backend storage.StorageBackend, source string) error {
	metaSource := source + ".meta.json"
	reader, err := backend.Load(metaSource)
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}
	defer reader.Close()

	var record models.BackupRecord
	if err := json.NewDecoder(reader).Decode(&record); err != nil {
		return fmt.Errorf("invalid metadata format: %w", err)
	}

	backupReader, err := backend.Load(source)
	if err != nil {
		return fmt.Errorf("failed to reload backup for verification: %w", err)
	}
	defer backupReader.Close()

	h := sha256.New()
	if _, err := io.Copy(h, backupReader); err != nil {
		return fmt.Errorf("failed to hash backup: %w", err)
	}

	checksum := hex.EncodeToString(h.Sum(nil))
	if checksum != record.Checksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", record.Checksum, checksum)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(restoreCmd)
	restoreCmd.Flags().String("source", "", "Backup source path or key")
	restoreCmd.Flags().String("db", "", "Database type: mysql | postgres | mongodb | sqlite")
	restoreCmd.Flags().String("storage", "local", "Storage backend: local | s3")
	restoreCmd.Flags().String("compress", "", "Compression method for the backup: gzip | none")
	restoreCmd.Flags().String("local-path", "", "Local backup storage path")
	restoreCmd.Flags().String("s3-bucket", "", "S3 bucket name")
	restoreCmd.Flags().String("s3-region", "", "S3 region")
	restoreCmd.Flags().String("s3-access-key", "", "S3 access key")
	restoreCmd.Flags().String("s3-secret-key", "", "S3 secret key")
	restoreCmd.Flags().String("s3-endpoint", "", "S3 or MinIO endpoint URL")
	restoreCmd.Flags().Bool("s3-force-path-style", false, "Use path style for S3 endpoints")
	restoreCmd.Flags().String("s3-prefix", "", "S3 object key prefix")
	restoreCmd.Flags().Bool("verify", true, "Verify checksum before restore")

	viper.BindPFlag("database.type", restoreCmd.Flags().Lookup("db"))
	viper.BindPFlag("storage.type", restoreCmd.Flags().Lookup("storage"))
	viper.BindPFlag("storage.local.path", restoreCmd.Flags().Lookup("local-path"))
	viper.BindPFlag("storage.s3.bucket", restoreCmd.Flags().Lookup("s3-bucket"))
	viper.BindPFlag("storage.s3.region", restoreCmd.Flags().Lookup("s3-region"))
	viper.BindPFlag("storage.s3.access_key", restoreCmd.Flags().Lookup("s3-access-key"))
	viper.BindPFlag("storage.s3.secret_key", restoreCmd.Flags().Lookup("s3-secret-key"))
	viper.BindPFlag("storage.s3.endpoint", restoreCmd.Flags().Lookup("s3-endpoint"))
	viper.BindPFlag("storage.s3.force_path_style", restoreCmd.Flags().Lookup("s3-force-path-style"))
	viper.BindPFlag("storage.s3.prefix", restoreCmd.Flags().Lookup("s3-prefix"))
	viper.BindPFlag("backup.compression", restoreCmd.Flags().Lookup("compress"))
}
