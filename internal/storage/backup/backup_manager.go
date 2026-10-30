package backup

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"certificate-generator/internal/domain"
	"certificate-generator/internal/repository/sqlite"
	"certificate-generator/internal/storage/filesystem"

	"github.com/google/uuid"
)

// BackupManager handles backup generation and secure restore.
type BackupManager struct {
	pRepo  sqlite.ProjectRepository
	tRepo  sqlite.TemplateRepository
	rRepo  sqlite.RecipientRepository
	db     *sql.DB
	sm     *filesystem.StorageManager
	logger *slog.Logger
}

// NewBackupManager constructs a BackupManager.
func NewBackupManager(
	pRepo sqlite.ProjectRepository,
	tRepo sqlite.TemplateRepository,
	rRepo sqlite.RecipientRepository,
	db *sql.DB,
	sm *filesystem.StorageManager,
	logger *slog.Logger,
) *BackupManager {
	if logger == nil {
		logger = slog.Default()
	}
	return &BackupManager{
		pRepo:  pRepo,
		tRepo:  tRepo,
		rRepo:  rRepo,
		db:     db,
		sm:     sm,
		logger: logger,
	}
}

// BackupProject creates a standalone .zip archive of a single project.
func (b *BackupManager) BackupProject(ctx context.Context, projectID string, outputPath string) (*domain.BackupManifest, error) {
	proj, err := b.pRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	projDir, err := b.sm.ProjectDir(projectID)
	if err != nil {
		return nil, err
	}

	// Fetch all project templates and recipients for SQLite portable export
	templates, _ := b.tRepo.ListByProject(ctx, projectID)
	recipients, _ := b.rRepo.GetAllByProject(ctx, projectID)

	// Create temp working dir
	tempDir, err := os.MkdirTemp(os.TempDir(), "backup-proj-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)

	// Copy project directory tree into tempDir/data/
	dataDir := filepath.Join(tempDir, "data")
	_ = os.MkdirAll(dataDir, 0755)
	if err := copyDir(projDir, dataDir); err != nil {
		return nil, fmt.Errorf("failed to copy project files: %w", err)
	}

	// Export database rows to db_export.json
	dbExport := struct {
		Project    *domain.Project    `json:"project"`
		Templates  []domain.Template  `json:"templates"`
		Recipients []domain.Recipient `json:"recipients"`
	}{
		Project:    proj,
		Templates:  templates,
		Recipients: recipients,
	}
	dbJSON, _ := json.MarshalIndent(dbExport, "", "  ")
	_ = os.WriteFile(filepath.Join(tempDir, "db_export.json"), dbJSON, 0644)

	// Calculate checksums for all files
	var checksums []domain.FileChecksum
	var totalBytes int64
	fileCount := 0

	err = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(tempDir, path)
		hash, err := calculateSHA256(path)
		if err != nil {
			return err
		}
		checksums = append(checksums, domain.FileChecksum{
			Path:   filepath.ToSlash(rel),
			SHA256: hash,
		})
		totalBytes += info.Size()
		fileCount++
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to compute checksums: %w", err)
	}

	// Write manifest.json
	manifest := domain.BackupManifest{
		SchemaVersion: 1,
		AppVersion:    "1.0.0",
		BackupType:    domain.BackupTypeProject,
		ProjectID:     proj.ID,
		ProjectName:   proj.Name,
		CreatedAt:     time.Now().UTC(),
		FileCount:     fileCount,
		TotalBytes:    totalBytes,
		Checksums:     checksums,
	}
	manifestBytes, _ := json.MarshalIndent(manifest, "", "  ")
	_ = os.WriteFile(filepath.Join(tempDir, "manifest.json"), manifestBytes, 0644)

	// Package into ZIP
	if outputPath == "" {
		backupsDir := b.sm.BackupsDir()
		safeName := filesystem.SanitizeFilename(fmt.Sprintf("%s_backup_%s.zip", proj.Name, time.Now().Format("20060102-150405")))
		outputPath = filepath.Join(backupsDir, safeName)
	}
	_ = os.MkdirAll(filepath.Dir(outputPath), 0755)

	if err := packageDirToZip(tempDir, outputPath); err != nil {
		return nil, fmt.Errorf("failed to create zip package: %w", err)
	}

	b.logger.Info("project backup created successfully", "projectId", projectID, "outputPath", outputPath)
	return &manifest, nil
}

// BackupAll creates a full backup archive of the entire application data directory.
func (b *BackupManager) BackupAll(ctx context.Context, outputPath string) (*domain.BackupManifest, error) {
	tempDir, err := os.MkdirTemp(os.TempDir(), "backup-all-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)

	dataDir := filepath.Join(tempDir, "root")
	_ = os.MkdirAll(dataDir, 0755)

	// Copy managed root (database, projects, app settings)
	rootDir := b.sm.GetRootDir()
	_ = copyDir(rootDir, dataDir)

	var checksums []domain.FileChecksum
	var totalBytes int64
	fileCount := 0

	err = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(tempDir, path)
		hash, err := calculateSHA256(path)
		if err != nil {
			return err
		}
		checksums = append(checksums, domain.FileChecksum{
			Path:   filepath.ToSlash(rel),
			SHA256: hash,
		})
		totalBytes += info.Size()
		fileCount++
		return nil
	})
	if err != nil {
		return nil, err
	}

	manifest := domain.BackupManifest{
		SchemaVersion: 1,
		AppVersion:    "1.0.0",
		BackupType:    domain.BackupTypeAll,
		CreatedAt:     time.Now().UTC(),
		FileCount:     fileCount,
		TotalBytes:    totalBytes,
		Checksums:     checksums,
	}
	manifestBytes, _ := json.MarshalIndent(manifest, "", "  ")
	_ = os.WriteFile(filepath.Join(tempDir, "manifest.json"), manifestBytes, 0644)

	if outputPath == "" {
		backupsDir := b.sm.BackupsDir()
		safeName := filesystem.SanitizeFilename(fmt.Sprintf("CertificateGenerator_full_backup_%s.zip", time.Now().Format("20060102-150405")))
		outputPath = filepath.Join(backupsDir, safeName)
	}
	_ = os.MkdirAll(filepath.Dir(outputPath), 0755)

	if err := packageDirToZip(tempDir, outputPath); err != nil {
		return nil, err
	}

	b.logger.Info("full app backup created successfully", "outputPath", outputPath)
	return &manifest, nil
}

// RestoreProject restores a project from an archive with ZIP Slip protection, manifest verification, and conflict resolution.
func (b *BackupManager) RestoreProject(ctx context.Context, opts domain.RestoreOptions) (*domain.RestoreResult, error) {
	if _, err := os.Stat(opts.ArchiveFilePath); err != nil {
		return nil, fmt.Errorf("archive file not found: %w", err)
	}

	tempDir, err := os.MkdirTemp(os.TempDir(), "restore-proj-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir) // Guaranteed cleanup of temporary files

	// 1. Secure extraction with ZIP Slip protection
	if err := safeExtractZip(opts.ArchiveFilePath, tempDir); err != nil {
		return nil, err
	}

	// 2. Validate manifest presence
	manifestPath := filepath.Join(tempDir, "manifest.json")
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, domain.ErrMissingManifest
	}

	var manifest domain.BackupManifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return nil, fmt.Errorf("invalid manifest JSON: %w", err)
	}

	if manifest.BackupType != domain.BackupTypeProject {
		return nil, fmt.Errorf("archive is not a project backup (type=%s)", manifest.BackupType)
	}

	// 3. Verify SHA-256 Checksums
	for _, cs := range manifest.Checksums {
		targetPath := filepath.Join(tempDir, filepath.FromSlash(cs.Path))
		actualHash, err := calculateSHA256(targetPath)
		if err != nil || actualHash != cs.SHA256 {
			return nil, fmt.Errorf("%w for file %s", domain.ErrChecksumMismatch, cs.Path)
		}
	}

	// 4. Read db_export.json
	dbExportPath := filepath.Join(tempDir, "db_export.json")
	dbData, err := os.ReadFile(dbExportPath)
	if err != nil {
		return nil, fmt.Errorf("missing db_export.json in archive: %w", err)
	}

	var dbExport struct {
		Project    *domain.Project    `json:"project"`
		Templates  []domain.Template  `json:"templates"`
		Recipients []domain.Recipient `json:"recipients"`
	}
	if err := json.Unmarshal(dbData, &dbExport); err != nil {
		return nil, fmt.Errorf("failed to parse db_export.json: %w", err)
	}
	if dbExport.Project == nil {
		return nil, fmt.Errorf("invalid db_export.json: project record missing")
	}

	// 5. Conflict Resolution
	existingProj, _ := b.pRepo.GetByID(ctx, dbExport.Project.ID)
	targetProjectID := dbExport.Project.ID
	targetProjectName := dbExport.Project.Name
	resultStatus := "restored"

	if existingProj != nil {
		switch opts.ConflictStrategy {
		case domain.RestoreConflictAbort:
			return nil, fmt.Errorf("%w: project ID '%s' already exists", domain.ErrProjectConflict, existingProj.ID)

		case domain.RestoreConflictOverwrite:
			// Delete existing project first
			_ = b.tRepo.Delete(ctx, existingProj.ID)
			_ = b.rRepo.DeleteAllByProject(ctx, existingProj.ID)
			_ = b.pRepo.Delete(ctx, existingProj.ID)
			_ = b.sm.DeleteProjectDirectory(existingProj.ID)
			resultStatus = "overwritten"

		case domain.RestoreConflictCreateNew:
			fallthrough
		default:
			// Generate fresh ID
			targetProjectID = uuid.New().String()
			targetProjectName = fmt.Sprintf("%s (Restored)", dbExport.Project.Name)
			resultStatus = "created_as_copy"
		}
	}

	// 6. Transactional Database Restore
	tx, err := b.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	now := time.Now().UTC()
	nowStr := now.Format(time.RFC3339)

	// Insert project
	_, err = tx.ExecContext(ctx, `
		INSERT INTO projects (id, name, description, created_at, updated_at, is_archived)
		VALUES (?, ?, ?, ?, ?, ?)
	`, targetProjectID, targetProjectName, dbExport.Project.Description, nowStr, nowStr, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to insert restored project: %w", err)
	}

	// Insert templates with remapped ProjectID
	tmplStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO templates (id, project_id, name, document_json, schema_version, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return nil, err
	}
	defer tmplStmt.Close()

	for _, tmpl := range dbExport.Templates {
		newTmplID := tmpl.ID
		if targetProjectID != dbExport.Project.ID {
			newTmplID = uuid.New().String()
		}
		_, err := tmplStmt.ExecContext(ctx, newTmplID, targetProjectID, tmpl.Name, tmpl.DocumentJSON, tmpl.SchemaVersion, nowStr, nowStr)
		if err != nil {
			return nil, fmt.Errorf("failed to insert restored template: %w", err)
		}
	}

	// Insert recipients with remapped ProjectID
	recStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO recipients (id, project_id, full_name, email, organization, event_name, issue_date, certificate_id, custom_fields_json, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return nil, err
	}
	defer recStmt.Close()

	for _, rec := range dbExport.Recipients {
		newRecID := rec.ID
		if targetProjectID != dbExport.Project.ID {
			newRecID = uuid.New().String()
		}
		_, err := recStmt.ExecContext(ctx, newRecID, targetProjectID, rec.FullName, rec.Email, rec.Organization, rec.EventName, rec.IssueDate, rec.CertificateID, rec.CustomFieldsJSON, nowStr, nowStr)
		if err != nil {
			return nil, fmt.Errorf("failed to insert restored recipient: %w", err)
		}
	}

	// 7. Atomic Filesystem Restore
	targetProjDir, err := b.sm.ProjectDir(targetProjectID)
	if err != nil {
		return nil, err
	}
	_ = os.RemoveAll(targetProjDir)

	srcDataDir := filepath.Join(tempDir, "data")
	if info, err := os.Stat(srcDataDir); err == nil && info.IsDir() {
		if err := copyDir(srcDataDir, targetProjDir); err != nil {
			_ = os.RemoveAll(targetProjDir)
			return nil, fmt.Errorf("failed to copy restored directory tree: %w", err)
		}
	} else {
		_, _ = b.sm.CreateProjectDirectory(targetProjectID)
	}

	// Update project.json manifest with final ID and name
	finalManifest := domain.ProjectManifest{
		SchemaVersion: 1,
		ID:            targetProjectID,
		Name:          targetProjectName,
		Description:   dbExport.Project.Description,
		CreatedAt:     nowStr,
		UpdatedAt:     nowStr,
	}
	_ = b.sm.WriteProjectManifest(targetProjectID, finalManifest)

	// Commit database transaction
	if err := tx.Commit(); err != nil {
		_ = os.RemoveAll(targetProjDir)
		return nil, fmt.Errorf("failed to commit restore transaction: %w", err)
	}

	b.logger.Info("project restored successfully",
		"projectId", targetProjectID,
		"name", targetProjectName,
		"status", resultStatus,
	)

	return &domain.RestoreResult{
		ProjectID:   targetProjectID,
		ProjectName: targetProjectName,
		Status:      resultStatus,
		Message:     fmt.Sprintf("Proyek '%s' berhasil dipulihkan", targetProjectName),
	}, nil
}

// safeExtractZip extracts zip entries enforcing strict ZIP Slip protection.
func safeExtractZip(zipPath, destDir string) error {
	cleanDest := filepath.Clean(destDir)

	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrInvalidBackupArchive, err)
	}
	defer zr.Close()

	for _, f := range zr.File {
		cleanName := filepath.Clean(f.Name)

		// Check for malicious paths (ZIP Slip prevention)
		if strings.HasPrefix(cleanName, "..") || filepath.IsAbs(cleanName) || strings.Contains(f.Name, "../") || strings.Contains(f.Name, "..\\") {
			return fmt.Errorf("%w: entry %q attempts path traversal", domain.ErrZipSlipDetected, f.Name)
		}

		targetPath := filepath.Join(cleanDest, cleanName)
		if !strings.HasPrefix(targetPath, cleanDest+string(filepath.Separator)) && targetPath != cleanDest {
			return fmt.Errorf("%w: entry %q escapes destination", domain.ErrZipSlipDetected, f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func calculateSHA256(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func packageDirToZip(sourceDir, zipPath string) error {
	tempZip, err := os.CreateTemp(filepath.Dir(zipPath), "pkg-*.zip")
	if err != nil {
		return err
	}
	tempName := tempZip.Name()
	defer os.Remove(tempName)

	zw := zip.NewWriter(tempZip)

	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(rel)
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		w, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(w, f)
		return err
	})

	if err != nil {
		zw.Close()
		tempZip.Close()
		return err
	}

	if err := zw.Close(); err != nil {
		tempZip.Close()
		return err
	}
	if err := tempZip.Close(); err != nil {
		return err
	}

	return os.Rename(tempName, zipPath)
}

func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
