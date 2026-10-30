package services

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"certificate-generator/internal/domain"
	"certificate-generator/internal/repository/sqlite"
	"certificate-generator/internal/storage/filesystem"

	"github.com/google/uuid"
)

// TemplateService manages certificate template creation, loading, updating, and syncing with filesystem.
type TemplateService struct {
	templateRepo sqlite.TemplateRepository
	projectRepo  sqlite.ProjectRepository
	sm           *filesystem.StorageManager
	logger       *slog.Logger
}

// NewTemplateService constructs a new TemplateService.
func NewTemplateService(
	templateRepo sqlite.TemplateRepository,
	projectRepo sqlite.ProjectRepository,
	sm *filesystem.StorageManager,
	logger *slog.Logger,
) *TemplateService {
	if logger == nil {
		logger = slog.Default()
	}
	return &TemplateService{
		templateRepo: templateRepo,
		projectRepo:  projectRepo,
		sm:           sm,
		logger:       logger,
	}
}

// CreateTemplate validates input and persists a new template to SQLite and disk.
func (s *TemplateService) CreateTemplate(ctx context.Context, input domain.CreateTemplateInput) (*domain.Template, error) {
	cleanName, err := domain.ValidateTemplateName(input.Name)
	if err != nil {
		return nil, err
	}

	if err := filesystem.ValidateID(input.ProjectID); err != nil {
		return nil, domain.ErrInvalidProjectID
	}

	// Verify project exists
	if _, err := s.projectRepo.GetByID(ctx, input.ProjectID); err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	docJSON := input.DocumentJSON
	if docJSON == "" {
		doc := domain.NewDefaultA4LandscapeDocument()
		var err error
		docJSON, err = doc.ToJSON()
		if err != nil {
			return nil, fmt.Errorf("failed to generate default document json: %w", err)
		}
	} else {
		// Validate incoming document model
		if _, err := domain.DocumentFromJSON(docJSON); err != nil {
			return nil, fmt.Errorf("invalid template document: %w", err)
		}
	}

	templateID := uuid.New().String()
	now := time.Now().UTC()

	tmpl := &domain.Template{
		ID:            templateID,
		ProjectID:     input.ProjectID,
		Name:          cleanName,
		DocumentJSON:  docJSON,
		SchemaVersion: 1,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// 1. Insert into database
	if err := s.templateRepo.Create(ctx, tmpl); err != nil {
		return nil, fmt.Errorf("failed to save template in database: %w", err)
	}

	// 2. Sync to project directory
	if err := s.syncTemplateToDisk(tmpl); err != nil {
		s.logger.Warn("failed to sync template to disk file", "templateId", tmpl.ID, "error", err)
	}

	s.logger.Info("template created successfully", "templateId", tmpl.ID, "name", tmpl.Name)
	return tmpl, nil
}

// GetTemplate retrieves a template by ID.
func (s *TemplateService) GetTemplate(ctx context.Context, id string) (*domain.Template, error) {
	if err := filesystem.ValidateID(id); err != nil {
		return nil, domain.ErrInvalidTemplateID
	}
	return s.templateRepo.GetByID(ctx, id)
}

// ListTemplates retrieves all templates belonging to a project.
func (s *TemplateService) ListTemplates(ctx context.Context, projectID string) ([]domain.Template, error) {
	if err := filesystem.ValidateID(projectID); err != nil {
		return nil, domain.ErrInvalidProjectID
	}
	return s.templateRepo.ListByProject(ctx, projectID)
}

// UpdateTemplate updates a template's name and document JSON.
func (s *TemplateService) UpdateTemplate(ctx context.Context, id string, name string, docJSON string) (*domain.Template, error) {
	if err := filesystem.ValidateID(id); err != nil {
		return nil, domain.ErrInvalidTemplateID
	}

	cleanName, err := domain.ValidateTemplateName(name)
	if err != nil {
		return nil, err
	}

	// Validate document JSON schema
	doc, err := domain.DocumentFromJSON(docJSON)
	if err != nil {
		return nil, fmt.Errorf("invalid template document: %w", err)
	}

	tmpl, err := s.templateRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	tmpl.Name = cleanName
	tmpl.DocumentJSON = docJSON
	tmpl.SchemaVersion = doc.SchemaVersion
	tmpl.UpdatedAt = time.Now().UTC()

	// 1. Update in SQLite
	if err := s.templateRepo.Update(ctx, tmpl); err != nil {
		return nil, fmt.Errorf("failed to update template in database: %w", err)
	}

	// 2. Sync updated document to disk
	if err := s.syncTemplateToDisk(tmpl); err != nil {
		s.logger.Warn("failed to sync updated template to disk", "templateId", id, "error", err)
	}

	s.logger.Info("template updated successfully", "templateId", id, "name", tmpl.Name)
	return tmpl, nil
}

// DeleteTemplate removes a template from SQLite and filesystem.
func (s *TemplateService) DeleteTemplate(ctx context.Context, id string) error {
	if err := filesystem.ValidateID(id); err != nil {
		return domain.ErrInvalidTemplateID
	}

	tmpl, err := s.templateRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 1. Delete from database
	if err := s.templateRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete template from database: %w", err)
	}

	// 2. Delete file from disk
	projDir, err := s.sm.ProjectDir(tmpl.ProjectID)
	if err == nil {
		filePath := filepath.Join(projDir, "templates", tmpl.ID+".json")
		_ = os.Remove(filePath)
	}

	s.logger.Info("template deleted successfully", "templateId", id)
	return nil
}

func (s *TemplateService) syncTemplateToDisk(tmpl *domain.Template) error {
	projDir, err := s.sm.ProjectDir(tmpl.ProjectID)
	if err != nil {
		return err
	}
	templatesDir := filepath.Join(projDir, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		return err
	}

	filePath := filepath.Join(templatesDir, tmpl.ID+".json")
	return filesystem.AtomicWriteFile(filePath, []byte(tmpl.DocumentJSON), 0644)
}
