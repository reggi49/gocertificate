package services

import (
	"context"
	"log/slog"

	"certificate-generator/internal/domain"
	certgen "certificate-generator/internal/generator/certificate"
	"certificate-generator/internal/repository/sqlite"
	"certificate-generator/internal/storage/filesystem"
)

// JobService coordinates batch generation execution, status queries, and cancellation.
type JobService struct {
	generator *certgen.BatchGenerator
	jobRepo   sqlite.JobRepository
	sm        *filesystem.StorageManager
	logger    *slog.Logger
}

// NewJobService constructs a new JobService.
func NewJobService(
	generator *certgen.BatchGenerator,
	jobRepo sqlite.JobRepository,
	sm *filesystem.StorageManager,
	logger *slog.Logger,
) *JobService {
	if logger == nil {
		logger = slog.Default()
	}
	return &JobService{
		generator: generator,
		jobRepo:   jobRepo,
		sm:        sm,
		logger:    logger,
	}
}

// StartBatchGeneration initiates a batch job and executes in background or synchronous wait.
func (s *JobService) StartBatchGeneration(
	ctx context.Context,
	cfg domain.BatchConfig,
	onProgress certgen.ProgressCallback,
) (*domain.GenerationJob, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return s.generator.ExecuteBatch(ctx, cfg, onProgress)
}

// CancelJob cancels an active job.
func (s *JobService) CancelJob(jobID string) error {
	if err := filesystem.ValidateID(jobID); err != nil {
		return domain.ErrInvalidProjectID
	}
	return s.generator.CancelJob(jobID)
}

// RetryFailedItems retries failed items of a job.
func (s *JobService) RetryFailedItems(
	ctx context.Context,
	jobID string,
	onProgress certgen.ProgressCallback,
) (*domain.GenerationJob, error) {
	if err := filesystem.ValidateID(jobID); err != nil {
		return nil, domain.ErrInvalidProjectID
	}
	return s.generator.RetryFailedItems(ctx, jobID, onProgress)
}

// GetJob retrieves a generation job by ID.
func (s *JobService) GetJob(ctx context.Context, id string) (*domain.GenerationJob, error) {
	if err := filesystem.ValidateID(id); err != nil {
		return nil, domain.ErrJobNotFound
	}
	return s.jobRepo.GetJobByID(ctx, id)
}

// ListJobsByProject returns all jobs for a project.
func (s *JobService) ListJobsByProject(ctx context.Context, projectID string) ([]domain.GenerationJob, error) {
	if err := filesystem.ValidateID(projectID); err != nil {
		return nil, domain.ErrInvalidProjectID
	}
	return s.jobRepo.ListJobsByProject(ctx, projectID)
}

// ListJobItems returns individual rendered items for a job.
func (s *JobService) ListJobItems(ctx context.Context, jobID string) ([]domain.GenerationJobItem, error) {
	if err := filesystem.ValidateID(jobID); err != nil {
		return nil, domain.ErrJobNotFound
	}
	return s.jobRepo.ListJobItems(ctx, jobID)
}

// DeleteJob permanently removes a batch generation job and its rendered items from database.
func (s *JobService) DeleteJob(ctx context.Context, id string) error {
	if err := filesystem.ValidateID(id); err != nil {
		return domain.ErrJobNotFound
	}
	return s.jobRepo.DeleteJob(ctx, id)
}

// CheckInterruptedJobs discovers unfinished jobs across restarts and marks them failed or cancelled cleanly.
func (s *JobService) CheckInterruptedJobs(ctx context.Context) ([]domain.GenerationJob, error) {
	interrupted, err := s.jobRepo.GetInterruptedJobs(ctx)
	if err != nil {
		return nil, err
	}

	for _, j := range interrupted {
		s.logger.Warn("detected interrupted job from previous session", "jobId", j.ID, "status", j.Status)
		_ = s.jobRepo.FinalizeJob(ctx, j.ID, domain.JobStatusFailed, "", "Proses terhenti saat aplikasi ditutup")
	}

	return interrupted, nil
}
