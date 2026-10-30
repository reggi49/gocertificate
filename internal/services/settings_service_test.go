package services_test

import (
	"context"
	"path/filepath"
	"testing"

	"certificate-generator/internal/domain"
	"certificate-generator/internal/repository/sqlite"
	"certificate-generator/internal/services"
)

func TestSettingsService(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "settings.db")

	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := sqlite.Migrate(ctx, db); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}

	repo := sqlite.NewSettingsRepository(db)
	svc := services.NewSettingsService(repo)

	// 1. Initial settings should be default
	initial, err := svc.GetSettings(ctx)
	if err != nil {
		t.Fatalf("GetSettings failed: %v", err)
	}
	if initial.Theme != "system" || initial.Language != "id" || initial.AutoSaveIntervalSec != 10 {
		t.Errorf("Unexpected default settings: %+v", initial)
	}

	// 2. Update settings
	update := domain.AppSettings{
		DefaultOutputDir:    "/Users/test/Certificates",
		Theme:               "dark",
		Language:            "en",
		AutoSaveIntervalSec: 15,
	}
	if err := svc.UpdateSettings(ctx, update); err != nil {
		t.Fatalf("UpdateSettings failed: %v", err)
	}

	// 3. Verify persistence
	reloaded, err := svc.GetSettings(ctx)
	if err != nil {
		t.Fatalf("GetSettings reloaded failed: %v", err)
	}
	if reloaded.Theme != "dark" || reloaded.Language != "en" || reloaded.AutoSaveIntervalSec != 15 || reloaded.DefaultOutputDir != "/Users/test/Certificates" {
		t.Errorf("Settings mismatch after reload: %+v", reloaded)
	}
}
