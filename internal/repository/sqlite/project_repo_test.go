package sqlite_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"certificate-generator/internal/domain"
	"certificate-generator/internal/repository/sqlite"
)

func setupTestDB(t *testing.T) (*sqlite.SQLiteProjectRepository, *sqlite.SQLiteSettingsRepository, *sql.DB, func()) {
	t.Helper()
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_repo.db")

	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	ctx := context.Background()
	if err := sqlite.Migrate(ctx, db); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}

	pRepo := sqlite.NewProjectRepository(db)
	sRepo := sqlite.NewSettingsRepository(db)

	cleanup := func() {
		_ = db.Close()
	}

	return pRepo, sRepo, db, cleanup
}

func TestProjectRepositoryCRUD(t *testing.T) {
	pRepo, _, _, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().UTC()

	proj := &domain.Project{
		ID:          "proj-uuid-1",
		Name:        "Wisuda Akbar 2026",
		Description: "Sertifikat kelulusan angkatan 2026",
		CreatedAt:   now,
		UpdatedAt:   now,
		IsArchived:  false,
	}

	// 1. Create
	if err := pRepo.Create(ctx, proj); err != nil {
		t.Fatalf("pRepo.Create failed: %v", err)
	}

	// 2. GetByID
	fetched, err := pRepo.GetByID(ctx, proj.ID)
	if err != nil {
		t.Fatalf("pRepo.GetByID failed: %v", err)
	}
	if fetched.Name != proj.Name || fetched.Description != proj.Description {
		t.Errorf("Fetched project mismatch: %+v vs %+v", fetched, proj)
	}

	// 3. GetByName
	byName, err := pRepo.GetByName(ctx, proj.Name)
	if err != nil {
		t.Fatalf("pRepo.GetByName failed: %v", err)
	}
	if byName.ID != proj.ID {
		t.Errorf("GetByName ID mismatch: %s vs %s", byName.ID, proj.ID)
	}

	// 4. Update
	proj.Name = "Wisuda Akbar 2026 (Updated)"
	proj.Description = "Deskripsi baru"
	proj.UpdatedAt = time.Now().UTC()
	if err := pRepo.Update(ctx, proj); err != nil {
		t.Fatalf("pRepo.Update failed: %v", err)
	}

	updated, err := pRepo.GetByID(ctx, proj.ID)
	if err != nil {
		t.Fatalf("pRepo.GetByID failed: %v", err)
	}
	if updated.Name != "Wisuda Akbar 2026 (Updated)" {
		t.Errorf("Update failed to change name: %s", updated.Name)
	}

	// 5. Count & List
	total, active, archived, err := pRepo.Count(ctx)
	if err != nil {
		t.Fatalf("pRepo.Count failed: %v", err)
	}
	if total != 1 || active != 1 || archived != 0 {
		t.Errorf("Unexpected count: total=%d, active=%d, archived=%d", total, active, archived)
	}

	// 6. SetArchived
	if err := pRepo.SetArchived(ctx, proj.ID, true); err != nil {
		t.Fatalf("pRepo.SetArchived failed: %v", err)
	}
	activeList, err := pRepo.List(ctx, false)
	if err != nil {
		t.Fatalf("pRepo.List(active) failed: %v", err)
	}
	if len(activeList) != 0 {
		t.Errorf("Expected 0 active projects after archiving, got %d", len(activeList))
	}

	allList, err := pRepo.List(ctx, true)
	if err != nil {
		t.Fatalf("pRepo.List(all) failed: %v", err)
	}
	if len(allList) != 1 {
		t.Errorf("Expected 1 project in all list, got %d", len(allList))
	}

	// 7. Delete
	if err := pRepo.Delete(ctx, proj.ID); err != nil {
		t.Fatalf("pRepo.Delete failed: %v", err)
	}

	_, err = pRepo.GetByID(ctx, proj.ID)
	if err != domain.ErrProjectNotFound {
		t.Errorf("Expected ErrProjectNotFound after delete, got %v", err)
	}
}
