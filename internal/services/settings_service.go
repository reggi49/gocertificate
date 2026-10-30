package services

import (
	"context"
	"strconv"

	"certificate-generator/internal/domain"
	"certificate-generator/internal/repository/sqlite"
)

// SettingsService manages reading and writing user preferences.
type SettingsService struct {
	repo sqlite.SettingsRepository
}

// NewSettingsService constructs a new SettingsService.
func NewSettingsService(repo sqlite.SettingsRepository) *SettingsService {
	return &SettingsService{repo: repo}
}

// GetSettings loads settings from SQLite, filling in defaults where missing.
func (s *SettingsService) GetSettings(ctx context.Context) (*domain.AppSettings, error) {
	settings := domain.DefaultSettings()

	all, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	if val, ok := all["defaultOutputDir"]; ok && val != "" {
		settings.DefaultOutputDir = val
	}
	if val, ok := all["theme"]; ok && val != "" {
		settings.Theme = val
	}
	if val, ok := all["language"]; ok && val != "" {
		settings.Language = val
	}
	if val, ok := all["autoSaveIntervalSec"]; ok && val != "" {
		if interval, err := strconv.Atoi(val); err == nil && interval > 0 {
			settings.AutoSaveIntervalSec = interval
		}
	}

	return &settings, nil
}

// UpdateSettings validates and updates application settings.
func (s *SettingsService) UpdateSettings(ctx context.Context, settings domain.AppSettings) error {
	if err := domain.ValidateSettings(&settings); err != nil {
		return err
	}

	if err := s.repo.Set(ctx, "defaultOutputDir", settings.DefaultOutputDir); err != nil {
		return err
	}
	if err := s.repo.Set(ctx, "theme", settings.Theme); err != nil {
		return err
	}
	if err := s.repo.Set(ctx, "language", settings.Language); err != nil {
		return err
	}
	if err := s.repo.Set(ctx, "autoSaveIntervalSec", strconv.Itoa(settings.AutoSaveIntervalSec)); err != nil {
		return err
	}

	return nil
}
