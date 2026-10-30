package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"certificate-generator/internal/domain"
)

// JobRepository defines persistence operations for generation jobs and item records.
type JobRepository interface {
	CreateJob(ctx context.Context, job *domain.GenerationJob) error
	GetJobByID(ctx context.Context, id string) (*domain.GenerationJob, error)
	ListJobsByProject(ctx context.Context, projectID string) ([]domain.GenerationJob, error)
	UpdateJobStatus(ctx context.Context, id, targetStatus, errorMsg string) error
	UpdateJobProgress(ctx context.Context, id string, completed, failed int) error
	FinalizeJob(ctx context.Context, id, finalStatus, outputPath, errorMsg string) error

	CreateJobItems(ctx context.Context, items []domain.GenerationJobItem) error
	UpdateJobItem(ctx context.Context, itemID, status, outputFileName, errorMsg string) error
	ListJobItems(ctx context.Context, jobID string) ([]domain.GenerationJobItem, error)
	ListFailedJobItems(ctx context.Context, jobID string) ([]domain.GenerationJobItem, error)

	GetInterruptedJobs(ctx context.Context) ([]domain.GenerationJob, error)
	DeleteJob(ctx context.Context, id string) error
}

// SQLiteJobRepository implements JobRepository.
type SQLiteJobRepository struct {
	db *sql.DB
}

// NewJobRepository constructs a new SQLiteJobRepository.
func NewJobRepository(db *sql.DB) *SQLiteJobRepository {
	return &SQLiteJobRepository{db: db}
}

// CreateJob persists a new generation job record.
func (r *SQLiteJobRepository) CreateJob(ctx context.Context, job *domain.GenerationJob) error {
	query := `
		INSERT INTO generation_jobs (
			id, project_id, template_id, status, total_count,
			completed_count, failed_count, output_path, error_message, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	nowStr := time.Now().UTC().Format(time.RFC3339)
	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now().UTC()
	}
	if job.UpdatedAt.IsZero() {
		job.UpdatedAt = time.Now().UTC()
	}

	_, err := r.db.ExecContext(ctx, query,
		job.ID,
		job.ProjectID,
		job.TemplateID,
		job.Status,
		job.TotalCount,
		job.CompletedCount,
		job.FailedCount,
		toNullString(job.OutputPath),
		toNullString(job.ErrorMessage),
		job.CreatedAt.Format(time.RFC3339),
		nowStr,
	)
	if err != nil {
		return fmt.Errorf("failed to insert generation job: %w", err)
	}
	return nil
}

// GetJobByID retrieves a generation job by ID.
func (r *SQLiteJobRepository) GetJobByID(ctx context.Context, id string) (*domain.GenerationJob, error) {
	query := `
		SELECT id, project_id, template_id, status, total_count,
		       completed_count, failed_count, output_path, error_message, created_at, updated_at
		FROM generation_jobs
		WHERE id = ?
	`
	var job domain.GenerationJob
	var outputPath, errorMsg sql.NullString
	var createdAtStr, updatedAtStr string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&job.ID,
		&job.ProjectID,
		&job.TemplateID,
		&job.Status,
		&job.TotalCount,
		&job.CompletedCount,
		&job.FailedCount,
		&outputPath,
		&errorMsg,
		&createdAtStr,
		&updatedAtStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrJobNotFound
		}
		return nil, fmt.Errorf("failed to query generation job: %w", err)
	}

	job.OutputPath = outputPath.String
	job.ErrorMessage = errorMsg.String
	job.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	job.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)

	return &job, nil
}

// ListJobsByProject returns all generation jobs for a project ordered by updated_at desc.
func (r *SQLiteJobRepository) ListJobsByProject(ctx context.Context, projectID string) ([]domain.GenerationJob, error) {
	query := `
		SELECT id, project_id, template_id, status, total_count,
		       completed_count, failed_count, output_path, error_message, created_at, updated_at
		FROM generation_jobs
		WHERE project_id = ?
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list generation jobs: %w", err)
	}
	defer rows.Close()

	var jobs []domain.GenerationJob
	for rows.Next() {
		var job domain.GenerationJob
		var outputPath, errorMsg sql.NullString
		var createdAtStr, updatedAtStr string

		if err := rows.Scan(
			&job.ID,
			&job.ProjectID,
			&job.TemplateID,
			&job.Status,
			&job.TotalCount,
			&job.CompletedCount,
			&job.FailedCount,
			&outputPath,
			&errorMsg,
			&createdAtStr,
			&updatedAtStr,
		); err != nil {
			return nil, err
		}

		job.OutputPath = outputPath.String
		job.ErrorMessage = errorMsg.String
		job.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		job.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
		jobs = append(jobs, job)
	}

	if jobs == nil {
		jobs = []domain.GenerationJob{}
	}
	return jobs, rows.Err()
}

// UpdateJobStatus updates the status enforcing state machine rules.
func (r *SQLiteJobRepository) UpdateJobStatus(ctx context.Context, id, targetStatus, errorMsg string) error {
	currentJob, err := r.GetJobByID(ctx, id)
	if err != nil {
		return err
	}

	if !domain.CanTransition(currentJob.Status, targetStatus) {
		return fmt.Errorf("%w: cannot transition from %s to %s",
			domain.ErrInvalidStateTransition, currentJob.Status, targetStatus)
	}

	nowStr := time.Now().UTC().Format(time.RFC3339)
	query := `
		UPDATE generation_jobs
		SET status = ?, error_message = COALESCE(NULLIF(?, ''), error_message), updated_at = ?
		WHERE id = ?
	`
	_, err = r.db.ExecContext(ctx, query, targetStatus, errorMsg, nowStr, id)
	if err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}
	return nil
}

// UpdateJobProgress updates current completed and failed counts during batch execution.
func (r *SQLiteJobRepository) UpdateJobProgress(ctx context.Context, id string, completed, failed int) error {
	nowStr := time.Now().UTC().Format(time.RFC3339)
	query := `
		UPDATE generation_jobs
		SET completed_count = ?, failed_count = ?, updated_at = ?
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query, completed, failed, nowStr, id)
	if err != nil {
		return fmt.Errorf("failed to update job progress: %w", err)
	}
	return nil
}

// FinalizeJob completes a generation job with output path and final terminal status.
func (r *SQLiteJobRepository) FinalizeJob(ctx context.Context, id, finalStatus, outputPath, errorMsg string) error {
	currentJob, err := r.GetJobByID(ctx, id)
	if err != nil {
		return err
	}

	if !domain.CanTransition(currentJob.Status, finalStatus) {
		return fmt.Errorf("%w: cannot finalize from %s to %s",
			domain.ErrInvalidStateTransition, currentJob.Status, finalStatus)
	}

	nowStr := time.Now().UTC().Format(time.RFC3339)
	query := `
		UPDATE generation_jobs
		SET status = ?, output_path = ?, error_message = ?, updated_at = ?
		WHERE id = ?
	`
	_, err = r.db.ExecContext(ctx, query, finalStatus, toNullString(outputPath), toNullString(errorMsg), nowStr, id)
	if err != nil {
		return fmt.Errorf("failed to finalize job: %w", err)
	}
	return nil
}

// CreateJobItems inserts item rows for a batch job.
func (r *SQLiteJobRepository) CreateJobItems(ctx context.Context, items []domain.GenerationJobItem) error {
	if len(items) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO generation_job_items (
			id, job_id, recipient_id, recipient_name, certificate_id,
			output_filename, status, error_message, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	nowStr := time.Now().UTC().Format(time.RFC3339)
	for _, item := range items {
		_, err := stmt.ExecContext(ctx,
			item.ID,
			item.JobID,
			item.RecipientID,
			item.RecipientName,
			toNullString(item.CertificateID),
			toNullString(item.OutputFileName),
			item.Status,
			toNullString(item.ErrorMessage),
			nowStr,
			nowStr,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// UpdateJobItem updates rendering status and result filename for one recipient item.
func (r *SQLiteJobRepository) UpdateJobItem(ctx context.Context, itemID, status, outputFileName, errorMsg string) error {
	nowStr := time.Now().UTC().Format(time.RFC3339)
	query := `
		UPDATE generation_job_items
		SET status = ?, output_filename = ?, error_message = ?, updated_at = ?
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query, status, toNullString(outputFileName), toNullString(errorMsg), nowStr, itemID)
	return err
}

// ListJobItems returns all item results for a job.
func (r *SQLiteJobRepository) ListJobItems(ctx context.Context, jobID string) ([]domain.GenerationJobItem, error) {
	query := `
		SELECT id, job_id, recipient_id, recipient_name, certificate_id,
		       output_filename, status, error_message, created_at, updated_at
		FROM generation_job_items
		WHERE job_id = ?
		ORDER BY created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.GenerationJobItem
	for rows.Next() {
		var item domain.GenerationJobItem
		var certID, outputFileName, errorMsg sql.NullString
		var createdAtStr, updatedAtStr string

		if err := rows.Scan(
			&item.ID,
			&item.JobID,
			&item.RecipientID,
			&item.RecipientName,
			&certID,
			&outputFileName,
			&item.Status,
			&errorMsg,
			&createdAtStr,
			&updatedAtStr,
		); err != nil {
			return nil, err
		}

		item.CertificateID = certID.String
		item.OutputFileName = outputFileName.String
		item.ErrorMessage = errorMsg.String
		item.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		item.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
		items = append(items, item)
	}

	if items == nil {
		items = []domain.GenerationJobItem{}
	}
	return items, rows.Err()
}

// ListFailedJobItems retrieves all items with status 'failed' within a job for retry operations.
func (r *SQLiteJobRepository) ListFailedJobItems(ctx context.Context, jobID string) ([]domain.GenerationJobItem, error) {
	query := `
		SELECT id, job_id, recipient_id, recipient_name, certificate_id,
		       output_filename, status, error_message, created_at, updated_at
		FROM generation_job_items
		WHERE job_id = ? AND status = ?
		ORDER BY created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, jobID, domain.ItemStatusFailed)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.GenerationJobItem
	for rows.Next() {
		var item domain.GenerationJobItem
		var certID, outputFileName, errorMsg sql.NullString
		var createdAtStr, updatedAtStr string

		if err := rows.Scan(
			&item.ID,
			&item.JobID,
			&item.RecipientID,
			&item.RecipientName,
			&certID,
			&outputFileName,
			&item.Status,
			&errorMsg,
			&createdAtStr,
			&updatedAtStr,
		); err != nil {
			return nil, err
		}

		item.CertificateID = certID.String
		item.OutputFileName = outputFileName.String
		item.ErrorMessage = errorMsg.String
		item.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		item.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
		items = append(items, item)
	}

	if items == nil {
		items = []domain.GenerationJobItem{}
	}
	return items, rows.Err()
}

// GetInterruptedJobs queries jobs that were left in non-terminal states (running, validating, pending) across app restarts.
func (r *SQLiteJobRepository) GetInterruptedJobs(ctx context.Context) ([]domain.GenerationJob, error) {
	query := `
		SELECT id, project_id, template_id, status, total_count,
		       completed_count, failed_count, output_path, error_message, created_at, updated_at
		FROM generation_jobs
		WHERE status IN ('pending', 'validating', 'running', 'cancelling')
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []domain.GenerationJob
	for rows.Next() {
		var job domain.GenerationJob
		var outputPath, errorMsg sql.NullString
		var createdAtStr, updatedAtStr string

		if err := rows.Scan(
			&job.ID,
			&job.ProjectID,
			&job.TemplateID,
			&job.Status,
			&job.TotalCount,
			&job.CompletedCount,
			&job.FailedCount,
			&outputPath,
			&errorMsg,
			&createdAtStr,
			&updatedAtStr,
		); err != nil {
			return nil, err
		}

		job.OutputPath = outputPath.String
		job.ErrorMessage = errorMsg.String
		job.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		job.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
		jobs = append(jobs, job)
	}

	if jobs == nil {
		jobs = []domain.GenerationJob{}
	}
	return jobs, rows.Err()
}

// DeleteJob permanently removes a job and its associated items.
func (r *SQLiteJobRepository) DeleteJob(ctx context.Context, id string) error {
	query := `DELETE FROM generation_jobs WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete generation job: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrJobNotFound
	}
	return nil
}
