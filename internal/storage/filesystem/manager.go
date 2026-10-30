package filesystem

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"certificate-generator/internal/domain"
)

var (
	ErrPathTraversal = errors.New("access denied: path traversal detected outside managed root")
)

// StorageManager manages application directories, project file trees, and atomic writes.
type StorageManager struct {
	rootDir string
}

// NewStorageManager constructs a new StorageManager rooted at rootDir.
func NewStorageManager(rootDir string) (*StorageManager, error) {
	cleanRoot := filepath.Clean(rootDir)
	if err := os.MkdirAll(cleanRoot, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage root directory: %w", err)
	}

	sm := &StorageManager{rootDir: cleanRoot}
	if err := sm.ensureBaseDirectories(); err != nil {
		return nil, err
	}
	return sm, nil
}

// GetRootDir returns the absolute root directory path.
func (sm *StorageManager) GetRootDir() string {
	return sm.rootDir
}

// DatabaseDir returns the path to the database directory.
func (sm *StorageManager) DatabaseDir() string {
	return filepath.Join(sm.rootDir, "database")
}

// ProjectsDir returns the path to the projects directory.
func (sm *StorageManager) ProjectsDir() string {
	return filepath.Join(sm.rootDir, "projects")
}

// BackupsDir returns the path to the backups directory.
func (sm *StorageManager) BackupsDir() string {
	return filepath.Join(sm.rootDir, "backups")
}

// LogsDir returns the path to the logs directory.
func (sm *StorageManager) LogsDir() string {
	return filepath.Join(sm.rootDir, "logs")
}

// CacheDir returns the path to the cache directory.
func (sm *StorageManager) CacheDir() string {
	return filepath.Join(sm.rootDir, "cache")
}

func (sm *StorageManager) ensureBaseDirectories() error {
	dirs := []string{
		sm.DatabaseDir(),
		sm.ProjectsDir(),
		sm.BackupsDir(),
		sm.LogsDir(),
		sm.CacheDir(),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create base directory %s: %w", dir, err)
		}
	}
	return nil
}

// SafeResolve resolves and validates that a relative path stays within sm.rootDir.
func (sm *StorageManager) SafeResolve(subpath string) (string, error) {
	cleanTarget := filepath.Clean(filepath.Join(sm.rootDir, subpath))
	if cleanTarget != sm.rootDir && !strings.HasPrefix(cleanTarget, sm.rootDir+string(filepath.Separator)) {
		return "", ErrPathTraversal
	}
	return cleanTarget, nil
}

// ProjectDir returns the validated directory path for a project.
func (sm *StorageManager) ProjectDir(projectID string) (string, error) {
	if err := ValidateID(projectID); err != nil {
		return "", fmt.Errorf("invalid project id: %w", err)
	}
	return sm.SafeResolve(filepath.Join("projects", projectID))
}

// CreateProjectDirectory scaffolds the standard directory layout for a project.
func (sm *StorageManager) CreateProjectDirectory(projectID string) (string, error) {
	projDir, err := sm.ProjectDir(projectID)
	if err != nil {
		return "", err
	}

	subdirs := []string{
		"templates",
		filepath.Join("assets", "backgrounds"),
		filepath.Join("assets", "logos"),
		filepath.Join("assets", "signatures"),
		filepath.Join("assets", "fonts"),
		"exports",
	}

	for _, sub := range subdirs {
		dirPath := filepath.Join(projDir, sub)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			_ = os.RemoveAll(projDir) // cleanup on failure
			return "", fmt.Errorf("failed to scaffold project directory %s: %w", sub, err)
		}
	}

	return projDir, nil
}

// DeleteProjectDirectory safely removes the project's folder.
func (sm *StorageManager) DeleteProjectDirectory(projectID string) error {
	projDir, err := sm.ProjectDir(projectID)
	if err != nil {
		return err
	}
	if _, err := os.Stat(projDir); os.IsNotExist(err) {
		return nil
	}
	return os.RemoveAll(projDir)
}

// WriteProjectManifest writes the project manifest atomically using a temp file.
func (sm *StorageManager) WriteProjectManifest(projectID string, manifest domain.ProjectManifest) error {
	projDir, err := sm.ProjectDir(projectID)
	if err != nil {
		return err
	}

	manifestPath := filepath.Join(projDir, "project.json")
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal project manifest: %w", err)
	}

	return AtomicWriteFile(manifestPath, data, 0644)
}

// ReadProjectManifest reads the project manifest.
func (sm *StorageManager) ReadProjectManifest(projectID string) (*domain.ProjectManifest, error) {
	projDir, err := sm.ProjectDir(projectID)
	if err != nil {
		return nil, err
	}

	manifestPath := filepath.Join(projDir, "project.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read project manifest: %w", err)
	}

	var manifest domain.ProjectManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal project manifest: %w", err)
	}

	return &manifest, nil
}

// DuplicateProjectDirectory copies the source project directory tree to a new project directory.
func (sm *StorageManager) DuplicateProjectDirectory(sourceProjectID, newProjectID string) error {
	srcDir, err := sm.ProjectDir(sourceProjectID)
	if err != nil {
		return err
	}
	destDir, err := sm.ProjectDir(newProjectID)
	if err != nil {
		return err
	}

	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return fmt.Errorf("source project directory does not exist: %s", sourceProjectID)
	}

	// Copy directory tree
	err = copyDir(srcDir, destDir)
	if err != nil {
		_ = os.RemoveAll(destDir)
		return fmt.Errorf("failed to copy project directory: %w", err)
	}

	return nil
}

// CalculateDirSize returns the total bytes occupied by a directory.
func CalculateDirSize(dirPath string) (int64, error) {
	var size int64
	err := filepath.Walk(dirPath, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

// AtomicWriteFile writes data to a temporary file and atomically renames it to targetPath.
func AtomicWriteFile(targetPath string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(targetPath)
	tempFile, err := os.CreateTemp(dir, "atomic-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tempName := tempFile.Name()
	defer func() {
		_ = os.Remove(tempName) // no-op if renamed
	}()

	if _, err := tempFile.Write(data); err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("failed to write data to temp file: %w", err)
	}

	if err := tempFile.Sync(); err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("failed to sync temp file: %w", err)
	}

	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	if err := os.Chmod(tempName, perm); err != nil {
		return fmt.Errorf("failed to chmod temp file: %w", err)
	}

	if err := os.Rename(tempName, targetPath); err != nil {
		return fmt.Errorf("failed to atomically rename temp file to %s: %w", targetPath, err)
	}

	return nil
}

// copyDir recursively copies a directory tree.
func copyDir(src string, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
	}()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

// DefaultExportDirectory returns a clean user-accessible directory (e.g. ~/Downloads/Sertifikat/<ProjectName>).
func DefaultExportDirectory(projectName string) string {
	userHome, err := os.UserHomeDir()
	if err != nil {
		userHome = "."
	}

	downloads := filepath.Join(userHome, "Downloads")
	baseDir := filepath.Join(downloads, "Sertifikat")
	if _, err := os.Stat(downloads); err != nil {
		// Fallback to Documents if Downloads directory doesn't exist
		docs := filepath.Join(userHome, "Documents")
		if _, err := os.Stat(docs); err == nil {
			baseDir = filepath.Join(docs, "Sertifikat")
		} else {
			baseDir = filepath.Join(userHome, "Sertifikat")
		}
	}

	if projectName != "" {
		cleanProject := SanitizeFilename(projectName)
		if cleanProject != "" {
			return filepath.Join(baseDir, cleanProject)
		}
	}
	return baseDir
}
