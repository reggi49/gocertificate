package app_test

import (
	"path/filepath"
	"testing"

	"certificate-generator/app"
	"certificate-generator/internal/domain"
)

func TestAppLifecycleAndPersistence(t *testing.T) {
	tempDir := t.TempDir()
	dataDir := filepath.Join(tempDir, "CertificateGenerator")
	fontDir := filepath.Join("..", "assets", "fonts")

	// Phase 1: Initialize App instance 1
	app1, err := app.NewApp(app.Config{
		DataDir: dataDir,
		FontDir: fontDir,
	})
	if err != nil {
		t.Fatalf("Failed to initialize app1: %v", err)
	}

	// 1. Create Project
	created, err := app1.CreateProject(domain.CreateProjectInput{
		Name:        "Wisuda Sarjana 2026",
		Description: "Fakultas Teknik Informatika",
	})
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}
	if created.ID == "" || created.Name != "Wisuda Sarjana 2026" {
		t.Errorf("Unexpected created project: %+v", created)
	}

	// 2. Rename Project
	renamed, err := app1.RenameProject(created.ID, "Wisuda Sarjana 2026 (Revisi)", "Deskripsi Revisi")
	if err != nil {
		t.Fatalf("RenameProject failed: %v", err)
	}
	if renamed.Name != "Wisuda Sarjana 2026 (Revisi)" {
		t.Errorf("Renamed project name mismatch: %s", renamed.Name)
	}

	// 3. Duplicate Project
	duplicated, err := app1.DuplicateProject(domain.DuplicateProjectInput{
		SourceProjectID: created.ID,
		NewName:         "Wisuda Magister 2026",
	})
	if err != nil {
		t.Fatalf("DuplicateProject failed: %v", err)
	}
	if duplicated.Name != "Wisuda Magister 2026" {
		t.Errorf("Duplicated project name mismatch: %s", duplicated.Name)
	}

	// 4. Archive Project
	if err := app1.ArchiveProject(created.ID, true); err != nil {
		t.Fatalf("ArchiveProject failed: %v", err)
	}

	// 5. Update Settings
	if err := app1.UpdateSettings(domain.AppSettings{
		DefaultOutputDir:    "/exports/cert",
		Theme:               "dark",
		Language:            "en",
		AutoSaveIntervalSec: 15,
	}); err != nil {
		t.Fatalf("UpdateSettings failed: %v", err)
	}

	// 6. Check Storage Info
	storageInfo, err := app1.GetStorageInfo()
	if err != nil {
		t.Fatalf("GetStorageInfo failed: %v", err)
	}
	if storageInfo.TotalProjects != 2 || storageInfo.ActiveProjects != 1 || storageInfo.ArchivedProjects != 1 {
		t.Errorf("StorageInfo mismatch: %+v", storageInfo)
	}

	// Gracefully shutdown App 1
	if err := app1.Close(); err != nil {
		t.Fatalf("app1.Close failed: %v", err)
	}

	// ==========================================
	// Phase 2: Restart Persistence Verification
	// Open App instance 2 with the same DataDir
	// ==========================================
	app2, err := app.NewApp(app.Config{
		DataDir: dataDir,
		FontDir: fontDir,
	})
	if err != nil {
		t.Fatalf("Failed to restart app2 with existing data dir: %v", err)
	}
	defer app2.Close()

	// Verify projects retained in SQLite & filesystem
	projects, err := app2.ListProjects(true)
	if err != nil {
		t.Fatalf("ListProjects after restart failed: %v", err)
	}
	if len(projects) != 2 {
		t.Fatalf("Expected 2 projects retained after restart, found %d", len(projects))
	}

	// Verify project details
	proj, err := app2.GetProject(duplicated.ID)
	if err != nil {
		t.Fatalf("GetProject after restart failed: %v", err)
	}
	if proj.Name != "Wisuda Magister 2026" {
		t.Errorf("Persisted project name mismatch: %s", proj.Name)
	}

	// Verify settings retained
	settings, err := app2.GetSettings()
	if err != nil {
		t.Fatalf("GetSettings after restart failed: %v", err)
	}
	if settings.Theme != "dark" || settings.Language != "en" || settings.AutoSaveIntervalSec != 15 || settings.DefaultOutputDir != "/exports/cert" {
		t.Errorf("Persisted settings mismatch after restart: %+v", settings)
	}

	// 7. Delete duplicated project
	if err := app2.DeleteProject(duplicated.ID); err != nil {
		t.Fatalf("DeleteProject failed: %v", err)
	}

	remaining, err := app2.ListProjects(true)
	if err != nil {
		t.Fatalf("ListProjects after delete failed: %v", err)
	}
	if len(remaining) != 1 {
		t.Errorf("Expected 1 project remaining after delete, found %d", len(remaining))
	}
}

func TestAppTemplatePersistenceAcrossRestarts(t *testing.T) {
	tempDir := t.TempDir()
	dataDir := filepath.Join(tempDir, "CertificateGenerator")
	fontDir := filepath.Join("..", "assets", "fonts")

	// Phase 1: Initialize App instance 1
	app1, err := app.NewApp(app.Config{
		DataDir: dataDir,
		FontDir: fontDir,
	})
	if err != nil {
		t.Fatalf("Failed to initialize app1: %v", err)
	}

	// Create project
	proj, err := app1.CreateProject(domain.CreateProjectInput{
		Name: "Proyek Wisuda Template Test",
	})
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	// Create Template
	tmpl, err := app1.CreateTemplate(domain.CreateTemplateInput{
		ProjectID: proj.ID,
		Name:      "Template Desain Utama",
	})
	if err != nil {
		t.Fatalf("CreateTemplate failed: %v", err)
	}

	// Add elements to template document
	doc, err := domain.DocumentFromJSON(tmpl.DocumentJSON)
	if err != nil {
		t.Fatalf("DocumentFromJSON failed: %v", err)
	}
	doc.Elements = append(doc.Elements, domain.Element{
		ID:          "text-recipient",
		Type:        domain.ElementTypeText,
		XMm:         30,
		YMm:         85,
		WidthMm:     237,
		HeightMm:    25,
		RotationDeg: 0,
		Opacity:     1.0,
		Visible:     true,
		Properties: domain.ElementProperties{
			Text:       "{{recipient_name}}",
			FontFamily: "LiberationSerif",
			FontSizePt: 30,
			FitMode:    "shrink",
		},
	})
	updatedDocJSON, err := doc.ToJSON()
	if err != nil {
		t.Fatalf("doc.ToJSON failed: %v", err)
	}

	updatedTmpl, err := app1.UpdateTemplate(tmpl.ID, "Template Desain Utama Revisi", updatedDocJSON)
	if err != nil {
		t.Fatalf("UpdateTemplate failed: %v", err)
	}
	if updatedTmpl.Name != "Template Desain Utama Revisi" {
		t.Errorf("Updated template name mismatch: %s", updatedTmpl.Name)
	}

	// Close App 1
	if err := app1.Close(); err != nil {
		t.Fatalf("app1.Close failed: %v", err)
	}

	// Phase 2: Restart App with same dataDir
	app2, err := app.NewApp(app.Config{
		DataDir: dataDir,
		FontDir: fontDir,
	})
	if err != nil {
		t.Fatalf("Failed to initialize app2: %v", err)
	}
	defer app2.Close()

	// List templates for project
	templates, err := app2.ListTemplates(proj.ID)
	if err != nil {
		t.Fatalf("ListTemplates failed: %v", err)
	}
	if len(templates) != 1 {
		t.Fatalf("Expected 1 template, got %d", len(templates))
	}

	// Verify document elements retained
	reopenedTmpl, err := app2.GetTemplate(tmpl.ID)
	if err != nil {
		t.Fatalf("GetTemplate failed: %v", err)
	}
	if reopenedTmpl.Name != "Template Desain Utama Revisi" {
		t.Errorf("Reopened template name mismatch: %s", reopenedTmpl.Name)
	}

	reopenedDoc, err := domain.DocumentFromJSON(reopenedTmpl.DocumentJSON)
	if err != nil {
		t.Fatalf("Reopened document validation failed: %v", err)
	}
	if len(reopenedDoc.Elements) != 1 || reopenedDoc.Elements[0].ID != "text-recipient" {
		t.Errorf("Reopened document elements mismatch: %+v", reopenedDoc.Elements)
	}
	if reopenedDoc.Elements[0].Properties.Text != "{{recipient_name}}" {
		t.Errorf("Element text mismatch: %s", reopenedDoc.Elements[0].Properties.Text)
	}

	// Delete template
	if err := app2.DeleteTemplate(tmpl.ID); err != nil {
		t.Fatalf("DeleteTemplate failed: %v", err)
	}
	afterDelete, err := app2.ListTemplates(proj.ID)
	if err != nil {
		t.Fatalf("ListTemplates after delete failed: %v", err)
	}
	if len(afterDelete) != 0 {
		t.Errorf("Expected 0 templates after delete, got %d", len(afterDelete))
	}
}

func TestAppBatchGenerationWorkflow(t *testing.T) {
	tempDir := t.TempDir()
	dataDir := filepath.Join(tempDir, "CertificateGenerator")
	fontDir := filepath.Join("..", "assets", "fonts")

	appInstance, err := app.NewApp(app.Config{
		DataDir: dataDir,
		FontDir: fontDir,
	})
	if err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}
	defer appInstance.Close()

	// 1. Create project
	proj, err := appInstance.CreateProject(domain.CreateProjectInput{
		Name: "Proyek Batch Test App",
	})
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	// 2. Create template
	tmpl, err := appInstance.CreateTemplate(domain.CreateTemplateInput{
		ProjectID: proj.ID,
		Name:      "Template Sertifikat App",
	})
	if err != nil {
		t.Fatalf("CreateTemplate failed: %v", err)
	}

	// 3. Create recipients
	_, _ = appInstance.CreateRecipient(domain.Recipient{
		ProjectID:     proj.ID,
		FullName:      "Budi Santoso",
		CertificateID: "CERT-001",
	})
	_, _ = appInstance.CreateRecipient(domain.Recipient{
		ProjectID:     proj.ID,
		FullName:      "Siti Aminah",
		CertificateID: "CERT-002",
	})

	// 4. Run Batch Generation
	cfg := domain.BatchConfig{
		ProjectID:   proj.ID,
		TemplateID:  tmpl.ID,
		Concurrency: 1,
		ExportZip:   true,
	}

	job, err := appInstance.StartBatchGeneration(cfg)
	if err != nil {
		t.Fatalf("StartBatchGeneration failed: %v", err)
	}
	if job.Status != domain.JobStatusCompleted || job.CompletedCount != 2 {
		t.Errorf("Unexpected batch result: %+v", job)
	}

	// 5. Query batch jobs & items
	jobs, err := appInstance.ListBatchJobs(proj.ID)
	if err != nil || len(jobs) != 1 {
		t.Errorf("ListBatchJobs mismatch: len=%d, err=%v", len(jobs), err)
	}

	items, err := appInstance.ListBatchJobItems(job.ID)
	if err != nil || len(items) != 2 {
		t.Errorf("ListBatchJobItems mismatch: len=%d, err=%v", len(items), err)
	}
}

func TestAppRenderPreviewPDF(t *testing.T) {
	tempDir := t.TempDir()
	appInstance, err := app.NewApp(app.Config{
		DataDir: filepath.Join(tempDir, "CertificateGenerator"),
		FontDir: filepath.Join("..", "assets", "fonts"),
	})
	if err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}
	defer appInstance.Close()

	doc := domain.Document{
		SchemaVersion: 1,
		Page: domain.PageSettings{
			WidthMm:         297,
			HeightMm:        210,
			Orientation:     "landscape",
			BackgroundColor: "#FFFFFF",
		},
		Elements: []domain.Element{
			{
				ID:       "txt-1",
				Type:     domain.ElementTypeText,
				XMm:      20,
				YMm:      20,
				WidthMm:  100,
				HeightMm: 15,
				Visible:  true,
				Properties: domain.ElementProperties{
					Text:       "Hello {{name}}",
					FontSizePt: 16,
				},
			},
		},
	}

	docJSON, err := doc.ToJSON()
	if err != nil {
		t.Fatalf("doc.ToJSON failed: %v", err)
	}

	pdfBytes, err := appInstance.RenderPreviewPDF(docJSON, map[string]string{
		"name": "Dunia",
	})
	if err != nil {
		t.Fatalf("RenderPreviewPDF failed: %v", err)
	}
	if len(pdfBytes) < 500 {
		t.Errorf("PDF output unexpectedly small: %d bytes", len(pdfBytes))
	}
}
