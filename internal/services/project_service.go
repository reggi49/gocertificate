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

// ProjectService coordinates business workflows between SQLite repository and StorageManager.
type ProjectService struct {
	repo     sqlite.ProjectRepository
	tmplRepo sqlite.TemplateRepository
	sm       *filesystem.StorageManager
	logger   *slog.Logger
}

// NewProjectService constructs a new ProjectService.
func NewProjectService(repo sqlite.ProjectRepository, sm *filesystem.StorageManager, logger *slog.Logger) *ProjectService {
	if logger == nil {
		logger = slog.Default()
	}
	return &ProjectService{
		repo:   repo,
		sm:     sm,
		logger: logger,
	}
}

// SetTemplateRepository attaches the template repository to enable automatic default template seeding.
func (s *ProjectService) SetTemplateRepository(tmplRepo sqlite.TemplateRepository) {
	s.tmplRepo = tmplRepo
}

// CreateProject creates a project record and corresponding filesystem structure with rollback on failure.
func (s *ProjectService) CreateProject(ctx context.Context, input domain.CreateProjectInput) (*domain.Project, error) {
	cleanInput, err := domain.ValidateCreateProjectInput(input)
	if err != nil {
		return nil, err
	}

	projectID := uuid.New().String()
	now := time.Now().UTC()

	project := &domain.Project{
		ID:            projectID,
		Name:          cleanInput.Name,
		Description:   cleanInput.Description,
		EventDate:     cleanInput.EventDate,
		EventLocation: cleanInput.EventLocation,
		EventAddress:  cleanInput.EventAddress,
		CreatedAt:     now,
		UpdatedAt:     now,
		IsArchived:    false,
	}

	// 1. Scaffold project directory
	_, err = s.sm.CreateProjectDirectory(projectID)
	if err != nil {
		s.logger.Error("failed to create project directory", "projectId", projectID, "error", err)
		return nil, fmt.Errorf("failed to create project directory: %w", err)
	}

	// 2. Write project.json manifest
	manifest := domain.ProjectManifest{
		SchemaVersion: 1,
		ID:            projectID,
		Name:          project.Name,
		Description:   project.Description,
		EventDate:     project.EventDate,
		EventLocation: project.EventLocation,
		EventAddress:  project.EventAddress,
		CreatedAt:     project.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     project.UpdatedAt.Format(time.RFC3339),
	}
	if err := s.sm.WriteProjectManifest(projectID, manifest); err != nil {
		// Rollback directory creation
		_ = s.sm.DeleteProjectDirectory(projectID)
		s.logger.Error("failed to write project manifest, rolled back directory", "projectId", projectID, "error", err)
		return nil, fmt.Errorf("failed to write project manifest: %w", err)
	}

	// 3. Insert record into database
	if err := s.repo.Create(ctx, project); err != nil {
		// Compensating rollback: delete created directory
		_ = s.sm.DeleteProjectDirectory(projectID)
		s.logger.Error("failed to insert project to database, rolled back filesystem", "projectId", projectID, "error", err)
		return nil, fmt.Errorf("failed to save project to database: %w", err)
	}

	// 4. Automatically seed default Janaiz template into project
	if s.tmplRepo != nil {
		defaultDocJSON := domain.DefaultJanaizDocumentJSON()
		defaultTmpl := &domain.Template{
			ID:            uuid.New().String(),
			ProjectID:     projectID,
			Name:          "Template Sertifikat Janaiz",
			DocumentJSON:  defaultDocJSON,
			SchemaVersion: 1,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if err := s.tmplRepo.Create(ctx, defaultTmpl); err != nil {
			s.logger.Warn("failed to seed default template", "projectId", projectID, "error", err)
		} else {
			// Also sync to disk
			templatesDir := filepath.Join(s.sm.ProjectsDir(), projectID, "templates")
			_ = os.MkdirAll(templatesDir, 0755)
			_ = filesystem.AtomicWriteFile(filepath.Join(templatesDir, defaultTmpl.ID+".json"), []byte(defaultDocJSON), 0644)
		}
	}

	s.logger.Info("project created successfully", "projectId", projectID, "name", project.Name)
	return project, nil
}

// GetProject retrieves project details by ID.
func (s *ProjectService) GetProject(ctx context.Context, id string) (*domain.Project, error) {
	if err := filesystem.ValidateID(id); err != nil {
		return nil, domain.ErrInvalidProjectID
	}
	return s.repo.GetByID(ctx, id)
}

// ListProjects retrieves all projects.
func (s *ProjectService) ListProjects(ctx context.Context, includeArchived bool) ([]domain.Project, error) {
	return s.repo.List(ctx, includeArchived)
}

// RecentProjects retrieves up to limit recent active projects.
func (s *ProjectService) RecentProjects(ctx context.Context, limit int) ([]domain.Project, error) {
	return s.repo.Recent(ctx, limit)
}

// RenameProject renames a project and updates both DB and project.json.
func (s *ProjectService) RenameProject(ctx context.Context, id string, newName, description string) (*domain.Project, error) {
	if err := filesystem.ValidateID(id); err != nil {
		return nil, domain.ErrInvalidProjectID
	}

	cleanInput, err := domain.ValidateUpdateProjectInput(domain.UpdateProjectInput{
		Name:        newName,
		Description: description,
	})
	if err != nil {
		return nil, err
	}

	project, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	project.Name = cleanInput.Name
	project.Description = cleanInput.Description
	project.UpdatedAt = time.Now().UTC()

	// Update DB
	if err := s.repo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update project in database: %w", err)
	}

	// Update manifest
	manifest := domain.ProjectManifest{
		SchemaVersion: 1,
		ID:            project.ID,
		Name:          project.Name,
		Description:   project.Description,
		CreatedAt:     project.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     project.UpdatedAt.Format(time.RFC3339),
	}
	if err := s.sm.WriteProjectManifest(project.ID, manifest); err != nil {
		s.logger.Warn("failed to update manifest file on disk", "projectId", id, "error", err)
	}

	s.logger.Info("project renamed successfully", "projectId", id, "newName", project.Name)
	return project, nil
}

// DuplicateProject duplicates a project and its folder structure.
func (s *ProjectService) DuplicateProject(ctx context.Context, input domain.DuplicateProjectInput) (*domain.Project, error) {
	if err := filesystem.ValidateID(input.SourceProjectID); err != nil {
		return nil, domain.ErrInvalidProjectID
	}

	cleanName, err := domain.ValidateProjectName(input.NewName)
	if err != nil {
		return nil, err
	}

	sourceProj, err := s.repo.GetByID(ctx, input.SourceProjectID)
	if err != nil {
		return nil, fmt.Errorf("source project not found: %w", err)
	}

	newID := uuid.New().String()
	now := time.Now().UTC()

	newProj := &domain.Project{
		ID:            newID,
		Name:          cleanName,
		Description:   sourceProj.Description,
		EventDate:     sourceProj.EventDate,
		EventLocation: sourceProj.EventLocation,
		EventAddress:  sourceProj.EventAddress,
		CreatedAt:     now,
		UpdatedAt:     now,
		IsArchived:    false,
	}

	// 1. Copy directory
	if err := s.sm.DuplicateProjectDirectory(sourceProj.ID, newID); err != nil {
		return nil, fmt.Errorf("failed to copy project directory: %w", err)
	}

	// 2. Write new manifest
	newManifest := domain.ProjectManifest{
		SchemaVersion: 1,
		ID:            newID,
		Name:          newProj.Name,
		Description:   newProj.Description,
		EventDate:     newProj.EventDate,
		EventLocation: newProj.EventLocation,
		EventAddress:  newProj.EventAddress,
		CreatedAt:     newProj.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     newProj.UpdatedAt.Format(time.RFC3339),
	}
	if err := s.sm.WriteProjectManifest(newID, newManifest); err != nil {
		_ = s.sm.DeleteProjectDirectory(newID)
		return nil, fmt.Errorf("failed to write duplicated manifest: %w", err)
	}

	// 3. Save to database
	if err := s.repo.Create(ctx, newProj); err != nil {
		_ = s.sm.DeleteProjectDirectory(newID)
		return nil, fmt.Errorf("failed to save duplicated project in database: %w", err)
	}

	s.logger.Info("project duplicated successfully", "sourceId", sourceProj.ID, "newId", newID, "name", newProj.Name)
	return newProj, nil
}

// ArchiveProject toggles archive state of a project.
func (s *ProjectService) ArchiveProject(ctx context.Context, id string, archive bool) error {
	if err := filesystem.ValidateID(id); err != nil {
		return domain.ErrInvalidProjectID
	}
	return s.repo.SetArchived(ctx, id, archive)
}

// DeleteProject deletes project from DB and removes its directory on disk.
func (s *ProjectService) DeleteProject(ctx context.Context, id string) error {
	if err := filesystem.ValidateID(id); err != nil {
		return domain.ErrInvalidProjectID
	}

	// 1. Delete from DB
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete project from database: %w", err)
	}

	// 2. Delete filesystem directory
	if err := s.sm.DeleteProjectDirectory(id); err != nil {
		s.logger.Warn("failed to delete project folder from disk", "projectId", id, "error", err)
	}

	s.logger.Info("project deleted successfully", "projectId", id)
	return nil
}

// GetStorageInfo returns storage statistics.
func (s *ProjectService) GetStorageInfo(ctx context.Context) (*domain.StorageInfo, error) {
	total, active, archived, err := s.repo.Count(ctx)
	if err != nil {
		return nil, err
	}

	dbPath := filepath.Join(s.sm.DatabaseDir(), "app.db")
	var dbSize int64
	if info, err := os.Stat(dbPath); err == nil {
		dbSize = info.Size()
	}

	projectsSize, _ := filesystem.CalculateDirSize(s.sm.ProjectsDir())

	return &domain.StorageInfo{
		DataDir:          s.sm.GetRootDir(),
		DatabaseSizeByte: dbSize,
		ProjectsSizeByte: projectsSize,
		TotalProjects:    total,
		ActiveProjects:   active,
		ArchivedProjects: archived,
	}, nil
}
