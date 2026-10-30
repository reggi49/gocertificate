package sqlite_test

import (
	"context"
	"testing"
)

func TestSettingsRepository(t *testing.T) {
	_, sRepo, _, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Initial get on empty key
	val, err := sRepo.Get(ctx, "theme")
	if err != nil {
		t.Fatalf("sRepo.Get failed: %v", err)
	}
	if val != "" {
		t.Errorf("Expected empty string for unset setting, got %q", val)
	}

	// Set setting
	if err := sRepo.Set(ctx, "theme", "dark"); err != nil {
		t.Fatalf("sRepo.Set failed: %v", err)
	}

	val, err = sRepo.Get(ctx, "theme")
	if err != nil {
		t.Fatalf("sRepo.Get failed: %v", err)
	}
	if val != "dark" {
		t.Errorf("Expected 'dark', got %q", val)
	}

	// Update setting
	if err := sRepo.Set(ctx, "theme", "light"); err != nil {
		t.Fatalf("sRepo.Set update failed: %v", err)
	}
	val, err = sRepo.Get(ctx, "theme")
	if err != nil {
		t.Fatalf("sRepo.Get failed: %v", err)
	}
	if val != "light" {
		t.Errorf("Expected 'light', got %q", val)
	}

	// GetAll
	_ = sRepo.Set(ctx, "language", "id")
	all, err := sRepo.GetAll(ctx)
	if err != nil {
		t.Fatalf("sRepo.GetAll failed: %v", err)
	}
	if len(all) != 2 || all["theme"] != "light" || all["language"] != "id" {
		t.Errorf("GetAll returned unexpected map: %+v", all)
	}
}
