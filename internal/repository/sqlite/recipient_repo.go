package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"certificate-generator/internal/domain"
)

// RecipientRepository defines persistence operations for recipients.
type RecipientRepository interface {
	Create(ctx context.Context, r *domain.Recipient) error
	GetByID(ctx context.Context, id string) (*domain.Recipient, error)
	Update(ctx context.Context, r *domain.Recipient) error
	Delete(ctx context.Context, id string) error
	DeleteBatch(ctx context.Context, ids []string) error
	DeleteAllByProject(ctx context.Context, projectID string) error
	List(ctx context.Context, filter domain.RecipientFilter) (*domain.PaginatedRecipients, error)
	GetAllByProject(ctx context.Context, projectID string) ([]domain.Recipient, error)
	BatchCreate(ctx context.Context, projectID string, recipients []domain.Recipient, strategy domain.ConflictStrategy) (*domain.ImportSummary, error)
	Count(ctx context.Context, projectID string) (int, error)
}

// SQLiteRecipientRepository implements RecipientRepository.
type SQLiteRecipientRepository struct {
	db *sql.DB
}

// NewRecipientRepository constructs a new SQLiteRecipientRepository.
func NewRecipientRepository(db *sql.DB) *SQLiteRecipientRepository {
	return &SQLiteRecipientRepository{db: db}
}

// Create inserts a single recipient record.
func (r *SQLiteRecipientRepository) Create(ctx context.Context, rec *domain.Recipient) error {
	query := `
		INSERT INTO recipients (
			id, project_id, full_name, email, organization, event_name,
			issue_date, certificate_id, custom_fields_json, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	createdAtStr := rec.CreatedAt.UTC().Format(time.RFC3339)
	updatedAtStr := rec.UpdatedAt.UTC().Format(time.RFC3339)

	_, err := r.db.ExecContext(ctx, query,
		rec.ID,
		rec.ProjectID,
		rec.FullName,
		toNullString(rec.Email),
		toNullString(rec.Organization),
		toNullString(rec.EventName),
		toNullString(rec.IssueDate),
		toNullString(rec.CertificateID),
		toNullString(rec.CustomFieldsJSON),
		createdAtStr,
		updatedAtStr,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return domain.ErrDuplicateCertID
		}
		return fmt.Errorf("failed to insert recipient: %w", err)
	}
	return nil
}

// GetByID retrieves a recipient by ID.
func (r *SQLiteRecipientRepository) GetByID(ctx context.Context, id string) (*domain.Recipient, error) {
	query := `
		SELECT id, project_id, full_name, email, organization, event_name,
		       issue_date, certificate_id, custom_fields_json, created_at, updated_at
		FROM recipients
		WHERE id = ?
	`
	var rec domain.Recipient
	var email, org, event, issueDate, certID, customFields sql.NullString
	var createdAtStr, updatedAtStr string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rec.ID,
		&rec.ProjectID,
		&rec.FullName,
		&email,
		&org,
		&event,
		&issueDate,
		&certID,
		&customFields,
		&createdAtStr,
		&updatedAtStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrRecipientNotFound
		}
		return nil, fmt.Errorf("failed to query recipient by id: %w", err)
	}

	rec.Email = email.String
	rec.Organization = org.String
	rec.EventName = event.String
	rec.IssueDate = issueDate.String
	rec.CertificateID = certID.String
	rec.CustomFieldsJSON = customFields.String
	rec.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	rec.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)

	return &rec, nil
}

// Update updates an existing recipient.
func (r *SQLiteRecipientRepository) Update(ctx context.Context, rec *domain.Recipient) error {
	query := `
		UPDATE recipients
		SET full_name = ?, email = ?, organization = ?, event_name = ?,
		    issue_date = ?, certificate_id = ?, custom_fields_json = ?, updated_at = ?
		WHERE id = ?
	`
	updatedAtStr := rec.UpdatedAt.UTC().Format(time.RFC3339)
	res, err := r.db.ExecContext(ctx, query,
		rec.FullName,
		toNullString(rec.Email),
		toNullString(rec.Organization),
		toNullString(rec.EventName),
		toNullString(rec.IssueDate),
		toNullString(rec.CertificateID),
		toNullString(rec.CustomFieldsJSON),
		updatedAtStr,
		rec.ID,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return domain.ErrDuplicateCertID
		}
		return fmt.Errorf("failed to update recipient: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrRecipientNotFound
	}
	return nil
}

// Delete removes a recipient by ID.
func (r *SQLiteRecipientRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM recipients WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete recipient: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrRecipientNotFound
	}
	return nil
}

// DeleteBatch deletes multiple recipients by IDs in a single query.
func (r *SQLiteRecipientRepository) DeleteBatch(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf("DELETE FROM recipients WHERE id IN (%s)", strings.Join(placeholders, ","))
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete recipient batch: %w", err)
	}
	return nil
}

// DeleteAllByProject removes all recipients belonging to a project.
func (r *SQLiteRecipientRepository) DeleteAllByProject(ctx context.Context, projectID string) error {
	query := `DELETE FROM recipients WHERE project_id = ?`
	_, err := r.db.ExecContext(ctx, query, projectID)
	if err != nil {
		return fmt.Errorf("failed to delete all recipients for project: %w", err)
	}
	return nil
}

// List queries recipients with search and pagination.
func (r *SQLiteRecipientRepository) List(ctx context.Context, filter domain.RecipientFilter) (*domain.PaginatedRecipients, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 25
	}

	baseQuery := `FROM recipients WHERE project_id = ?`
	args := []any{filter.ProjectID}

	search := strings.TrimSpace(filter.Search)
	if search != "" {
		baseQuery += ` AND (full_name LIKE ? OR email LIKE ? OR certificate_id LIKE ?)`
		like := "%" + search + "%"
		args = append(args, like, like, like)
	}

	// 1. Get total count
	var total int
	countQuery := `SELECT COUNT(*) ` + baseQuery
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count recipients: %w", err)
	}

	// 2. Fetch page items
	offset := (filter.Page - 1) * filter.PageSize
	selectQuery := `
		SELECT id, project_id, full_name, email, organization, event_name,
		       issue_date, certificate_id, custom_fields_json, created_at, updated_at
	` + baseQuery + ` ORDER BY created_at ASC LIMIT ? OFFSET ?`

	pageArgs := append(args, filter.PageSize, offset)
	rows, err := r.db.QueryContext(ctx, selectQuery, pageArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to query recipient page: %w", err)
	}
	defer rows.Close()

	var items []domain.Recipient
	for rows.Next() {
		var rec domain.Recipient
		var email, org, event, issueDate, certID, customFields sql.NullString
		var createdAtStr, updatedAtStr string

		if err := rows.Scan(
			&rec.ID,
			&rec.ProjectID,
			&rec.FullName,
			&email,
			&org,
			&event,
			&issueDate,
			&certID,
			&customFields,
			&createdAtStr,
			&updatedAtStr,
		); err != nil {
			return nil, fmt.Errorf("failed to scan recipient row: %w", err)
		}

		rec.Email = email.String
		rec.Organization = org.String
		rec.EventName = event.String
		rec.IssueDate = issueDate.String
		rec.CertificateID = certID.String
		rec.CustomFieldsJSON = customFields.String
		rec.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		rec.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
		items = append(items, rec)
	}

	if items == nil {
		items = []domain.Recipient{}
	}

	totalPages := (total + filter.PageSize - 1) / filter.PageSize
	if totalPages == 0 {
		totalPages = 1
	}

	return &domain.PaginatedRecipients{
		Items:      items,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	}, rows.Err()
}

// GetAllByProject returns all recipients for a project.
func (r *SQLiteRecipientRepository) GetAllByProject(ctx context.Context, projectID string) ([]domain.Recipient, error) {
	query := `
		SELECT id, project_id, full_name, email, organization, event_name,
		       issue_date, certificate_id, custom_fields_json, created_at, updated_at
		FROM recipients
		WHERE project_id = ?
		ORDER BY created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get all recipients: %w", err)
	}
	defer rows.Close()

	var items []domain.Recipient
	for rows.Next() {
		var rec domain.Recipient
		var email, org, event, issueDate, certID, customFields sql.NullString
		var createdAtStr, updatedAtStr string

		if err := rows.Scan(
			&rec.ID,
			&rec.ProjectID,
			&rec.FullName,
			&email,
			&org,
			&event,
			&issueDate,
			&certID,
			&customFields,
			&createdAtStr,
			&updatedAtStr,
		); err != nil {
			return nil, fmt.Errorf("failed to scan recipient: %w", err)
		}

		rec.Email = email.String
		rec.Organization = org.String
		rec.EventName = event.String
		rec.IssueDate = issueDate.String
		rec.CertificateID = certID.String
		rec.CustomFieldsJSON = customFields.String
		rec.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		rec.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
		items = append(items, rec)
	}

	if items == nil {
		items = []domain.Recipient{}
	}
	return items, rows.Err()
}

// BatchCreate inserts multiple recipients transactionally according to conflict strategy.
func (r *SQLiteRecipientRepository) BatchCreate(
	ctx context.Context,
	projectID string,
	recipients []domain.Recipient,
	strategy domain.ConflictStrategy,
) (*domain.ImportSummary, error) {
	summary := &domain.ImportSummary{
		TotalRows: len(recipients),
		Errors:    []domain.ImportRowError{},
	}

	if len(recipients) == 0 {
		return summary, nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	insertStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO recipients (
			id, project_id, full_name, email, organization, event_name,
			issue_date, certificate_id, custom_fields_json, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer insertStmt.Close()

	updateStmt, err := tx.PrepareContext(ctx, `
		UPDATE recipients
		SET full_name = ?, email = ?, organization = ?, event_name = ?,
		    issue_date = ?, custom_fields_json = ?, updated_at = ?
		WHERE project_id = ? AND certificate_id = ?
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare update statement: %w", err)
	}
	defer updateStmt.Close()

	checkStmt, err := tx.PrepareContext(ctx, `
		SELECT COUNT(*) FROM recipients WHERE project_id = ? AND certificate_id = ?
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare check statement: %w", err)
	}
	defer checkStmt.Close()

	for idx, rec := range recipients {
		rowNumber := idx + 2 // +1 for 1-based, +1 for header row
		nowStr := time.Now().UTC().Format(time.RFC3339)

		hasCertID := rec.CertificateID != ""
		isDuplicate := false

		if hasCertID {
			var count int
			_ = checkStmt.QueryRowContext(ctx, projectID, rec.CertificateID).Scan(&count)
			if count > 0 {
				isDuplicate = true
			}
		}

		if isDuplicate {
			switch strategy {
			case domain.ConflictStrategySkip:
				summary.SkippedRows++
				continue

			case domain.ConflictStrategyFail:
				summary.Errors = append(summary.Errors, domain.ImportRowError{
					RowNumber: rowNumber,
					Column:    "certificate_id",
					Category:  "DUPLICATE_CERTIFICATE_ID",
					Message:   fmt.Sprintf("Certificate ID '%s' already exists in this project", rec.CertificateID),
					RawValue:  rec.CertificateID,
				})
				return summary, fmt.Errorf("import failed on row %d: duplicate certificate id %s", rowNumber, rec.CertificateID)

			case domain.ConflictStrategyUpdate:
				_, err := updateStmt.ExecContext(ctx,
					rec.FullName,
					toNullString(rec.Email),
					toNullString(rec.Organization),
					toNullString(rec.EventName),
					toNullString(rec.IssueDate),
					toNullString(rec.CustomFieldsJSON),
					nowStr,
					projectID,
					rec.CertificateID,
				)
				if err != nil {
					return nil, fmt.Errorf("failed to update duplicate recipient: %w", err)
				}
				summary.UpdatedRows++
				summary.ImportedRows++
				continue
			}
		}

		// Insert new record
		_, err = insertStmt.ExecContext(ctx,
			rec.ID,
			projectID,
			rec.FullName,
			toNullString(rec.Email),
			toNullString(rec.Organization),
			toNullString(rec.EventName),
			toNullString(rec.IssueDate),
			toNullString(rec.CertificateID),
			toNullString(rec.CustomFieldsJSON),
			nowStr,
			nowStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert recipient on row %d: %w", rowNumber, err)
		}
		summary.ImportedRows++
	}

	summary.ValidRows = summary.ImportedRows

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit batch recipients: %w", err)
	}

	return summary, nil
}

// Count returns the total recipient count for a project.
func (r *SQLiteRecipientRepository) Count(ctx context.Context, projectID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM recipients WHERE project_id = ?`, projectID).Scan(&count)
	return count, err
}

func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
