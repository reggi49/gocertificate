package importer

import "errors"

const (
	// MaxImportFileSize limits maximum uploaded file to 50 MB to prevent denial of service
	MaxImportFileSize = 50 * 1024 * 1024
)

var (
	ErrFileTooLarge      = errors.New("file size exceeds maximum allowed limit (50 MB)")
	ErrUnsupportedFormat = errors.New("unsupported file format: only CSV and XLSX files are accepted")
	ErrMacroDetected     = errors.New("security alert: Excel files with macros (.xlsm) or embedded scripts are rejected")
	ErrEmptyFile         = errors.New("the imported file contains no readable data")
	ErrWorksheetNotFound = errors.New("selected worksheet was not found in workbook")
)

// FileMetadata contains basic information about an import file.
type FileMetadata struct {
	FileName   string   `json:"fileName"`
	FileSize   int64    `json:"fileSize"`
	FileType   string   `json:"fileType"`             // "csv" or "xlsx"
	Worksheets []string `json:"worksheets,omitempty"` // For XLSX
}

// TablePreview contains header detection and initial rows for column mapping.
type TablePreview struct {
	Headers   []string   `json:"headers"`
	Rows      [][]string `json:"rows"`
	TotalRows int        `json:"totalRows"`
}
