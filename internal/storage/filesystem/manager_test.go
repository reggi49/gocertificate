package filesystem_test

import (
	"os"
	"path/filepath"
	"testing"

	"certificate-generator/internal/domain"
	"certificate-generator/internal/storage/filesystem"
)

func TestStorageManager(t *testing.T) {
	tempDir := t.TempDir()
	rootDir := filepath.Join(tempDir, "CertificateGenerator")

	sm, err := filesystem.NewStorageManager(rootDir)
	if err != nil {
		t.Fatalf("NewStorageManager failed: %v", err)
	}

	// Verify base directories
	expectedDirs := []string{
		sm.DatabaseDir(),
		sm.ProjectsDir(),
		sm.BackupsDir(),
		sm.LogsDir(),
		sm.CacheDir(),
	}
	for _, dir := range expectedDirs {
		info, err := os.Stat(dir)
		if err != nil || !info.IsDir() {
			t.Errorf("Expected base directory %s to exist as dir", dir)
		}
	}

	// Test SafeResolve and path traversal protection
	traversalPaths := []string{
		"../../etc/passwd",
		"projects/../../outside",
		"../outside",
	}
	for _, p := range traversalPaths {
		_, err := sm.SafeResolve(p)
		if err != filesystem.ErrPathTraversal {
			t.Errorf("Expected ErrPathTraversal for %q, got: %v", p, err)
		}
	}

	validSub, err := sm.SafeResolve("projects/proj-1")
	if err != nil {
		t.Errorf("Expected valid path resolution, got %v", err)
	}
	if !filepath.IsAbs(validSub) {
		t.Errorf("Expected absolute path, got %s", validSub)
	}

	// Test CreateProjectDirectory scaffolding
	projectID := "proj-test-101"
	projDir, err := sm.CreateProjectDirectory(projectID)
	if err != nil {
		t.Fatalf("CreateProjectDirectory failed: %v", err)
	}

	subdirs := []string{
		"templates",
		filepath.Join("assets", "backgrounds"),
		filepath.Join("assets", "logos"),
		filepath.Join("assets", "signatures"),
		filepath.Join("assets", "fonts"),
		"exports",
	}
	for _, sub := range subdirs {
		fullSub := filepath.Join(projDir, sub)
		info, err := os.Stat(fullSub)
		if err != nil || !info.IsDir() {
			t.Errorf("Expected subfolder %s to exist", fullSub)
		}
	}

	// Test WriteProjectManifest & ReadProjectManifest
	manifest := domain.ProjectManifest{
		SchemaVersion: 1,
		ID:            projectID,
		Name:          "Test Scaffolding Project",
		Description:   "Description for test",
		CreatedAt:     "2026-10-30T10:00:00Z",
		UpdatedAt:     "2026-10-30T10:00:00Z",
	}
	if err := sm.WriteProjectManifest(projectID, manifest); err != nil {
		t.Fatalf("WriteProjectManifest failed: %v", err)
	}

	readManifest, err := sm.ReadProjectManifest(projectID)
	if err != nil {
		t.Fatalf("ReadProjectManifest failed: %v", err)
	}
	if readManifest.Name != manifest.Name || readManifest.ID != manifest.ID {
		t.Errorf("Manifest content mismatch: %+v vs %+v", readManifest, manifest)
	}

	// Test DuplicateProjectDirectory
	dupID := "proj-test-101-copy"
	if err := sm.DuplicateProjectDirectory(projectID, dupID); err != nil {
		t.Fatalf("DuplicateProjectDirectory failed: %v", err)
	}
	dupManifest, err := sm.ReadProjectManifest(dupID)
	if err != nil {
		t.Fatalf("ReadProjectManifest on duplicated project failed: %v", err)
	}
	if dupManifest.Name != manifest.Name {
		t.Errorf("Expected copied manifest to match original, got %s", dupManifest.Name)
	}

	// Test CalculateDirSize
	size, err := filesystem.CalculateDirSize(projDir)
	if err != nil {
		t.Fatalf("CalculateDirSize failed: %v", err)
	}
	if size <= 0 {
		t.Errorf("Expected non-zero dir size with manifest file, got %d", size)
	}

	// Test DeleteProjectDirectory
	if err := sm.DeleteProjectDirectory(dupID); err != nil {
		t.Fatalf("DeleteProjectDirectory failed: %v", err)
	}
	dupDir, _ := sm.ProjectDir(dupID)
	if _, err := os.Stat(dupDir); !os.IsNotExist(err) {
		t.Errorf("Expected duplicated directory to be deleted, but it still exists")
	}
}
