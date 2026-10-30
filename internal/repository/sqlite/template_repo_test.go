package sqlite_test

import (
	"context"
	"testing"
	"time"

	"certificate-generator/internal/domain"
	"certificate-generator/internal/repository/sqlite"
)

func TestTemplateRepositoryCRUD(t *testing.T) {
	pRepo, _, db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().UTC()

	// 1. Create parent project first
	proj := &domain.Project{
		ID:         "proj-for-template-1",
		Name:       "Proyek Template Test",
		CreatedAt:  now,
		UpdatedAt:  now,
		IsArchived: false,
	}
	if err := pRepo.Create(ctx, proj); err != nil {
		t.Fatalf("pRepo.Create failed: %v", err)
	}

	// 2. Create Template
	tRepo := sqlite.NewTemplateRepository(db)
	doc := domain.NewDefaultA4LandscapeDocument()
	docJSON, _ := doc.ToJSON()

	tmpl := &domain.Template{
		ID:            "tmpl-1",
		ProjectID:     proj.ID,
		Name:          "Template Sertifikat Utama",
		DocumentJSON:  docJSON,
		SchemaVersion: 1,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := tRepo.Create(ctx, tmpl); err != nil {
		t.Fatalf("tRepo.Create failed: %v", err)
	}

	// 3. GetByID
	fetched, err := tRepo.GetByID(ctx, tmpl.ID)
	if err != nil {
		t.Fatalf("tRepo.GetByID failed: %v", err)
	}
	if fetched.Name != tmpl.Name || fetched.ProjectID != tmpl.ProjectID {
		t.Errorf("Fetched template mismatch: %+v vs %+v", fetched, tmpl)
	}

	// 4. ListByProject
	list, err := tRepo.ListByProject(ctx, proj.ID)
	if err != nil {
		t.Fatalf("tRepo.ListByProject failed: %v", err)
	}
	if len(list) != 1 || list[0].ID != tmpl.ID {
		t.Errorf("ListByProject returned unexpected slice: %+v", list)
	}

	// 5. Update
	tmpl.Name = "Template Sertifikat Utama (Revisi)"
	tmpl.UpdatedAt = time.Now().UTC()
	if err := tRepo.Update(ctx, tmpl); err != nil {
		t.Fatalf("tRepo.Update failed: %v", err)
	}

	updated, err := tRepo.GetByID(ctx, tmpl.ID)
	if err != nil {
		t.Fatalf("tRepo.GetByID after update failed: %v", err)
	}
	if updated.Name != "Template Sertifikat Utama (Revisi)" {
		t.Errorf("Expected updated template name, got %s", updated.Name)
	}

	// 6. Delete
	if err := tRepo.Delete(ctx, tmpl.ID); err != nil {
		t.Fatalf("tRepo.Delete failed: %v", err)
	}

	_, err = tRepo.GetByID(ctx, tmpl.ID)
	if err != domain.ErrTemplateNotFound {
		t.Errorf("Expected ErrTemplateNotFound after delete, got %v", err)
	}
}
