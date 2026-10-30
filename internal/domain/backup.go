package domain

import (
	"errors"
	"time"
)

// Backup Type constants
const (
	BackupTypeProject = "project"
	BackupTypeAll     = "full"
)

// Restore Conflict Strategy constants
const (
	RestoreConflictCreateNew = "create_new" // Generate new Project ID and rename if existing
	RestoreConflictOverwrite = "overwrite"  // Overwrite existing project data
	RestoreConflictAbort     = "abort"      // Abort restore operation
)

var (
	ErrInvalidBackupArchive = errors.New("invalid backup archive")
	ErrMissingManifest      = errors.New("backup manifest.json not found in archive")
	ErrChecksumMismatch     = errors.New("integrity verification failed: archive checksum mismatch")
	ErrZipSlipDetected      = errors.New("security alert: malicious ZIP Slip path traversal detected in archive entry")
	ErrProjectConflict      = errors.New("a project with this ID or name already exists")
)

// FileChecksum holds relative file path and SHA-256 hash.
type FileChecksum struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

// BackupManifest represents the portable metadata stored at manifest.json within backup archives.
type BackupManifest struct {
	SchemaVersion int            `json:"schemaVersion"`
	AppVersion    string         `json:"appVersion"`
	BackupType    string         `json:"backupType"` // "project" or "full"
	ProjectID     string         `json:"projectId,omitempty"`
	ProjectName   string         `json:"projectName,omitempty"`
	CreatedAt     time.Time      `json:"createdAt"`
	FileCount     int            `json:"fileCount"`
	TotalBytes    int64          `json:"totalBytes"`
	Checksums     []FileChecksum `json:"checksums,omitempty"`
}

// RestoreOptions holds user options when restoring an archive.
type RestoreOptions struct {
	ArchiveFilePath  string `json:"archiveFilePath"`
	ConflictStrategy string `json:"conflictStrategy"` // "create_new", "overwrite", "abort"
}

// RestoreResult represents the outcome of an archive restoration.
type RestoreResult struct {
	ProjectID   string `json:"projectId"`
	ProjectName string `json:"projectName"`
	Status      string `json:"status"` // "restored", "overwritten", "created_as_copy"
	Message     string `json:"message"`
}
