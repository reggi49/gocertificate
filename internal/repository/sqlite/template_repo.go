package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"certificate-generator/internal/domain"
)

// TemplateRepository defines persistence operations for certificate templates.
type TemplateRepository interface {
	Create(ctx context.Context, t *domain.Template) error
	GetByID(ctx context.Context, id string) (*domain.Template, error)
	ListByProject(ctx context.Context, projectID string) ([]domain.Template, error)
	Update(ctx context.Context, t *domain.Template) error
	Delete(ctx context.Context, id string) error
}

// SQLiteTemplateRepository implements TemplateRepository.
type SQLiteTemplateRepository struct {
	db *sql.DB
}

// NewTemplateRepository constructs a new SQLiteTemplateRepository.
func NewTemplateRepository(db *sql.DB) *SQLiteTemplateRepository {
	return &SQLiteTemplateRepository{db: db}
}

// Create inserts a new template record.
func (r *SQLiteTemplateRepository) Create(ctx context.Context, t *domain.Template) error {
	query := `
		INSERT INTO templates (id, project_id, name, document_json, schema_version, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	createdAtStr := t.CreatedAt.UTC().Format(time.RFC3339)
	updatedAtStr := t.UpdatedAt.UTC().Format(time.RFC3339)

	_, err := r.db.ExecContext(ctx, query, t.ID, t.ProjectID, t.Name, t.DocumentJSON, t.SchemaVersion, createdAtStr, updatedAtStr)
	if err != nil {
		return fmt.Errorf("failed to insert template: %w", err)
	}
	return nil
}

// GetByID retrieves a template by ID.
func (r *SQLiteTemplateRepository) GetByID(ctx context.Context, id string) (*domain.Template, error) {
	query := `
		SELECT id, project_id, name, document_json, schema_version, created_at, updated_at
		FROM templates
		WHERE id = ?
	`
	var t domain.Template
	var createdAtStr, updatedAtStr string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID,
		&t.ProjectID,
		&t.Name,
		&t.DocumentJSON,
		&t.SchemaVersion,
		&createdAtStr,
		&updatedAtStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrTemplateNotFound
		}
		return nil, fmt.Errorf("failed to query template by id: %w", err)
	}

	t.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	t.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)

	return &t, nil
}

// ListByProject returns all templates associated with a project ordered by updated_at desc.
func (r *SQLiteTemplateRepository) ListByProject(ctx context.Context, projectID string) ([]domain.Template, error) {
	query := `
		SELECT id, project_id, name, document_json, schema_version, created_at, updated_at
		FROM templates
		WHERE project_id = ?
		ORDER BY updated_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}
	defer rows.Close()

	var templates []domain.Template
	for rows.Next() {
		var t domain.Template
		var createdAtStr, updatedAtStr string

		if err := rows.Scan(
			&t.ID,
			&t.ProjectID,
			&t.Name,
			&t.DocumentJSON,
			&t.SchemaVersion,
			&createdAtStr,
			&updatedAtStr,
		); err != nil {
			return nil, fmt.Errorf("failed to scan template row: %w", err)
		}

		t.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		t.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
		templates = append(templates, t)
	}

	if templates == nil {
		templates = []domain.Template{}
	}
	return templates, rows.Err()
}

// Update updates template name, document_json, and updated_at.
func (r *SQLiteTemplateRepository) Update(ctx context.Context, t *domain.Template) error {
	query := `
		UPDATE templates
		SET name = ?, document_json = ?, schema_version = ?, updated_at = ?
		WHERE id = ?
	`
	updatedAtStr := t.UpdatedAt.UTC().Format(time.RFC3339)
	res, err := r.db.ExecContext(ctx, query, t.Name, t.DocumentJSON, t.SchemaVersion, updatedAtStr, t.ID)
	if err != nil {
		return fmt.Errorf("failed to update template: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrTemplateNotFound
	}
	return nil
}

// Delete permanently removes a template from SQLite.
func (r *SQLiteTemplateRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM templates WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrTemplateNotFound
	}
	return nil
}
