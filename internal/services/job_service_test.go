package services_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"certificate-generator/internal/domain"
	certgen "certificate-generator/internal/generator/certificate"
	"certificate-generator/internal/renderer/pdf"
	"certificate-generator/internal/repository/sqlite"
	"certificate-generator/internal/services"
	"certificate-generator/internal/storage/filesystem"
)

func TestJobServiceInterruptedGenerationRecovery(t *testing.T) {
	tempDir := t.TempDir()
	dataDir := filepath.Join(tempDir, "CertificateGenerator")
	fontDir := filepath.Join("..", "..", "assets", "fonts")

	sm, err := filesystem.NewStorageManager(dataDir)
	if err != nil {
		t.Fatalf("NewStorageManager failed: %v", err)
	}

	dbPath := filepath.Join(sm.DatabaseDir(), "app.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("sqlite.Open failed: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := sqlite.Migrate(ctx, db); err != nil {
		t.Fatalf("sqlite.Migrate failed: %v", err)
	}

	pRepo := sqlite.NewProjectRepository(db)
	tRepo := sqlite.NewTemplateRepository(db)
	rRepo := sqlite.NewRecipientRepository(db)
	jRepo := sqlite.NewJobRepository(db)
	renderer := pdf.NewGoPDFRenderer(fontDir)

	// Create project and template
	now := time.Now().UTC()
	proj := &domain.Project{
		ID:        "proj-job-recovery",
		Name:      "Proyek Uji Pemulihan",
		CreatedAt: now,
		UpdatedAt: now,
	}
	_ = pRepo.Create(ctx, proj)

	doc := domain.NewDefaultA4LandscapeDocument()
	docJSON, _ := doc.ToJSON()
	tmpl := &domain.Template{
		ID:           "tmpl-job-recovery",
		ProjectID:    proj.ID,
		Name:         "Template Recovery",
		DocumentJSON: docJSON,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	_ = tRepo.Create(ctx, tmpl)

	// Simulate an interrupted job left in 'running' status (e.g. power loss or process kill)
	staleJob := &domain.GenerationJob{
		ID:             "job-stale-1",
		ProjectID:      proj.ID,
		TemplateID:     tmpl.ID,
		Status:         domain.JobStatusRunning,
		TotalCount:     50,
		CompletedCount: 20,
		FailedCount:    0,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := jRepo.CreateJob(ctx, staleJob); err != nil {
		t.Fatalf("CreateJob failed: %v", err)
	}

	// Initialize generator and JobService
	generator := certgen.NewBatchGenerator(renderer, jRepo, rRepo, tRepo, pRepo, sm, nil)
	jobSvc := services.NewJobService(generator, jRepo, sm, nil)

	// Check and recover interrupted jobs
	interrupted, err := jobSvc.CheckInterruptedJobs(ctx)
	if err != nil {
		t.Fatalf("CheckInterruptedJobs failed: %v", err)
	}
	if len(interrupted) != 1 || interrupted[0].ID != staleJob.ID {
		t.Errorf("Expected 1 interrupted job discovered, got %d", len(interrupted))
	}

	// Verify that the job was finalized cleanly as 'failed' with reason
	recoveredJob, err := jobSvc.GetJob(ctx, staleJob.ID)
	if err != nil {
		t.Fatalf("GetJob failed: %v", err)
	}
	if recoveredJob.Status != domain.JobStatusFailed {
		t.Errorf("Expected recovered job status to be 'failed', got '%s'", recoveredJob.Status)
	}
	if recoveredJob.ErrorMessage == "" {
		t.Errorf("Expected explanatory error message on recovered job, got empty string")
	}

	// A second check should find 0 interrupted jobs (clean recovery)
	secondCheck, err := jobSvc.CheckInterruptedJobs(ctx)
	if err != nil {
		t.Fatalf("Second CheckInterruptedJobs failed: %v", err)
	}
	if len(secondCheck) != 0 {
		t.Errorf("Expected 0 interrupted jobs after recovery, found %d", len(secondCheck))
	}
}
