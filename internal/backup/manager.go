package backup

import (
	"github.com/dbvault/dbvault/internal/db"
	"github.com/dbvault/dbvault/internal/models"
)

type BackupManager struct {
	Connector db.DBConnector
	Config    *models.AppConfig
}

func NewBackupManager(connector db.DBConnector, cfg *models.AppConfig) *BackupManager {
	return &BackupManager{Connector: connector, Config: cfg}
}

func (m *BackupManager) Run() error {
	return nil
}
