package sqlite_test

import (
	"context"
	"path/filepath"
	"testing"

	"certificate-generator/internal/repository/sqlite"
)

func TestOpenAndMigrate(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("sqlite.Open failed: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := sqlite.Migrate(ctx, db); err != nil {
		t.Fatalf("sqlite.Migrate failed: %v", err)
	}

	// Verify table existence
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query schema_migrations: %v", err)
	}
	if count < 1 {
		t.Errorf("Expected at least 1 applied migration, got %d", count)
	}

	// Running migrate a second time should be idempotent
	if err := sqlite.Migrate(ctx, db); err != nil {
		t.Fatalf("Secondary sqlite.Migrate failed: %v", err)
	}
}
