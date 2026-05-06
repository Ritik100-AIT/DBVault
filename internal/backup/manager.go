package backup

import (
	"fmt"
	"io"
	"time"

	"github.com/dbvault/dbvault/internal/compress"
	"github.com/dbvault/dbvault/internal/db"
	"github.com/dbvault/dbvault/internal/models"
	"github.com/dbvault/dbvault/internal/storage"
)

// BackupManager orchestrates backup execution using a DB connector and storage backend.
type BackupManager struct {
	Connector db.DBConnector
	Backend   storage.StorageBackend
	Config    *models.AppConfig
}

// NewBackupManager creates a backup manager for the provided connector, backend, and config.
func NewBackupManager(connector db.DBConnector, backend storage.StorageBackend, cfg *models.AppConfig) *BackupManager {
	return &BackupManager{Connector: connector, Backend: backend, Config: cfg}
}

// Run executes the backup workflow: test connection, capture dump, compress, and store.
func (m *BackupManager) Run() error {
	if m.Backend == nil {
		return fmt.Errorf("storage backend is required")
	}

	fmt.Printf("Testing connection to %s database...\n", m.Config.Database.Type)
	if err := m.Connector.TestConnection(); err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}

	fmt.Printf("Starting %s backup for database %s...\n", m.Config.Backup.Type, m.Config.Database.Name)
	raw, err := m.Connector.Backup()
	if err != nil {
		return fmt.Errorf("backup extraction failed: %w", err)
	}

	var payload io.Reader = raw
	if m.Config.Backup.Compression == "gzip" {
		payload, err = compress.GzipCompress(raw)
		if err != nil {
			return fmt.Errorf("compression failed: %w", err)
		}
	}

	outputName := fmt.Sprintf("%s-%s-%d", m.Config.Database.Type, m.Config.Database.Name, time.Now().Unix())
	if m.Config.Backup.Compression == "gzip" {
		outputName += ".gz"
	} else {
		outputName += ".bak"
	}

	savedPath, err := m.Backend.Save(outputName, payload)
	if err != nil {
		return fmt.Errorf("storage save failed: %w", err)
	}

	fmt.Printf("Backup stored at %s\n", savedPath)
	fmt.Println("Backup completed successfully.")
	return nil
}
