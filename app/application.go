package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"certificate-generator/internal/domain"
	certgen "certificate-generator/internal/generator/certificate"
	"certificate-generator/internal/importer"
	"certificate-generator/internal/renderer/pdf"
	"certificate-generator/internal/repository/sqlite"
	"certificate-generator/internal/services"
	"certificate-generator/internal/storage/backup"
	"certificate-generator/internal/storage/filesystem"

	wailsApp "github.com/wailsapp/wails/v3/pkg/application"
)

// Config holds runtime configuration paths.
type Config struct {
	DataDir string
	FontDir string
}

// App encapsulates desktop application lifecycle and exposed services.
type App struct {
	Config           Config
	DB               *sql.DB
	Logger           *slog.Logger
	Renderer         pdf.Renderer
	StorageManager   *filesystem.StorageManager
	ProjectService   *services.ProjectService
	SettingsService  *services.SettingsService
	TemplateService  *services.TemplateService
	RecipientService *services.RecipientService
	JobService       *services.JobService
	BackupManager    *backup.BackupManager
}

// DefaultDataDirectory returns standard OS user application support path.
func DefaultDataDirectory() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		userHome, _ := os.UserHomeDir()
		return filepath.Join(userHome, ".certificate-generator")
	}
	return filepath.Join(configDir, "CertificateGenerator")
}

// NewApp initializes application storage, database, repositories, and services.
func NewApp(cfg Config) (*App, error) {
	if cfg.DataDir == "" {
		cfg.DataDir = DefaultDataDirectory()
	}

	sm, err := filesystem.NewStorageManager(cfg.DataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage manager: %w", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	dbPath := filepath.Join(sm.DatabaseDir(), "app.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	ctx := context.Background()
	if err := sqlite.Migrate(ctx, db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("database migration failed: %w", err)
	}

	pRepo := sqlite.NewProjectRepository(db)
	sRepo := sqlite.NewSettingsRepository(db)
	tRepo := sqlite.NewTemplateRepository(db)
	rRepo := sqlite.NewRecipientRepository(db)
	jRepo := sqlite.NewJobRepository(db)

	renderer := pdf.NewGoPDFRenderer(cfg.FontDir)

	projectSvc := services.NewProjectService(pRepo, sm, logger)
	settingsSvc := services.NewSettingsService(sRepo)
	templateSvc := services.NewTemplateService(tRepo, pRepo, sm, logger)
	recipientSvc := services.NewRecipientService(rRepo, pRepo, logger)

	generator := certgen.NewBatchGenerator(renderer, jRepo, rRepo, tRepo, pRepo, sm, logger)
	jobSvc := services.NewJobService(generator, jRepo, sm, logger)

	backupMgr := backup.NewBackupManager(pRepo, tRepo, rRepo, db, sm, logger)

	// Clean up any stale interrupted jobs from past sudden exits
	_, _ = jobSvc.CheckInterruptedJobs(ctx)

	return &App{
		Config:           cfg,
		DB:               db,
		Logger:           logger,
		Renderer:         renderer,
		StorageManager:   sm,
		ProjectService:   projectSvc,
		SettingsService:  settingsSvc,
		TemplateService:  templateSvc,
		RecipientService: recipientSvc,
		JobService:       jobSvc,
		BackupManager:    backupMgr,
	}, nil
}

// Close gracefully releases application resources.
func (a *App) Close() error {
	if a.DB != nil {
		return a.DB.Close()
	}
	return nil
}

// RenderPreviewPDF handles on-demand rendering request from the UI.
func (a *App) RenderPreviewPDF(docJSON string, variables map[string]string) ([]byte, error) {
	doc, err := domain.DocumentFromJSON(docJSON)
	if err != nil {
		return nil, fmt.Errorf("invalid document json: %w", err)
	}
	assetDir := filepath.Join(a.Config.DataDir, "assets")
	return a.Renderer.Render(doc, variables, assetDir)
}

// Project Service Wrappers for Wails Service Binding

func (a *App) CreateProject(input domain.CreateProjectInput) (*domain.Project, error) {
	return a.ProjectService.CreateProject(context.Background(), input)
}

func (a *App) GetProject(id string) (*domain.Project, error) {
	return a.ProjectService.GetProject(context.Background(), id)
}

func (a *App) ListProjects(includeArchived bool) ([]domain.Project, error) {
	return a.ProjectService.ListProjects(context.Background(), includeArchived)
}

func (a *App) RecentProjects(limit int) ([]domain.Project, error) {
	return a.ProjectService.RecentProjects(context.Background(), limit)
}

func (a *App) RenameProject(id string, newName, description string) (*domain.Project, error) {
	return a.ProjectService.RenameProject(context.Background(), id, newName, description)
}

func (a *App) DuplicateProject(input domain.DuplicateProjectInput) (*domain.Project, error) {
	return a.ProjectService.DuplicateProject(context.Background(), input)
}

func (a *App) ArchiveProject(id string, archive bool) error {
	return a.ProjectService.ArchiveProject(context.Background(), id, archive)
}

func (a *App) DeleteProject(id string) error {
	return a.ProjectService.DeleteProject(context.Background(), id)
}

func (a *App) GetStorageInfo() (*domain.StorageInfo, error) {
	return a.ProjectService.GetStorageInfo(context.Background())
}

// Settings Service Wrappers

func (a *App) GetSettings() (*domain.AppSettings, error) {
	return a.SettingsService.GetSettings(context.Background())
}

func (a *App) UpdateSettings(settings domain.AppSettings) error {
	return a.SettingsService.UpdateSettings(context.Background(), settings)
}

// Template Service Wrappers for Wails Service Binding

func (a *App) CreateTemplate(input domain.CreateTemplateInput) (*domain.Template, error) {
	return a.TemplateService.CreateTemplate(context.Background(), input)
}

func (a *App) GetTemplate(id string) (*domain.Template, error) {
	return a.TemplateService.GetTemplate(context.Background(), id)
}

func (a *App) ListTemplates(projectID string) ([]domain.Template, error) {
	return a.TemplateService.ListTemplates(context.Background(), projectID)
}

func (a *App) UpdateTemplate(id string, name string, docJSON string) (*domain.Template, error) {
	return a.TemplateService.UpdateTemplate(context.Background(), id, name, docJSON)
}

func (a *App) DeleteTemplate(id string) error {
	return a.TemplateService.DeleteTemplate(context.Background(), id)
}

// Recipient Service Wrappers for Wails Service Binding

func (a *App) InspectImportFile(filePath string) (*importer.FileMetadata, error) {
	return a.RecipientService.InspectFile(filePath)
}

func (a *App) PreviewImportFile(filePath string, sheetName string, maxRows int) (*importer.TablePreview, error) {
	return a.RecipientService.PreviewFile(filePath, sheetName, maxRows)
}

func (a *App) ValidateImportMapping(filePath string, sheetName string, mapping domain.ColumnMapping) (*domain.ImportSummary, error) {
	return a.RecipientService.ValidateMapping(filePath, sheetName, mapping)
}

func (a *App) ImportRecipients(projectID string, filePath string, sheetName string, mapping domain.ColumnMapping, strategy domain.ConflictStrategy) (*domain.ImportSummary, error) {
	return a.RecipientService.ImportRecipients(context.Background(), projectID, filePath, sheetName, mapping, strategy)
}

func (a *App) CreateRecipient(rec domain.Recipient) (*domain.Recipient, error) {
	return a.RecipientService.CreateRecipient(context.Background(), rec)
}

func (a *App) GetRecipient(id string) (*domain.Recipient, error) {
	return a.RecipientService.GetRecipient(context.Background(), id)
}

func (a *App) UpdateRecipient(rec domain.Recipient) (*domain.Recipient, error) {
	return a.RecipientService.UpdateRecipient(context.Background(), rec)
}

func (a *App) DeleteRecipient(id string) error {
	return a.RecipientService.DeleteRecipient(context.Background(), id)
}

func (a *App) DeleteRecipients(ids []string) error {
	return a.RecipientService.DeleteRecipients(context.Background(), ids)
}

func (a *App) DeleteAllRecipients(projectID string) error {
	return a.RecipientService.DeleteAllRecipients(context.Background(), projectID)
}

func (a *App) ListRecipients(filter domain.RecipientFilter) (*domain.PaginatedRecipients, error) {
	return a.RecipientService.ListRecipients(context.Background(), filter)
}

func (a *App) CountRecipients(projectID string) (int, error) {
	return a.RecipientService.CountRecipients(context.Background(), projectID)
}

// Batch Generation Job Service Wrappers for Wails Service Binding

func (a *App) StartBatchGeneration(cfg domain.BatchConfig) (*domain.GenerationJob, error) {
	progressCb := func(ev domain.ProgressEvent) {
		if app := wailsApp.Get(); app != nil && app.Event != nil {
			app.Event.Emit("batch:progress", ev)
		}
	}
	return a.JobService.StartBatchGeneration(context.Background(), cfg, progressCb)
}

func (a *App) CancelBatchJob(jobID string) error {
	return a.JobService.CancelJob(jobID)
}

func (a *App) RetryFailedBatchItems(jobID string) (*domain.GenerationJob, error) {
	progressCb := func(ev domain.ProgressEvent) {
		if app := wailsApp.Get(); app != nil && app.Event != nil {
			app.Event.Emit("batch:progress", ev)
		}
	}
	return a.JobService.RetryFailedItems(context.Background(), jobID, progressCb)
}

func (a *App) GetBatchJob(id string) (*domain.GenerationJob, error) {
	return a.JobService.GetJob(context.Background(), id)
}

func (a *App) ListBatchJobs(projectID string) ([]domain.GenerationJob, error) {
	return a.JobService.ListJobsByProject(context.Background(), projectID)
}

func (a *App) ListBatchJobItems(jobID string) ([]domain.GenerationJobItem, error) {
	return a.JobService.ListJobItems(context.Background(), jobID)
}

// Backup & Restore Service Wrappers for Wails Service Binding

func (a *App) BackupProject(projectID string, outputPath string) (*domain.BackupManifest, error) {
	return a.BackupManager.BackupProject(context.Background(), projectID, outputPath)
}

func (a *App) BackupAll(outputPath string) (*domain.BackupManifest, error) {
	return a.BackupManager.BackupAll(context.Background(), outputPath)
}

func (a *App) RestoreProject(opts domain.RestoreOptions) (*domain.RestoreResult, error) {
	return a.BackupManager.RestoreProject(context.Background(), opts)
}

// Desktop Platform File and Directory Openers

func (a *App) OpenFile(filePath string) error {
	filePath = filepath.Clean(filePath)
	if _, err := os.Stat(filePath); err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", filePath)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", filePath)
	default:
		cmd = exec.Command("xdg-open", filePath)
	}
	return cmd.Start()
}

func (a *App) OpenDirectory(dirPath string) error {
	dirPath = filepath.Clean(dirPath)
	fi, err := os.Stat(dirPath)
	if err != nil {
		return fmt.Errorf("path not found: %w", err)
	}
	if !fi.IsDir() {
		dirPath = filepath.Dir(dirPath)
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", dirPath)
	case "windows":
		cmd = exec.Command("explorer", dirPath)
	default:
		cmd = exec.Command("xdg-open", dirPath)
	}
	return cmd.Start()
}
