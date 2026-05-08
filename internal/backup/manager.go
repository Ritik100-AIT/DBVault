package backup

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/dbvault/dbvault/internal/compress"
	"github.com/dbvault/dbvault/internal/db"
	"github.com/dbvault/dbvault/internal/logger"
	"github.com/dbvault/dbvault/internal/models"
	"github.com/dbvault/dbvault/internal/notify"
	"github.com/dbvault/dbvault/internal/storage"
)

// BackupManager orchestrates backup execution using a DB connector and storage backend.
type BackupManager struct {
	Connector db.DBConnector
	Backend   storage.StorageBackend
	Config    *models.AppConfig
	Notifier  *notify.SlackNotifier
	Logger    *logger.Logger
}

// NewBackupManager creates a backup manager for the provided connector, backend, config, notifier, and logger.
func NewBackupManager(connector db.DBConnector, backend storage.StorageBackend, cfg *models.AppConfig, notifier *notify.SlackNotifier, logger *logger.Logger) *BackupManager {
	return &BackupManager{Connector: connector, Backend: backend, Config: cfg, Notifier: notifier, Logger: logger}
}

// Run executes the backup workflow: test connection, capture dump, compress, and store.
func (m *BackupManager) Run() (err error) {
	if m.Backend == nil {
		return fmt.Errorf("storage backend is required")
	}

	var savedPath string
	defer func() {
		if m.Notifier != nil {
			status := "succeeded"
			if err != nil {
				status = "failed"
			}
			message := fmt.Sprintf("DBVault backup %s for %s/%s", status, m.Config.Database.Type, m.Config.Database.Name)
			if savedPath != "" {
				message = fmt.Sprintf("%s\nStored at: %s", message, savedPath)
			}
			if notifyErr := m.Notifier.Notify(message); notifyErr != nil {
				fmt.Printf("warning: failed to send Slack notification: %v\n", notifyErr)
			}
		}
	}()

	if m.Logger != nil {
		m.Logger.Info(fmt.Sprintf("Testing connection to %s database", m.Config.Database.Type))
	}
	fmt.Printf("Testing connection to %s database...\n", m.Config.Database.Type)
	if err = m.Connector.TestConnection(); err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}

	if m.Logger != nil {
		m.Logger.Info(fmt.Sprintf("Starting %s backup for database %s", m.Config.Backup.Type, m.Config.Database.Name))
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

	// Buffer compressed data into memory so S3 can seek for retries
	var buf bytes.Buffer
	h := sha256.New()
	mw := io.MultiWriter(&buf, h)
	if _, err := io.Copy(mw, payload); err != nil {
		return fmt.Errorf("failed to buffer backup data: %w", err)
	}

	outputName := fmt.Sprintf("%s-%s-%s", m.Config.Database.Type, m.Config.Database.Name, time.Now().UTC().Format("20060102T150405Z"))
	if m.Config.Backup.Compression == "gzip" {
		outputName += ".gz"
	} else {
		outputName += ".bak"
	}

	savedPath, err = m.Backend.Save(outputName, bytes.NewReader(buf.Bytes()))
	if err != nil {
		return fmt.Errorf("storage save failed: %w", err)
	}

	metadata := models.BackupRecord{
		ID:          fmt.Sprintf("backup-%d", time.Now().UTC().UnixNano()),
		Type:        string(m.Config.Backup.Type),
		Database:    string(m.Config.Database.Type),
		Path:        savedPath,
		Checksum:    hex.EncodeToString(h.Sum(nil)),
		SizeBytes:   int64(buf.Len()),
		Compression: m.Config.Backup.Compression,
		CreatedAt:   time.Now().UTC(),
	}

	metaName := outputName + ".meta.json"
	metaJSON, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("metadata marshaling failed: %w", err)
	}

	if _, err = m.Backend.Save(metaName, bytes.NewReader(metaJSON)); err != nil {
		return fmt.Errorf("metadata save failed: %w", err)
	}

	if m.Logger != nil {
		m.Logger.Info(fmt.Sprintf("Backup stored at %s", savedPath))
	}
	fmt.Printf("Backup stored at %s\n", savedPath)
	fmt.Println("Backup completed successfully.")
	return nil
}
