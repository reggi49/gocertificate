package certificate_test

import (
	"archive/zip"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"certificate-generator/internal/domain"
	certgen "certificate-generator/internal/generator/certificate"
	"certificate-generator/internal/renderer/pdf"
	"certificate-generator/internal/repository/sqlite"
	"certificate-generator/internal/storage/filesystem"
)

func setupBatchTestEnv(t *testing.T, count int) (
	*certgen.BatchGenerator,
	*domain.Project,
	*domain.Template,
	[]domain.Recipient,
	*filesystem.StorageManager,
	func(),
) {
	t.Helper()
	tempDir := t.TempDir()
	dataDir := filepath.Join(tempDir, "CertificateGenerator")
	fontDir := filepath.Join("..", "..", "..", "assets", "fonts")

	sm, err := filesystem.NewStorageManager(dataDir)
	if err != nil {
		t.Fatalf("NewStorageManager failed: %v", err)
	}

	dbPath := filepath.Join(sm.DatabaseDir(), "app.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("sqlite.Open failed: %v", err)
	}

	ctx := context.Background()
	if err := sqlite.Migrate(ctx, db); err != nil {
		t.Fatalf("sqlite.Migrate failed: %v", err)
	}

	pRepo := sqlite.NewProjectRepository(db)
	tRepo := sqlite.NewTemplateRepository(db)
	rRepo := sqlite.NewRecipientRepository(db)
	jRepo := sqlite.NewJobRepository(db)
	renderer := pdf.NewGoPDFRenderer(fontDir)

	generator := certgen.NewBatchGenerator(renderer, jRepo, rRepo, tRepo, pRepo, sm, nil)

	// Create Project & Project Folder
	projID := "proj-batch-1"
	_, _ = sm.CreateProjectDirectory(projID)
	proj := &domain.Project{
		ID:        projID,
		Name:      "Proyek Batch Uji",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	_ = pRepo.Create(ctx, proj)

	// Create Template
	doc := domain.NewDefaultA4LandscapeDocument()
	doc.Elements = append(doc.Elements, domain.Element{
		ID:       "txt-name",
		Type:     domain.ElementTypeText,
		XMm:      40,
		YMm:      80,
		WidthMm:  217,
		HeightMm: 25,
		Visible:  true,
		Properties: domain.ElementProperties{
			Text:       "{{recipient_name}}",
			FontFamily: "LiberationSerif",
			FontSizePt: 28,
			FitMode:    "shrink",
			Alignment:  "center",
		},
	})
	doc.Elements = append(doc.Elements, domain.Element{
		ID:       "qr-cert",
		Type:     domain.ElementTypeQRCode,
		XMm:      30,
		YMm:      140,
		WidthMm:  30,
		HeightMm: 30,
		Visible:  true,
		Properties: domain.ElementProperties{
			QRContent: "https://cert.local/verify?id={{certificate_id}}",
		},
	})
	docJSON, _ := doc.ToJSON()

	tmpl := &domain.Template{
		ID:           "tmpl-batch-1",
		ProjectID:    proj.ID,
		Name:         "Template Desain Batch",
		DocumentJSON: docJSON,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	_ = tRepo.Create(ctx, tmpl)

	// Create Recipients
	recipients := make([]domain.Recipient, count)
	for i := 0; i < count; i++ {
		recipients[i] = domain.Recipient{
			ID:            fmt.Sprintf("rec-%05d", i+1),
			ProjectID:     proj.ID,
			FullName:      fmt.Sprintf("Peserta Ke-%05d, S.Kom.", i+1),
			Email:         fmt.Sprintf("peserta%05d@mail.com", i+1),
			Organization:  "Universitas Indonesia",
			EventName:     "Konferensi Sains & AI 2026",
			IssueDate:     "2026-10-30",
			CertificateID: fmt.Sprintf("CERT-2026-%05d", i+1),
		}
	}
	_, _ = rRepo.BatchCreate(ctx, proj.ID, recipients, domain.ConflictStrategyFail)

	cleanup := func() {
		_ = db.Close()
	}

	return generator, proj, tmpl, recipients, sm, cleanup
}

func TestBatchGenerationSingleCertificate(t *testing.T) {
	generator, proj, tmpl, _, _, cleanup := setupBatchTestEnv(t, 1)
	defer cleanup()

	ctx := context.Background()
	cfg := domain.BatchConfig{
		ProjectID:   proj.ID,
		TemplateID:  tmpl.ID,
		Concurrency: 1,
		ExportZip:   true,
	}

	var events []domain.ProgressEvent
	job, err := generator.ExecuteBatch(ctx, cfg, func(ev domain.ProgressEvent) {
		events = append(events, ev)
	})

	if err != nil {
		t.Fatalf("ExecuteBatch failed: %v", err)
	}

	if job.Status != domain.JobStatusCompleted || job.CompletedCount != 1 || job.FailedCount != 0 {
		t.Errorf("Unexpected job status: %+v", job)
	}

	if len(events) == 0 {
		t.Errorf("Expected progress events, got 0")
	}

	// Verify ZIP exists and contains files
	if _, err := os.Stat(job.OutputPath); err != nil {
		t.Fatalf("ZIP output file not found: %s", job.OutputPath)
	}

	zr, err := zip.OpenReader(job.OutputPath)
	if err != nil {
		t.Fatalf("Failed to open generated ZIP: %v", err)
	}
	defer zr.Close()

	if len(zr.File) != 2 { // 1 PDF + 1 export-summary.csv
		t.Errorf("Expected 2 files in ZIP, found %d", len(zr.File))
	}
}

func TestBatchGeneration100CertificatesAndZipIntegrity(t *testing.T) {
	generator, proj, tmpl, _, _, cleanup := setupBatchTestEnv(t, 100)
	defer cleanup()

	ctx := context.Background()
	cfg := domain.BatchConfig{
		ProjectID:   proj.ID,
		TemplateID:  tmpl.ID,
		Concurrency: 2, // Concurrency 2
		ExportZip:   true,
	}

	start := time.Now()
	var lastPercent atomic.Int32
	job, err := generator.ExecuteBatch(ctx, cfg, func(ev domain.ProgressEvent) {
		lastPercent.Store(int32(ev.Percentage))
	})
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("ExecuteBatch 100 failed: %v", err)
	}

	if job.Status != domain.JobStatusCompleted || job.CompletedCount != 100 || job.FailedCount != 0 {
		t.Errorf("100-batch job mismatch: completed=%d, failed=%d, status=%s",
			job.CompletedCount, job.FailedCount, job.Status)
	}

	if lastPercent.Load() != 100 {
		t.Errorf("Expected final progress 100%%, got %d%%", lastPercent.Load())
	}

	t.Logf("Generated 100 certificates in %v (avg %.2f ms/cert)", duration, float64(duration.Milliseconds())/100.0)

	// Verify ZIP integrity
	zr, err := zip.OpenReader(job.OutputPath)
	if err != nil {
		t.Fatalf("Generated ZIP corrupted: %v", err)
	}
	defer zr.Close()

	// 100 PDFs + 1 export-summary.csv = 101 entries
	if len(zr.File) != 101 {
		t.Errorf("Expected 101 entries in ZIP archive, found %d", len(zr.File))
	}

	// Verify export-summary.csv is readable
	hasSummary := false
	for _, f := range zr.File {
		if f.Name == "export-summary.csv" {
			hasSummary = true
			break
		}
	}
	if !hasSummary {
		t.Errorf("export-summary.csv missing in ZIP archive")
	}
}

func TestBatchGenerationCancellation(t *testing.T) {
	generator, proj, tmpl, _, _, cleanup := setupBatchTestEnv(t, 50)
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	cfg := domain.BatchConfig{
		ProjectID:   proj.ID,
		TemplateID:  tmpl.ID,
		Concurrency: 1,
		ExportZip:   true,
	}

	job, err := generator.ExecuteBatch(ctx, cfg, nil)
	if err != nil {
		t.Fatalf("ExecuteBatch should return job even on cancel: %v", err)
	}

	if job.Status != domain.JobStatusCancelled {
		t.Errorf("Expected job status 'cancelled', got %s", job.Status)
	}
}

func TestBatchGenerationRetryFailedRecords(t *testing.T) {
	generator, proj, tmpl, _, _, cleanup := setupBatchTestEnv(t, 5)
	defer cleanup()

	ctx := context.Background()

	// 1. Initial batch execution
	cfg := domain.BatchConfig{
		ProjectID:   proj.ID,
		TemplateID:  tmpl.ID,
		Concurrency: 1,
		ExportZip:   true,
	}

	job, err := generator.ExecuteBatch(ctx, cfg, nil)
	if err != nil {
		t.Fatalf("ExecuteBatch failed: %v", err)
	}

	// 2. Simulate one item failed in DB
	items, err := generator.GetJobRepo().ListJobItems(ctx, job.ID)
	if err != nil || len(items) == 0 {
		t.Fatalf("ListJobItems failed: %v", err)
	}
	failedItemID := items[0].ID
	_ = generator.GetJobRepo().UpdateJobItem(ctx, failedItemID, domain.ItemStatusFailed, "", "Simulated rendering failure")

	// 3. Retry failed items
	retryJob, err := generator.RetryFailedItems(ctx, job.ID, nil)
	if err != nil {
		t.Fatalf("RetryFailedItems failed: %v", err)
	}

	if retryJob.TotalCount != 1 || retryJob.Status != domain.JobStatusCompleted {
		t.Errorf("Retry job mismatch: total=%d, status=%s", retryJob.TotalCount, retryJob.Status)
	}
}

func TestBatchGeneration1000CertificatesBenchmark(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping 1000-certificate benchmark in short mode")
	}

	generator, proj, tmpl, _, _, cleanup := setupBatchTestEnv(t, 1000)
	defer cleanup()

	ctx := context.Background()
	cfg := domain.BatchConfig{
		ProjectID:   proj.ID,
		TemplateID:  tmpl.ID,
		Concurrency: 2,
		ExportZip:   true,
	}

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	startTime := time.Now()
	job, err := generator.ExecuteBatch(ctx, cfg, nil)
	duration := time.Since(startTime)

	runtime.ReadMemStats(&m2)
	allocMB := float64(m2.TotalAlloc-m1.TotalAlloc) / (1024 * 1024)

	if err != nil {
		t.Fatalf("ExecuteBatch 1000 failed: %v", err)
	}

	if job.Status != domain.JobStatusCompleted || job.CompletedCount != 1000 {
		t.Errorf("1000 batch mismatch: completed=%d, status=%s", job.CompletedCount, job.Status)
	}

	t.Logf("=== 1000 Certificate Benchmark ===")
	t.Logf("Total Duration: %v (avg %.2f ms/cert)", duration, float64(duration.Milliseconds())/1000.0)
	t.Logf("Memory Allocated: %.2f MB", allocMB)
	t.Logf("Output ZIP Size: %.2f MB", float64(getFileSize(job.OutputPath))/(1024*1024))
}

func getFileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

func TestBatchGenerationEventVariablesFromProject(t *testing.T) {
	generator, proj, tmpl, _, sm, cleanup := setupBatchTestEnv(t, 1)
	defer cleanup()

	ctx := context.Background()

	// Update project event details
	proj.EventDate = "2026-10-30"
	proj.EventLocation = "Yayasan Janaiz"
	proj.EventAddress = "Jl Kelapa Gading Blok C"
	_ = generator.GetJobRepo() // repo access

	// Template with event placeholders (including uppercase)
	tmplDoc := &domain.Document{
		SchemaVersion: 1,
		Page:          domain.PageSettings{WidthMm: 210, HeightMm: 297, Orientation: "portrait"},
		Elements: []domain.Element{
			{
				ID:   "el-1",
				Type: domain.ElementTypeText,
				Properties: domain.ElementProperties{
					Text: "Diberikan kepada {{recipient_name}}\nPada {{ISSUE_DATE}}\nDi {{EVENT_LOCATION}}\n{{event_address}}",
				},
			},
		},
	}
	tmplJSON, _ := tmplDoc.ToJSON()
	tmpl.DocumentJSON = tmplJSON

	cfg := domain.BatchConfig{
		ProjectID:           proj.ID,
		TemplateID:          tmpl.ID,
		Concurrency:         1,
		ExportZip:           false,
		ExportIndividualPDF: true,
	}

	job, err := generator.ExecuteBatch(ctx, cfg, nil)
	if err != nil {
		t.Fatalf("ExecuteBatch failed: %v", err)
	}
	if job.Status != domain.JobStatusCompleted {
		t.Fatalf("Expected completed job, got %s", job.Status)
	}

	// Verify the individual file was generated and exists
	projDir, _ := sm.ProjectDir(proj.ID)
	files, err := os.ReadDir(filepath.Join(projDir, "exports"))
	if err != nil || len(files) == 0 {
		t.Fatalf("Expected exports files to exist: %v", err)
	}
}
