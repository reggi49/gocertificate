package domain_test

import (
	"testing"
	"time"

	"certificate-generator/internal/domain"
)

func TestBackupManifestValidation(t *testing.T) {
	manifest := domain.BackupManifest{
		SchemaVersion: 1,
		AppVersion:    "1.0.0",
		BackupType:    domain.BackupTypeProject,
		ProjectID:     "proj-uuid-1",
		ProjectName:   "Wisuda 2026",
		CreatedAt:     time.Now().UTC(),
		FileCount:     5,
		TotalBytes:    20480,
		Checksums: []domain.FileChecksum{
			{Path: "project.json", SHA256: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		},
	}

	if manifest.SchemaVersion != 1 {
		t.Errorf("Expected schemaVersion 1, got %d", manifest.SchemaVersion)
	}
	if manifest.BackupType != domain.BackupTypeProject {
		t.Errorf("Expected BackupTypeProject, got %s", manifest.BackupType)
	}
}
