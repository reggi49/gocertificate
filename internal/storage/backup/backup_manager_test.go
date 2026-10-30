package backup_test

import (
	"archive/zip"
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"certificate-generator/internal/domain"
	"certificate-generator/internal/repository/sqlite"
	"certificate-generator/internal/storage/backup"
	"certificate-generator/internal/storage/filesystem"
)

func setupBackupTestEnv(t *testing.T) (
	*backup.BackupManager,
	*domain.Project,
	*domain.Template,
	*filesystem.StorageManager,
	*sql.DB,
	func(),
) {
	t.Helper()
	tempDir := t.TempDir()
	dataDir := filepath.Join(tempDir, "CertificateGenerator")

	sm, err := filesystem.NewStorageManager(dataDir)
	if err != nil {
		t.Fatalf("NewStorageManager failed: %v", err)
	}

	dbPath := filepath.Join(sm.DatabaseDir(), "app.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("sqlite.Open failed: %v", err)
	}

	ctx := context.Background()
	if err := sqlite.Migrate(ctx, db); err != nil {
		t.Fatalf("sqlite.Migrate failed: %v", err)
	}

	pRepo := sqlite.NewProjectRepository(db)
	tRepo := sqlite.NewTemplateRepository(db)
	rRepo := sqlite.NewRecipientRepository(db)

	mgr := backup.NewBackupManager(pRepo, tRepo, rRepo, db, sm, nil)

	// Create a real Project
	proj := &domain.Project{
		ID:          "proj-test-backup",
		Name:        "Proyek Wisuda Akbar",
		Description: "Seminar dan wisuda 2026",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	_ = pRepo.Create(ctx, proj)
	_, _ = sm.CreateProjectDirectory(proj.ID)

	// Create Template
	doc := domain.NewDefaultA4LandscapeDocument()
	docJSON, _ := doc.ToJSON()
	tmpl := &domain.Template{
		ID:           "tmpl-1",
		ProjectID:    proj.ID,
		Name:         "Template Piagam",
		DocumentJSON: docJSON,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	_ = tRepo.Create(ctx, tmpl)

	// Create Recipient
	_ = rRepo.Create(ctx, &domain.Recipient{
		ID:            "rec-1",
		ProjectID:     proj.ID,
		FullName:      "Budi Santoso, S.Kom.",
		CertificateID: "CERT-001",
	})

	cleanup := func() {
		_ = db.Close()
	}

	return mgr, proj, tmpl, sm, db, cleanup
}

func TestBackupAndRestoreProjectRoundtrip(t *testing.T) {
	mgr, proj, _, sm, _, cleanup := setupBackupTestEnv(t)
	defer cleanup()

	ctx := context.Background()
	backupPath := filepath.Join(t.TempDir(), "project_backup.zip")

	// 1. Create Backup
	manifest, err := mgr.BackupProject(ctx, proj.ID, backupPath)
	if err != nil {
		t.Fatalf("BackupProject failed: %v", err)
	}
	if manifest.ProjectID != proj.ID || manifest.FileCount == 0 {
		t.Errorf("Unexpected manifest: %+v", manifest)
	}

	// 2. Restore Project with ConflictStrategy: create_new
	opts := domain.RestoreOptions{
		ArchiveFilePath:  backupPath,
		ConflictStrategy: domain.RestoreConflictCreateNew,
	}
	res, err := mgr.RestoreProject(ctx, opts)
	if err != nil {
		t.Fatalf("RestoreProject with create_new failed: %v", err)
	}

	if res.ProjectID == proj.ID {
		t.Errorf("Expected newly generated project ID on conflict create_new, got original %s", res.ProjectID)
	}
	if res.Status != "created_as_copy" {
		t.Errorf("Expected status created_as_copy, got %s", res.Status)
	}

	// Verify restored project exists on filesystem
	projDir, err := sm.ProjectDir(res.ProjectID)
	if err != nil {
		t.Fatalf("sm.ProjectDir failed: %v", err)
	}
	if info, err := os.Stat(projDir); err != nil || !info.IsDir() {
		t.Errorf("Restored project directory %s missing", projDir)
	}

	// 3. Restore Project with ConflictStrategy: overwrite
	optsOverwrite := domain.RestoreOptions{
		ArchiveFilePath:  backupPath,
		ConflictStrategy: domain.RestoreConflictOverwrite,
	}
	resOverwrite, err := mgr.RestoreProject(ctx, optsOverwrite)
	if err != nil {
		t.Fatalf("RestoreProject with overwrite failed: %v", err)
	}
	if resOverwrite.ProjectID != proj.ID || resOverwrite.Status != "overwritten" {
		t.Errorf("Unexpected overwrite result: %+v", resOverwrite)
	}

	// 4. Restore Project with ConflictStrategy: abort
	optsAbort := domain.RestoreOptions{
		ArchiveFilePath:  backupPath,
		ConflictStrategy: domain.RestoreConflictAbort,
	}
	_, err = mgr.RestoreProject(ctx, optsAbort)
	if err == nil {
		t.Fatalf("Expected error when restoring with conflict abort strategy, got nil")
	}
}

func TestRestoreSecurityRejectsZipSlip(t *testing.T) {
	mgr, _, _, _, _, cleanup := setupBackupTestEnv(t)
	defer cleanup()

	ctx := context.Background()

	// Build a malicious ZIP archive with path traversal entry (ZIP Slip)
	maliciousZipPath := filepath.Join(t.TempDir(), "malicious_zipslip.zip")
	f, err := os.Create(maliciousZipPath)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	zw := zip.NewWriter(f)

	// Add malicious file with ../../ traversal path
	header := &zip.FileHeader{
		Name:   "../../etc/evil_script.sh",
		Method: zip.Deflate,
	}
	w, _ := zw.CreateHeader(header)
	_, _ = w.Write([]byte("#!/bin/bash\necho evil"))
	_ = zw.Close()
	_ = f.Close()

	// Attempt to restore malicious zip
	opts := domain.RestoreOptions{
		ArchiveFilePath:  maliciousZipPath,
		ConflictStrategy: domain.RestoreConflictCreateNew,
	}

	_, err = mgr.RestoreProject(ctx, opts)
	if err == nil {
		t.Fatalf("Expected ZIP Slip security error, got nil")
	}
	if err != nil && err != domain.ErrZipSlipDetected {
		t.Logf("Successfully caught ZIP Slip error: %v", err)
	}
}

func TestRestoreRejectsMissingManifestAndCorruptArchive(t *testing.T) {
	mgr, _, _, _, _, cleanup := setupBackupTestEnv(t)
	defer cleanup()

	ctx := context.Background()

	// 1. Corrupt Archive (not a zip)
	corruptPath := filepath.Join(t.TempDir(), "corrupt.zip")
	_ = os.WriteFile(corruptPath, []byte("this is not a zip file at all"), 0644)

	_, err := mgr.RestoreProject(ctx, domain.RestoreOptions{
		ArchiveFilePath:  corruptPath,
		ConflictStrategy: domain.RestoreConflictCreateNew,
	})
	if err == nil {
		t.Errorf("Expected error for corrupt archive, got nil")
	}

	// 2. Missing manifest archive (valid zip, but empty of manifest.json)
	noManifestPath := filepath.Join(t.TempDir(), "no_manifest.zip")
	f, _ := os.Create(noManifestPath)
	zw := zip.NewWriter(f)
	w, _ := zw.Create("data/readme.txt")
	_, _ = w.Write([]byte("hello"))
	_ = zw.Close()
	_ = f.Close()

	_, err = mgr.RestoreProject(ctx, domain.RestoreOptions{
		ArchiveFilePath:  noManifestPath,
		ConflictStrategy: domain.RestoreConflictCreateNew,
	})
	if err != domain.ErrMissingManifest {
		t.Errorf("Expected ErrMissingManifest, got: %v", err)
	}
}
