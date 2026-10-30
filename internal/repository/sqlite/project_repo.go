package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"certificate-generator/internal/domain"
)

// ProjectRepository defines persistence operations for Projects.
type ProjectRepository interface {
	Create(ctx context.Context, p *domain.Project) error
	GetByID(ctx context.Context, id string) (*domain.Project, error)
	GetByName(ctx context.Context, name string) (*domain.Project, error)
	List(ctx context.Context, includeArchived bool) ([]domain.Project, error)
	Recent(ctx context.Context, limit int) ([]domain.Project, error)
	Update(ctx context.Context, p *domain.Project) error
	SetArchived(ctx context.Context, id string, isArchived bool) error
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context) (total int, active int, archived int, err error)
}

// SQLiteProjectRepository implements ProjectRepository.
type SQLiteProjectRepository struct {
	db *sql.DB
}

// NewProjectRepository constructs a new SQLiteProjectRepository.
func NewProjectRepository(db *sql.DB) *SQLiteProjectRepository {
	return &SQLiteProjectRepository{db: db}
}

// Create inserts a new project record.
func (r *SQLiteProjectRepository) Create(ctx context.Context, p *domain.Project) error {
	query := `
		INSERT INTO projects (id, name, description, event_date, event_location, event_address, created_at, updated_at, is_archived)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	isArchivedInt := 0
	if p.IsArchived {
		isArchivedInt = 1
	}

	createdAtStr := p.CreatedAt.UTC().Format(time.RFC3339)
	updatedAtStr := p.UpdatedAt.UTC().Format(time.RFC3339)

	_, err := r.db.ExecContext(ctx, query, p.ID, p.Name, p.Description, p.EventDate, p.EventLocation, p.EventAddress, createdAtStr, updatedAtStr, isArchivedInt)
	if err != nil {
		return fmt.Errorf("failed to insert project: %w", err)
	}
	return nil
}

// GetByID retrieves a project by its primary key ID.
func (r *SQLiteProjectRepository) GetByID(ctx context.Context, id string) (*domain.Project, error) {
	query := `
		SELECT id, name, description, event_date, event_location, event_address, created_at, updated_at, is_archived
		FROM projects
		WHERE id = ?
	`
	var p domain.Project
	var desc, eventDate, eventLoc, eventAddr sql.NullString
	var createdAtStr, updatedAtStr string
	var isArchivedInt int

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID,
		&p.Name,
		&desc,
		&eventDate,
		&eventLoc,
		&eventAddr,
		&createdAtStr,
		&updatedAtStr,
		&isArchivedInt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to query project by id: %w", err)
	}

	p.Description = desc.String
	p.EventDate = eventDate.String
	p.EventLocation = eventLoc.String
	p.EventAddress = eventAddr.String
	p.IsArchived = isArchivedInt == 1
	p.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	p.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)

	return &p, nil
}

// GetByName retrieves a project by its exact name.
func (r *SQLiteProjectRepository) GetByName(ctx context.Context, name string) (*domain.Project, error) {
	query := `
		SELECT id, name, description, event_date, event_location, event_address, created_at, updated_at, is_archived
		FROM projects
		WHERE name = ?
	`
	var p domain.Project
	var desc, eventDate, eventLoc, eventAddr sql.NullString
	var createdAtStr, updatedAtStr string
	var isArchivedInt int

	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&p.ID,
		&p.Name,
		&desc,
		&eventDate,
		&eventLoc,
		&eventAddr,
		&createdAtStr,
		&updatedAtStr,
		&isArchivedInt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to query project by name: %w", err)
	}

	p.Description = desc.String
	p.EventDate = eventDate.String
	p.EventLocation = eventLoc.String
	p.EventAddress = eventAddr.String
	p.IsArchived = isArchivedInt == 1
	p.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	p.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)

	return &p, nil
}

// List returns projects, optionally including archived projects.
func (r *SQLiteProjectRepository) List(ctx context.Context, includeArchived bool) ([]domain.Project, error) {
	query := `
		SELECT id, name, description, event_date, event_location, event_address, created_at, updated_at, is_archived
		FROM projects
	`
	if !includeArchived {
		query += " WHERE is_archived = 0"
	}
	query += " ORDER BY updated_at DESC"

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	defer rows.Close()

	var projects []domain.Project
	for rows.Next() {
		var p domain.Project
		var desc, eventDate, eventLoc, eventAddr sql.NullString
		var createdAtStr, updatedAtStr string
		var isArchivedInt int

		if err := rows.Scan(&p.ID, &p.Name, &desc, &eventDate, &eventLoc, &eventAddr, &createdAtStr, &updatedAtStr, &isArchivedInt); err != nil {
			return nil, fmt.Errorf("failed to scan project row: %w", err)
		}

		p.Description = desc.String
		p.EventDate = eventDate.String
		p.EventLocation = eventLoc.String
		p.EventAddress = eventAddr.String
		p.IsArchived = isArchivedInt == 1
		p.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		p.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
		projects = append(projects, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating projects: %w", err)
	}

	if projects == nil {
		projects = []domain.Project{}
	}
	return projects, nil
}

// Recent returns up to limit recent active projects.
func (r *SQLiteProjectRepository) Recent(ctx context.Context, limit int) ([]domain.Project, error) {
	if limit <= 0 {
		limit = 5
	}
	query := `
		SELECT id, name, description, event_date, event_location, event_address, created_at, updated_at, is_archived
		FROM projects
		WHERE is_archived = 0
		ORDER BY updated_at DESC
		LIMIT ?
	`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent projects: %w", err)
	}
	defer rows.Close()

	var projects []domain.Project
	for rows.Next() {
		var p domain.Project
		var desc, eventDate, eventLoc, eventAddr sql.NullString
		var createdAtStr, updatedAtStr string
		var isArchivedInt int

		if err := rows.Scan(&p.ID, &p.Name, &desc, &eventDate, &eventLoc, &eventAddr, &createdAtStr, &updatedAtStr, &isArchivedInt); err != nil {
			return nil, fmt.Errorf("failed to scan recent project: %w", err)
		}

		p.Description = desc.String
		p.EventDate = eventDate.String
		p.EventLocation = eventLoc.String
		p.EventAddress = eventAddr.String
		p.IsArchived = isArchivedInt == 1
		p.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		p.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
		projects = append(projects, p)
	}

	if projects == nil {
		projects = []domain.Project{}
	}
	return projects, nil
}

// Update updates name, description, event details, and updated_at.
func (r *SQLiteProjectRepository) Update(ctx context.Context, p *domain.Project) error {
	query := `
		UPDATE projects
		SET name = ?, description = ?, event_date = ?, event_location = ?, event_address = ?, updated_at = ?
		WHERE id = ?
	`
	updatedAtStr := p.UpdatedAt.UTC().Format(time.RFC3339)
	res, err := r.db.ExecContext(ctx, query, p.Name, p.Description, p.EventDate, p.EventLocation, p.EventAddress, updatedAtStr, p.ID)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrProjectNotFound
	}
	return nil
}

// SetArchived marks a project as archived or active.
func (r *SQLiteProjectRepository) SetArchived(ctx context.Context, id string, isArchived bool) error {
	query := `
		UPDATE projects
		SET is_archived = ?, updated_at = ?
		WHERE id = ?
	`
	isArchivedInt := 0
	if isArchived {
		isArchivedInt = 1
	}
	nowStr := time.Now().UTC().Format(time.RFC3339)

	res, err := r.db.ExecContext(ctx, query, isArchivedInt, nowStr, id)
	if err != nil {
		return fmt.Errorf("failed to set project archive state: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrProjectNotFound
	}
	return nil
}

// Delete permanently removes a project record from SQLite.
func (r *SQLiteProjectRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM projects WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrProjectNotFound
	}
	return nil
}

// Count returns the count of total, active, and archived projects.
func (r *SQLiteProjectRepository) Count(ctx context.Context) (total int, active int, archived int, err error) {
	query := `
		SELECT
			COUNT(*),
			COALESCE(SUM(CASE WHEN is_archived = 0 THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN is_archived = 1 THEN 1 ELSE 0 END), 0)
		FROM projects
	`
	err = r.db.QueryRowContext(ctx, query).Scan(&total, &active, &archived)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to count projects: %w", err)
	}
	return total, active, archived, nil
}
