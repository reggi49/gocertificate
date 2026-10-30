package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// SettingsRepository defines operations for key-value application settings.
type SettingsRepository interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string) error
	GetAll(ctx context.Context) (map[string]string, error)
}

// SQLiteSettingsRepository implements SettingsRepository.
type SQLiteSettingsRepository struct {
	db *sql.DB
}

// NewSettingsRepository constructs a new SQLiteSettingsRepository.
func NewSettingsRepository(db *sql.DB) *SQLiteSettingsRepository {
	return &SQLiteSettingsRepository{db: db}
}

// Get retrieves a setting by key.
func (r *SQLiteSettingsRepository) Get(ctx context.Context, key string) (string, error) {
	query := `SELECT value FROM app_settings WHERE key = ?`
	var val string
	err := r.db.QueryRowContext(ctx, query, key).Scan(&val)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("failed to get setting for %s: %w", key, err)
	}
	return val, nil
}

// Set inserts or updates a setting by key.
func (r *SQLiteSettingsRepository) Set(ctx context.Context, key, value string) error {
	query := `
		INSERT INTO app_settings (key, value, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET
			value = excluded.value,
			updated_at = excluded.updated_at
	`
	nowStr := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx, query, key, value, nowStr)
	if err != nil {
		return fmt.Errorf("failed to set setting for %s: %w", key, err)
	}
	return nil
}

// GetAll retrieves all settings as a map.
func (r *SQLiteSettingsRepository) GetAll(ctx context.Context) (map[string]string, error) {
	query := `SELECT key, value FROM app_settings`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all settings: %w", err)
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		settings[k] = v
	}

	return settings, rows.Err()
}
