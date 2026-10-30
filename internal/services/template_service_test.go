package services_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"certificate-generator/internal/domain"
	"certificate-generator/internal/repository/sqlite"
	"certificate-generator/internal/services"
)

func TestTemplateServiceCRUD(t *testing.T) {
	projSvc, pRepo, sm, _, cleanup := setupServiceTest(t)
	defer cleanup()

	ctx := context.Background()

	// 1. Create a parent project
	proj, err := projSvc.CreateProject(ctx, domain.CreateProjectInput{
		Name: "Proyek Wisuda 2026",
	})
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	dbPath := filepath.Join(sm.DatabaseDir(), "app.db")
	db, _ := sqlite.Open(dbPath)
	defer db.Close()
	tRepo := sqlite.NewTemplateRepository(db)
	tmplSvc := services.NewTemplateService(tRepo, pRepo, sm, nil)

	// 2. Create Template with default document
	created, err := tmplSvc.CreateTemplate(ctx, domain.CreateTemplateInput{
		ProjectID: proj.ID,
		Name:      "Template Landscape Default",
	})
	if err != nil {
		t.Fatalf("CreateTemplate failed: %v", err)
	}
	if created.ID == "" || created.Name != "Template Landscape Default" {
		t.Errorf("Unexpected created template: %+v", created)
	}

	// Verify template file synced on disk
	projDir, _ := sm.ProjectDir(proj.ID)
	templateFilePath := filepath.Join(projDir, "templates", created.ID+".json")
	if info, err := os.Stat(templateFilePath); err != nil || info.Size() == 0 {
		t.Errorf("Expected template file to be created on disk at %s", templateFilePath)
	}

	// 3. Get Template
	fetched, err := tmplSvc.GetTemplate(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetTemplate failed: %v", err)
	}
	if fetched.Name != created.Name {
		t.Errorf("Fetched template mismatch: %s vs %s", fetched.Name, created.Name)
	}

	// 4. Update Template with added element
	doc, err := domain.DocumentFromJSON(created.DocumentJSON)
	if err != nil {
		t.Fatalf("Failed to parse doc JSON: %v", err)
	}
	doc.Elements = append(doc.Elements, domain.Element{
		ID:       "el-title",
		Type:     domain.ElementTypeText,
		XMm:      50,
		YMm:      50,
		WidthMm:  150,
		HeightMm: 20,
		Visible:  true,
		Properties: domain.ElementProperties{
			Text:       "Sertifikat Kelulusan",
			FontSizePt: 24,
		},
	})
	updatedDocJSON, _ := doc.ToJSON()

	updated, err := tmplSvc.UpdateTemplate(ctx, created.ID, "Template Landscape Berisi Elemen", updatedDocJSON)
	if err != nil {
		t.Fatalf("UpdateTemplate failed: %v", err)
	}
	if updated.Name != "Template Landscape Berisi Elemen" {
		t.Errorf("Updated template name mismatch: %s", updated.Name)
	}

	// Verify updated file on disk
	diskContent, err := os.ReadFile(templateFilePath)
	if err != nil {
		t.Fatalf("Failed to read updated template on disk: %v", err)
	}
	if string(diskContent) != updatedDocJSON {
		t.Errorf("Disk file content was not updated with new document JSON")
	}

	// 5. List Templates
	list, err := tmplSvc.ListTemplates(ctx, proj.ID)
	if err != nil {
		t.Fatalf("ListTemplates failed: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("Expected 1 template in list, got %d", len(list))
	}

	// 6. Delete Template
	if err := tmplSvc.DeleteTemplate(ctx, created.ID); err != nil {
		t.Fatalf("DeleteTemplate failed: %v", err)
	}

	_, err = tmplSvc.GetTemplate(ctx, created.ID)
	if err != domain.ErrTemplateNotFound {
		t.Errorf("Expected ErrTemplateNotFound after delete, got %v", err)
	}

	// Verify file deleted from disk
	if _, err := os.Stat(templateFilePath); !os.IsNotExist(err) {
		t.Errorf("Expected template file to be removed from disk, but still exists")
	}
}
