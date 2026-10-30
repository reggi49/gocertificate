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
	SetDefault(ctx context.Context, projectID string, templateID string) error
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
	isDefaultInt := 0
	if t.IsDefault {
		isDefaultInt = 1
	} else {
		var count int
		_ = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM templates WHERE project_id = ?", t.ProjectID).Scan(&count)
		if count == 0 {
			isDefaultInt = 1
			t.IsDefault = true
		}
	}

	query := `
		INSERT INTO templates (id, project_id, name, document_json, schema_version, is_default, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	createdAtStr := t.CreatedAt.UTC().Format(time.RFC3339)
	updatedAtStr := t.UpdatedAt.UTC().Format(time.RFC3339)

	_, err := r.db.ExecContext(ctx, query, t.ID, t.ProjectID, t.Name, t.DocumentJSON, t.SchemaVersion, isDefaultInt, createdAtStr, updatedAtStr)
	if err != nil {
		return fmt.Errorf("failed to insert template: %w", err)
	}
	return nil
}

// GetByID retrieves a template by ID.
func (r *SQLiteTemplateRepository) GetByID(ctx context.Context, id string) (*domain.Template, error) {
	query := `
		SELECT id, project_id, name, document_json, schema_version, is_default, created_at, updated_at
		FROM templates
		WHERE id = ?
	`
	var t domain.Template
	var isDefaultInt int
	var createdAtStr, updatedAtStr string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID,
		&t.ProjectID,
		&t.Name,
		&t.DocumentJSON,
		&t.SchemaVersion,
		&isDefaultInt,
		&createdAtStr,
		&updatedAtStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrTemplateNotFound
		}
		return nil, fmt.Errorf("failed to query template by id: %w", err)
	}

	t.IsDefault = (isDefaultInt == 1)
	t.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	t.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)

	return &t, nil
}

// ListByProject returns all templates associated with a project ordered by is_default desc, updated_at desc.
func (r *SQLiteTemplateRepository) ListByProject(ctx context.Context, projectID string) ([]domain.Template, error) {
	query := `
		SELECT id, project_id, name, document_json, schema_version, is_default, created_at, updated_at
		FROM templates
		WHERE project_id = ?
		ORDER BY is_default DESC, updated_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}
	defer rows.Close()

	var templates []domain.Template
	for rows.Next() {
		var t domain.Template
		var isDefaultInt int
		var createdAtStr, updatedAtStr string

		if err := rows.Scan(
			&t.ID,
			&t.ProjectID,
			&t.Name,
			&t.DocumentJSON,
			&t.SchemaVersion,
			&isDefaultInt,
			&createdAtStr,
			&updatedAtStr,
		); err != nil {
			return nil, fmt.Errorf("failed to scan template row: %w", err)
		}

		t.IsDefault = (isDefaultInt == 1)
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
	var wasDefault int
	var projectID string
	_ = r.db.QueryRowContext(ctx, "SELECT is_default, project_id FROM templates WHERE id = ?", id).Scan(&wasDefault, &projectID)

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

	if wasDefault == 1 && projectID != "" {
		_, _ = r.db.ExecContext(ctx, `
			UPDATE templates SET is_default = 1
			WHERE id = (SELECT id FROM templates WHERE project_id = ? ORDER BY updated_at DESC LIMIT 1)
		`, projectID)
	}

	return nil
}

// SetDefault sets a template as the default for its project.
func (r *SQLiteTemplateRepository) SetDefault(ctx context.Context, projectID string, templateID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "UPDATE templates SET is_default = 0 WHERE project_id = ?", projectID); err != nil {
		return fmt.Errorf("failed to reset default templates: %w", err)
	}

	res, err := tx.ExecContext(ctx, "UPDATE templates SET is_default = 1 WHERE project_id = ? AND id = ?", projectID, templateID)
	if err != nil {
		return fmt.Errorf("failed to set default template: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrTemplateNotFound
	}

	return tx.Commit()
}
