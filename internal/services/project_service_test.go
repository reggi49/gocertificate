package services_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"certificate-generator/internal/domain"
	"certificate-generator/internal/repository/sqlite"
	"certificate-generator/internal/services"
	"certificate-generator/internal/storage/filesystem"
)

// mockFailingRepo simulates a database failure during insert to test rollback
type mockFailingRepo struct {
	sqlite.ProjectRepository
	failCreate bool
}

func (m *mockFailingRepo) Create(ctx context.Context, p *domain.Project) error {
	if m.failCreate {
		return errors.New("simulated database failure")
	}
	return m.ProjectRepository.Create(ctx, p)
}

func setupServiceTest(t *testing.T) (*services.ProjectService, *sqlite.SQLiteProjectRepository, *filesystem.StorageManager, string, func()) {
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
		t.Fatalf("Open failed: %v", err)
	}

	ctx := context.Background()
	if err := sqlite.Migrate(ctx, db); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}

	repo := sqlite.NewProjectRepository(db)
	svc := services.NewProjectService(repo, sm, nil)

	cleanup := func() {
		_ = db.Close()
	}

	return svc, repo, sm, dataDir, cleanup
}

func TestProjectServiceCRUD(t *testing.T) {
	svc, _, sm, _, cleanup := setupServiceTest(t)
	defer cleanup()

	ctx := context.Background()

	// 1. Create project
	created, err := svc.CreateProject(ctx, domain.CreateProjectInput{
		Name:        "Sertifikat Seminar Nasional 2026",
		Description: "Seminar sains & teknologi",
	})
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}
	if created.ID == "" || created.Name != "Sertifikat Seminar Nasional 2026" {
		t.Errorf("Unexpected created project: %+v", created)
	}

	// Verify filesystem tree
	projDir, err := sm.ProjectDir(created.ID)
	if err != nil {
		t.Fatalf("ProjectDir failed: %v", err)
	}
	if info, err := os.Stat(projDir); err != nil || !info.IsDir() {
		t.Errorf("Expected project directory %s to exist", projDir)
	}

	// Verify manifest
	manifest, err := sm.ReadProjectManifest(created.ID)
	if err != nil {
		t.Fatalf("ReadProjectManifest failed: %v", err)
	}
	if manifest.Name != created.Name {
		t.Errorf("Expected manifest name %s, got %s", created.Name, manifest.Name)
	}

	// 2. Get project
	fetched, err := svc.GetProject(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetProject failed: %v", err)
	}
	if fetched.Name != created.Name {
		t.Errorf("Fetched name mismatch: %s vs %s", fetched.Name, created.Name)
	}

	// 3. Rename project
	renamed, err := svc.RenameProject(ctx, created.ID, "Sertifikat Seminar Internasional 2026", "Deskripsi revisi")
	if err != nil {
		t.Fatalf("RenameProject failed: %v", err)
	}
	if renamed.Name != "Sertifikat Seminar Internasional 2026" {
		t.Errorf("Expected renamed project name, got %s", renamed.Name)
	}

	// Verify manifest was updated
	manifest, err = sm.ReadProjectManifest(created.ID)
	if err != nil {
		t.Fatalf("ReadProjectManifest failed: %v", err)
	}
	if manifest.Name != "Sertifikat Seminar Internasional 2026" {
		t.Errorf("Expected updated manifest name, got %s", manifest.Name)
	}

	// 4. Duplicate project
	duplicated, err := svc.DuplicateProject(ctx, domain.DuplicateProjectInput{
		SourceProjectID: created.ID,
		NewName:         "Sertifikat Seminar Batch 2",
	})
	if err != nil {
		t.Fatalf("DuplicateProject failed: %v", err)
	}
	if duplicated.ID == created.ID || duplicated.Name != "Sertifikat Seminar Batch 2" {
		t.Errorf("Unexpected duplicated project: %+v", duplicated)
	}

	dupDir, _ := sm.ProjectDir(duplicated.ID)
	if info, err := os.Stat(dupDir); err != nil || !info.IsDir() {
		t.Errorf("Expected duplicated directory to exist at %s", dupDir)
	}

	// 5. Archive project
	if err := svc.ArchiveProject(ctx, created.ID, true); err != nil {
		t.Fatalf("ArchiveProject failed: %v", err)
	}

	activeList, err := svc.ListProjects(ctx, false)
	if err != nil {
		t.Fatalf("ListProjects(active) failed: %v", err)
	}
	if len(activeList) != 1 || activeList[0].ID != duplicated.ID {
		t.Errorf("Expected only duplicated project to be active, got %v", activeList)
	}

	// 6. Delete project
	if err := svc.DeleteProject(ctx, created.ID); err != nil {
		t.Fatalf("DeleteProject failed: %v", err)
	}

	// Verify DB row deleted
	_, err = svc.GetProject(ctx, created.ID)
	if err != domain.ErrProjectNotFound {
		t.Errorf("Expected ErrProjectNotFound after delete, got %v", err)
	}

	// Verify folder deleted from disk
	if _, err := os.Stat(projDir); !os.IsNotExist(err) {
		t.Errorf("Expected deleted project folder %s to be removed from disk", projDir)
	}
}

func TestProjectServiceRollbackOnDBFailure(t *testing.T) {
	_, baseRepo, sm, _, cleanup := setupServiceTest(t)
	defer cleanup()

	failingRepo := &mockFailingRepo{
		ProjectRepository: baseRepo,
		failCreate:        true,
	}

	svcWithRollback := services.NewProjectService(failingRepo, sm, nil)
	ctx := context.Background()

	// Attempt creation
	_, err := svcWithRollback.CreateProject(ctx, domain.CreateProjectInput{
		Name: "Project Yang Akan Gagal",
	})
	if err == nil {
		t.Fatal("Expected error from CreateProject with failing repo, got nil")
	}

	// Ensure no orphan directory was left behind in projects/
	entries, err := os.ReadDir(sm.ProjectsDir())
	if err != nil {
		t.Fatalf("Failed to read projects dir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("Expected 0 orphan project directories after rollback, found %d", len(entries))
	}
}
