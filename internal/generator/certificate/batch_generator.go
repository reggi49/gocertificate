package certificate

import (
	"archive/zip"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"certificate-generator/internal/domain"
	"certificate-generator/internal/renderer/pdf"
	"certificate-generator/internal/repository/sqlite"
	"certificate-generator/internal/storage/filesystem"

	"github.com/google/uuid"
)

// ProgressCallback receives real-time progress events.
type ProgressCallback func(event domain.ProgressEvent)

// BatchGenerator orchestrates batch PDF production with bounded concurrency, streaming ZIP, and memory limits.
type BatchGenerator struct {
	renderer  pdf.Renderer
	jobRepo   sqlite.JobRepository
	recRepo   sqlite.RecipientRepository
	tmplRepo  sqlite.TemplateRepository
	projRepo  sqlite.ProjectRepository
	sm        *filesystem.StorageManager
	logger    *slog.Logger
	cancels   map[string]context.CancelFunc
	cancelsMu sync.Mutex
}

// NewBatchGenerator constructs a BatchGenerator.
func NewBatchGenerator(
	renderer pdf.Renderer,
	jobRepo sqlite.JobRepository,
	recRepo sqlite.RecipientRepository,
	tmplRepo sqlite.TemplateRepository,
	projRepo sqlite.ProjectRepository,
	sm *filesystem.StorageManager,
	logger *slog.Logger,
) *BatchGenerator {
	if logger == nil {
		logger = slog.Default()
	}
	return &BatchGenerator{
		renderer: renderer,
		jobRepo:  jobRepo,
		recRepo:  recRepo,
		tmplRepo: tmplRepo,
		projRepo: projRepo,
		sm:       sm,
		logger:   logger,
		cancels:  make(map[string]context.CancelFunc),
	}
}

// formatIndonesianDate converts a date (e.g. "2026-10-30") into standard Indonesian format (e.g. "30 OKTOBER 2026").
func formatIndonesianDate(rawDate string) string {
	rawDate = strings.TrimSpace(rawDate)
	if rawDate == "" {
		return ""
	}

	formats := []string{
		"2006-01-02",
		"02-01-2006",
		"02/01/2006",
		"2006/01/02",
	}

	for _, layout := range formats {
		if t, err := time.Parse(layout, rawDate); err == nil {
			indonesianMonths := [...]string{
				"JANUARI", "FEBRUARI", "MARET", "APRIL", "MEI", "JUNI",
				"JULI", "AGUSTUS", "SEPTEMBER", "OKTOBER", "NOVEMBER", "DESEMBER",
			}
			monthIdx := int(t.Month()) - 1
			if monthIdx >= 0 && monthIdx < 12 {
				return fmt.Sprintf("%d %s %d", t.Day(), indonesianMonths[monthIdx], t.Year())
			}
		}
	}

	return rawDate
}

// GetJobRepo exposes job repository for testing.
func (g *BatchGenerator) GetJobRepo() sqlite.JobRepository {
	return g.jobRepo
}

// ExecuteBatch runs the batch generation pipeline with a bounded worker pool.
func (g *BatchGenerator) ExecuteBatch(
	ctx context.Context,
	cfg domain.BatchConfig,
	onProgress ProgressCallback,
) (*domain.GenerationJob, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// 1. Fetch template (use provided ID, or fallback to project's default template)
	var tmpl *domain.Template
	if cfg.TemplateID != "" {
		t, err := g.tmplRepo.GetByID(ctx, cfg.TemplateID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch template: %w", err)
		}
		tmpl = t
	} else {
		tmpls, err := g.tmplRepo.ListByProject(ctx, cfg.ProjectID)
		if err == nil && len(tmpls) > 0 {
			tmpl = &tmpls[0]
		} else {
			// Auto create default template for project
			now := time.Now().UTC()
			defaultDocJSON := domain.DefaultJanaizDocumentJSON()
			defaultTmpl := &domain.Template{
				ID:            uuid.New().String(),
				ProjectID:     cfg.ProjectID,
				Name:          "Template Sertifikat Janaiz",
				DocumentJSON:  defaultDocJSON,
				SchemaVersion: 1,
				CreatedAt:     now,
				UpdatedAt:     now,
			}
			_ = g.tmplRepo.Create(ctx, defaultTmpl)
			tmpl = defaultTmpl
		}
		cfg.TemplateID = tmpl.ID
	}

	doc, err := domain.DocumentFromJSON(tmpl.DocumentJSON)
	if err != nil {
		return nil, fmt.Errorf("invalid template document: %w", err)
	}

	// 1b. Fetch project for event defaults (date, venue, address, event name)
	var proj *domain.Project
	if g.projRepo != nil {
		p, err := g.projRepo.GetByID(ctx, cfg.ProjectID)
		if err == nil && p != nil {
			proj = p
		}
	}

	var projEventName, projEventDate, projLocation, projAddress string
	if proj != nil {
		projEventName = proj.Name
		projEventDate = formatIndonesianDate(proj.EventDate)
		projLocation = proj.EventLocation
		projAddress = proj.EventAddress
	}

	// 2. Fetch recipients
	var recipients []domain.Recipient
	if len(cfg.RecipientIDs) > 0 {
		// Specific list
		for _, rID := range cfg.RecipientIDs {
			rec, err := g.recRepo.GetByID(ctx, rID)
			if err == nil && rec != nil {
				recipients = append(recipients, *rec)
			}
		}
	} else {
		// All recipients in project
		recipients, err = g.recRepo.GetAllByProject(ctx, cfg.ProjectID)
		if err != nil {
			return nil, fmt.Errorf("failed to load project recipients: %w", err)
		}
	}

	if len(recipients) == 0 {
		return nil, domain.ErrEmptyRecipientsForBatch
	}

	// 3. Create Job in database
	jobID := uuid.New().String()
	now := time.Now().UTC()
	job := &domain.GenerationJob{
		ID:             jobID,
		ProjectID:      cfg.ProjectID,
		TemplateID:     cfg.TemplateID,
		Status:         domain.JobStatusPending,
		TotalCount:     len(recipients),
		CompletedCount: 0,
		FailedCount:    0,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := g.jobRepo.CreateJob(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to persist generation job: %w", err)
	}

	// 4. Create Job Items
	jobItems := make([]domain.GenerationJobItem, len(recipients))
	for i, rec := range recipients {
		jobItems[i] = domain.GenerationJobItem{
			ID:            uuid.New().String(),
			JobID:         jobID,
			RecipientID:   rec.ID,
			RecipientName: pdf.ToTitleCase(rec.FullName),
			CertificateID: rec.CertificateID,
			Status:        domain.ItemStatusPending,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
	}
	if err := g.jobRepo.CreateJobItems(ctx, jobItems); err != nil {
		_ = g.jobRepo.UpdateJobStatus(ctx, jobID, domain.JobStatusFailed, err.Error())
		return nil, fmt.Errorf("failed to initialize job items: %w", err)
	}

	// Prepare cancellable context for this job
	jobCtx, cancel := context.WithCancel(ctx)
	g.cancelsMu.Lock()
	g.cancels[jobID] = cancel
	g.cancelsMu.Unlock()
	defer func() {
		g.cancelsMu.Lock()
		delete(g.cancels, jobID)
		g.cancelsMu.Unlock()
	}()

	// 5. Setup temporary directory for isolation & guaranteed cleanup
	projDir, _ := g.sm.ProjectDir(cfg.ProjectID)
	tempBatchDir, err := os.MkdirTemp(os.TempDir(), fmt.Sprintf("batch-%s-*", jobID))
	if err != nil {
		_ = g.jobRepo.UpdateJobStatus(ctx, jobID, domain.JobStatusFailed, err.Error())
		return nil, fmt.Errorf("failed to create temporary working directory: %w", err)
	}
	defer func() {
		// Guaranteed cleanup of intermediate temporary directory
		_ = os.RemoveAll(tempBatchDir)
	}()

	// 6. Transition state: pending -> validating -> running
	if err := g.jobRepo.UpdateJobStatus(jobCtx, jobID, domain.JobStatusValidating, ""); err != nil {
		return nil, err
	}
	if err := g.jobRepo.UpdateJobStatus(jobCtx, jobID, domain.JobStatusRunning, ""); err != nil {
		return nil, err
	}

	// Output target
	finalOutputDir := strings.TrimSpace(cfg.OutputDir)
	if finalOutputDir == "" {
		finalOutputDir = filepath.Join(projDir, "exports")
	}
	_ = os.MkdirAll(finalOutputDir, 0755)

	assetRootDir := filepath.Join(projDir, "assets")

	// 7. Bounded Worker Pool Processing (Smart Auto Concurrency: 75% of machine CPU cores)
	concurrency := cfg.Concurrency
	if concurrency <= 0 {
		cpus := runtime.NumCPU()
		concurrency = int(float64(cpus) * 0.75)
		if concurrency < 1 {
			concurrency = 1
		}
	}

	type workTask struct {
		index int
		item  domain.GenerationJobItem
		rec   domain.Recipient
	}

	taskChan := make(chan workTask, len(recipients))
	for i, r := range recipients {
		taskChan <- workTask{index: i, item: jobItems[i], rec: r}
	}
	close(taskChan)

	var completedCounter int32
	var failedCounter int32
	var summaryRows [][]string
	var summaryMu sync.Mutex

	summaryRows = append(summaryRows, []string{
		"Recipient ID", "Certificate ID", "Full Name", "Output File", "Status", "Error Message",
	})

	var wg sync.WaitGroup
	isCancelled := false

	for w := 0; w < concurrency; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					g.logger.Error("Worker goroutine panic recovered", "error", r)
				}
			}()

			for task := range taskChan {
				// Check context cancellation before item processing
				select {
				case <-jobCtx.Done():
					isCancelled = true
					return
				default:
				}

				// Resolve variables map for recipient with project defaults
				eventDate := projEventDate
				if strings.TrimSpace(task.rec.IssueDate) != "" {
					eventDate = formatIndonesianDate(task.rec.IssueDate)
				}

				eventName := projEventName
				if strings.TrimSpace(task.rec.EventName) != "" {
					eventName = task.rec.EventName
				}

				// Auto-format recipient name to Title Case (e.g. "muhamad reggi" -> "Muhamad Reggi")
				recipientName := pdf.ToTitleCase(task.rec.FullName)

				// Auto-format event venue, address, and date to UPPERCASE
				locationUpper := strings.ToUpper(strings.TrimSpace(projLocation))
				addressUpper := strings.ToUpper(strings.TrimSpace(projAddress))
				eventDateUpper := strings.ToUpper(strings.TrimSpace(eventDate))
				eventNameUpper := strings.ToUpper(strings.TrimSpace(eventName))

				vars := map[string]string{
					"recipient_name": recipientName,
					"email":          task.rec.Email,
					"organization":   task.rec.Organization,
					"certificate_id": task.rec.CertificateID,
					"event_name":     eventNameUpper,
					"nama_acara":     eventNameUpper,
					"issue_date":     eventDateUpper,
					"event_date":     eventDateUpper,
					"tanggal":        eventDateUpper,
					"tanggal_acara":  eventDateUpper,
					"event_location": locationUpper,
					"location":       locationUpper,
					"tempat":         locationUpper,
					"lokasi":         locationUpper,
					"event_address":  addressUpper,
					"address":        addressUpper,
					"alamat":         addressUpper,
				}

				// Unmarshal custom fields if any (highest precedence)
				if task.rec.CustomFieldsJSON != "" {
					var cf map[string]string
					if err := json.Unmarshal([]byte(task.rec.CustomFieldsJSON), &cf); err == nil {
						for k, v := range cf {
							vars[k] = v
						}
					}
				}

				// Generate sanitized filename: {certificate_id}_{recipient_name}.pdf or {index}_{name}.pdf
				var baseName string
				if task.rec.CertificateID != "" {
					baseName = fmt.Sprintf("%s_%s.pdf", task.rec.CertificateID, recipientName)
				} else {
					baseName = fmt.Sprintf("%04d_%s.pdf", task.index+1, recipientName)
				}
				cleanFileName := filesystem.SanitizeFilename(baseName)

				// Temporary output path for this individual certificate
				tempPdfPath := filepath.Join(tempBatchDir, cleanFileName)

				// Render individual certificate
				renderErr := g.renderer.RenderToFile(doc, vars, assetRootDir, tempPdfPath)

				var itemStatus string
				var errMsg string

				if renderErr != nil {
					itemStatus = domain.ItemStatusFailed
					errMsg = renderErr.Error()
					atomic.AddInt32(&failedCounter, 1)
				} else {
					itemStatus = domain.ItemStatusCompleted
					atomic.AddInt32(&completedCounter, 1)

					// Always copy individual PDF directly to final output directory
					// so user can open/inspect individual certificates directly from Riwayat Batch
					finalIndivPath := filepath.Join(finalOutputDir, cleanFileName)
					_ = copyFile(tempPdfPath, finalIndivPath)
				}

				// Update item record in SQLite
				_ = g.jobRepo.UpdateJobItem(jobCtx, task.item.ID, itemStatus, cleanFileName, errMsg)

				summaryMu.Lock()
				summaryRows = append(summaryRows, []string{
					task.rec.ID,
					task.rec.CertificateID,
					recipientName,
					cleanFileName,
					itemStatus,
					errMsg,
				})
				summaryMu.Unlock()

				// Update Job progress in SQLite periodically
				cCount := int(atomic.LoadInt32(&completedCounter))
				fCount := int(atomic.LoadInt32(&failedCounter))
				_ = g.jobRepo.UpdateJobProgress(jobCtx, jobID, cCount, fCount)

				// Dispatch Progress Event
				if onProgress != nil {
					percent := int((float64(cCount+fCount) / float64(len(recipients))) * 100)
					onProgress(domain.ProgressEvent{
						JobID:                jobID,
						CurrentItem:          cCount + fCount,
						Total:                len(recipients),
						Completed:            cCount,
						Failed:               fCount,
						Percentage:           percent,
						Status:               domain.JobStatusRunning,
						CurrentRecipientName: recipientName,
					})
				}
			}
		}()
	}

	wg.Wait()

	// 8. Handle Cancellation State
	if isCancelled || jobCtx.Err() != nil {
		bgCtx, bgCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer bgCancel()

		_ = g.jobRepo.FinalizeJob(bgCtx, jobID, domain.JobStatusCancelled, "", "Job was cancelled by user")
		g.logger.Info("batch generation cancelled", "jobId", jobID)
		if onProgress != nil {
			onProgress(domain.ProgressEvent{
				JobID:      jobID,
				Total:      len(recipients),
				Completed:  int(atomic.LoadInt32(&completedCounter)),
				Failed:     int(atomic.LoadInt32(&failedCounter)),
				Status:     domain.JobStatusCancelled,
				Message:    "Job dibatalkan oleh pengguna",
				Percentage: int((float64(completedCounter+failedCounter) / float64(len(recipients))) * 100),
			})
		}
		return g.jobRepo.GetJobByID(bgCtx, jobID)
	}

	// 9. Write export-summary.csv in temp dir
	summaryCSVPath := filepath.Join(tempBatchDir, "export-summary.csv")
	summaryFile, err := os.Create(summaryCSVPath)
	if err == nil {
		w := csv.NewWriter(summaryFile)
		_ = w.WriteAll(summaryRows)
		w.Flush()
		_ = summaryFile.Close()
	}

	// 10. Streaming ZIP Archive or Directory Finalization
	var finalOutputPath string
	timestampStr := time.Now().Format("20060102-150405")

	if cfg.ExportZip {
		zipFileName := filesystem.SanitizeFilename(fmt.Sprintf("%s_certificates_%s.zip", tmpl.Name, timestampStr))
		finalOutputPath = filepath.Join(finalOutputDir, zipFileName)

		if err := createStreamingZip(tempBatchDir, finalOutputPath); err != nil {
			_ = g.jobRepo.FinalizeJob(ctx, jobID, domain.JobStatusFailed, "", fmt.Sprintf("Failed to build ZIP archive: %v", err))
			return nil, err
		}
	} else {
		finalOutputPath = finalOutputDir
	}

	// 11. Finalize Job Status
	finalCompleted := int(atomic.LoadInt32(&completedCounter))
	finalFailed := int(atomic.LoadInt32(&failedCounter))

	var finalStatus string
	if finalFailed == 0 {
		finalStatus = domain.JobStatusCompleted
	} else if finalCompleted > 0 {
		finalStatus = domain.JobStatusCompletedWithErrors
	} else {
		finalStatus = domain.JobStatusFailed
	}

	if err := g.jobRepo.FinalizeJob(ctx, jobID, finalStatus, finalOutputPath, ""); err != nil {
		return nil, err
	}

	if onProgress != nil {
		onProgress(domain.ProgressEvent{
			JobID:      jobID,
			Total:      len(recipients),
			Completed:  finalCompleted,
			Failed:     finalFailed,
			Percentage: 100,
			Status:     finalStatus,
			Message:    fmt.Sprintf("Selesai: %d sukses, %d gagal", finalCompleted, finalFailed),
		})
	}

	g.logger.Info("batch generation finished",
		"jobId", jobID,
		"completed", finalCompleted,
		"failed", finalFailed,
		"outputPath", finalOutputPath,
	)

	return g.jobRepo.GetJobByID(ctx, jobID)
}

// CancelJob cancels an actively running batch job.
func (g *BatchGenerator) CancelJob(jobID string) error {
	g.cancelsMu.Lock()
	cancel, ok := g.cancels[jobID]
	g.cancelsMu.Unlock()

	if !ok {
		return fmt.Errorf("job %s is not actively running", jobID)
	}

	cancel()
	_ = g.jobRepo.UpdateJobStatus(context.Background(), jobID, domain.JobStatusCancelling, "Cancelling...")
	return nil
}

// RetryFailedItems retries generating only the failed items of a previous job.
func (g *BatchGenerator) RetryFailedItems(
	ctx context.Context,
	jobID string,
	onProgress ProgressCallback,
) (*domain.GenerationJob, error) {
	job, err := g.jobRepo.GetJobByID(ctx, jobID)
	if err != nil {
		return nil, err
	}

	failedItems, err := g.jobRepo.ListFailedJobItems(ctx, jobID)
	if err != nil {
		return nil, err
	}
	if len(failedItems) == 0 {
		return job, nil // Nothing to retry
	}

	failedRecipientIDs := make([]string, len(failedItems))
	for i, it := range failedItems {
		failedRecipientIDs[i] = it.RecipientID
	}

	// Create a new batch execution for failed items
	retryConfig := domain.BatchConfig{
		ProjectID:           job.ProjectID,
		TemplateID:          job.TemplateID,
		RecipientIDs:        failedRecipientIDs,
		Concurrency:         1,
		ExportZip:           true,
		ExportIndividualPDF: false,
	}

	return g.ExecuteBatch(ctx, retryConfig, onProgress)
}

// createStreamingZip streams all files from sourceDir into a destination ZIP file.
func createStreamingZip(sourceDir, zipPath string) error {
	// Atomic write for ZIP file
	tempZipFile, err := os.CreateTemp(filepath.Dir(zipPath), "temp-archive-*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temp zip file: %w", err)
	}
	tempZipPath := tempZipFile.Name()
	defer func() {
		_ = os.Remove(tempZipPath)
	}()

	zipWriter := zip.NewWriter(tempZipFile)

	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		// Security: sanitize archive entry header
		cleanRel := strings.TrimLeft(filepath.ToSlash(relPath), "/")
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = cleanRel
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		fileToZip, err := os.Open(path)
		if err != nil {
			return err
		}
		defer fileToZip.Close()

		_, err = io.Copy(writer, fileToZip)
		return err
	})

	if err != nil {
		_ = zipWriter.Close()
		_ = tempZipFile.Close()
		return fmt.Errorf("error writing zip contents: %w", err)
	}

	if err := zipWriter.Close(); err != nil {
		_ = tempZipFile.Close()
		return err
	}
	if err := tempZipFile.Sync(); err != nil {
		_ = tempZipFile.Close()
		return err
	}
	if err := tempZipFile.Close(); err != nil {
		return err
	}

	return os.Rename(tempZipPath, zipPath)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
