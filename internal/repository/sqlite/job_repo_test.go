package sqlite_test

import (
	"context"
	"testing"
	"time"

	"certificate-generator/internal/domain"
	"certificate-generator/internal/repository/sqlite"
)

func TestJobRepositoryCRUDAndTransitions(t *testing.T) {
	pRepo, _, db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().UTC()

	// 1. Create project, template, and recipients
	proj := &domain.Project{
		ID:        "proj-job-1",
		Name:      "Proyek Batch Job",
		CreatedAt: now,
		UpdatedAt: now,
	}
	_ = pRepo.Create(ctx, proj)

	tRepo := sqlite.NewTemplateRepository(db)
	doc := domain.NewDefaultA4LandscapeDocument()
	docJSON, _ := doc.ToJSON()
	tmpl := &domain.Template{
		ID:           "tmpl-job-1",
		ProjectID:    proj.ID,
		Name:         "Template Batch",
		DocumentJSON: docJSON,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	_ = tRepo.Create(ctx, tmpl)

	rRepo := sqlite.NewRecipientRepository(db)
	_ = rRepo.Create(ctx, &domain.Recipient{
		ID:        "rec-1",
		ProjectID: proj.ID,
		FullName:  "Budi Santoso",
	})
	_ = rRepo.Create(ctx, &domain.Recipient{
		ID:        "rec-2",
		ProjectID: proj.ID,
		FullName:  "Siti Aminah",
	})

	jRepo := sqlite.NewJobRepository(db)

	// 2. Create Job
	job := &domain.GenerationJob{
		ID:             "job-101",
		ProjectID:      proj.ID,
		TemplateID:     tmpl.ID,
		Status:         domain.JobStatusPending,
		TotalCount:     10,
		CompletedCount: 0,
		FailedCount:    0,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := jRepo.CreateJob(ctx, job); err != nil {
		t.Fatalf("CreateJob failed: %v", err)
	}

	// 3. GetJobByID
	fetched, err := jRepo.GetJobByID(ctx, job.ID)
	if err != nil {
		t.Fatalf("GetJobByID failed: %v", err)
	}
	if fetched.Status != domain.JobStatusPending || fetched.TotalCount != 10 {
		t.Errorf("Fetched job mismatch: %+v", fetched)
	}

	// 4. State Machine Transition: pending -> validating -> running
	if err := jRepo.UpdateJobStatus(ctx, job.ID, domain.JobStatusValidating, ""); err != nil {
		t.Fatalf("Update to validating failed: %v", err)
	}
	if err := jRepo.UpdateJobStatus(ctx, job.ID, domain.JobStatusRunning, ""); err != nil {
		t.Fatalf("Update to running failed: %v", err)
	}

	// Invalid transition: running -> pending (must be rejected)
	if err := jRepo.UpdateJobStatus(ctx, job.ID, domain.JobStatusPending, ""); err == nil {
		t.Errorf("Expected invalid state transition to be rejected, got nil")
	}

	// 5. Update Progress
	if err := jRepo.UpdateJobProgress(ctx, job.ID, 8, 2); err != nil {
		t.Fatalf("UpdateJobProgress failed: %v", err)
	}
	pJob, _ := jRepo.GetJobByID(ctx, job.ID)
	if pJob.CompletedCount != 8 || pJob.FailedCount != 2 {
		t.Errorf("Progress counts mismatch: completed=%d, failed=%d", pJob.CompletedCount, pJob.FailedCount)
	}

	// 6. Create Job Items
	items := []domain.GenerationJobItem{
		{
			ID:            "item-1",
			JobID:         job.ID,
			RecipientID:   "rec-1",
			RecipientName: "Budi Santoso",
			CertificateID: "CERT-001",
			Status:        domain.ItemStatusCompleted,
		},
		{
			ID:            "item-2",
			JobID:         job.ID,
			RecipientID:   "rec-2",
			RecipientName: "Siti Aminah",
			CertificateID: "CERT-002",
			Status:        domain.ItemStatusFailed,
			ErrorMessage:  "Corrupt font measurement",
		},
	}
	if err := jRepo.CreateJobItems(ctx, items); err != nil {
		t.Fatalf("CreateJobItems failed: %v", err)
	}

	// Query items
	allItems, err := jRepo.ListJobItems(ctx, job.ID)
	if err != nil || len(allItems) != 2 {
		t.Fatalf("ListJobItems failed: %v (len=%d)", err, len(allItems))
	}

	failedItems, err := jRepo.ListFailedJobItems(ctx, job.ID)
	if err != nil || len(failedItems) != 1 || failedItems[0].ID != "item-2" {
		t.Fatalf("ListFailedJobItems failed: %v (len=%d)", err, len(failedItems))
	}

	// 7. Finalize Job
	if err := jRepo.FinalizeJob(ctx, job.ID, domain.JobStatusCompletedWithErrors, "/exports/certs.zip", ""); err != nil {
		t.Fatalf("FinalizeJob failed: %v", err)
	}

	finalJob, _ := jRepo.GetJobByID(ctx, job.ID)
	if finalJob.Status != domain.JobStatusCompletedWithErrors || finalJob.OutputPath != "/exports/certs.zip" {
		t.Errorf("Final job mismatch: %+v", finalJob)
	}
}
