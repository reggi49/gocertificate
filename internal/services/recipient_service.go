package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"certificate-generator/internal/domain"
	"certificate-generator/internal/importer"
	csvparser "certificate-generator/internal/importer/csv"
	excelparser "certificate-generator/internal/importer/excel"
	"certificate-generator/internal/repository/sqlite"
	"certificate-generator/internal/storage/filesystem"

	"github.com/google/uuid"
)

// RecipientService coordinates recipient data import, validation, and CRUD operations.
type RecipientService struct {
	repo        sqlite.RecipientRepository
	projectRepo sqlite.ProjectRepository
	csvParser   *csvparser.CSVParser
	excelParser *excelparser.ExcelParser
	logger      *slog.Logger
}

// NewRecipientService constructs a new RecipientService.
func NewRecipientService(
	repo sqlite.RecipientRepository,
	projectRepo sqlite.ProjectRepository,
	logger *slog.Logger,
) *RecipientService {
	if logger == nil {
		logger = slog.Default()
	}
	return &RecipientService{
		repo:        repo,
		projectRepo: projectRepo,
		csvParser:   csvparser.NewCSVParser(),
		excelParser: excelparser.NewExcelParser(),
		logger:      logger,
	}
}

// InspectFile inspects file type, size, and available sheets (for Excel).
func (s *RecipientService) InspectFile(filePath string) (*importer.FileMetadata, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}
	if fileInfo.Size() > importer.MaxImportFileSize {
		return nil, importer.ErrFileTooLarge
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".csv":
		return &importer.FileMetadata{
			FileName: filepath.Base(filePath),
			FileSize: fileInfo.Size(),
			FileType: "csv",
		}, nil

	case ".xlsx":
		return s.excelParser.InspectWorkbook(filePath)

	case ".xlsm":
		return nil, importer.ErrMacroDetected

	default:
		return nil, importer.ErrUnsupportedFormat
	}
}

// PreviewFile reads headers and preview rows from CSV or XLSX.
func (s *RecipientService) PreviewFile(filePath string, sheetName string, maxRows int) (*importer.TablePreview, error) {
	if maxRows <= 0 {
		maxRows = 10
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".csv":
		return s.csvParser.Preview(filePath, maxRows)
	case ".xlsx":
		return s.excelParser.Preview(filePath, sheetName, maxRows)
	default:
		return nil, importer.ErrUnsupportedFormat
	}
}

// ValidateMapping validates table rows against the column mapping and returns validation summary & errors.
func (s *RecipientService) ValidateMapping(
	filePath string,
	sheetName string,
	mapping domain.ColumnMapping,
) (*domain.ImportSummary, error) {
	if strings.TrimSpace(mapping.FullNameColumn) == "" {
		return nil, domain.ErrMissingFullNameMapping
	}

	headers, rows, err := s.readAllTable(filePath, sheetName)
	if err != nil {
		return nil, err
	}

	headerMap := make(map[string]int)
	for i, h := range headers {
		headerMap[strings.TrimSpace(h)] = i
	}

	fullNameIdx, ok := headerMap[strings.TrimSpace(mapping.FullNameColumn)]
	if !ok {
		return nil, fmt.Errorf("full name column '%s' not found in headers", mapping.FullNameColumn)
	}

	emailIdx := getIndexOrMinusOne(headerMap, mapping.EmailColumn)
	orgIdx := getIndexOrMinusOne(headerMap, mapping.OrganizationColumn)
	eventIdx := getIndexOrMinusOne(headerMap, mapping.EventNameColumn)
	dateIdx := getIndexOrMinusOne(headerMap, mapping.IssueDateColumn)
	certIDIdx := getIndexOrMinusOne(headerMap, mapping.CertificateIDColumn)

	summary := &domain.ImportSummary{
		TotalRows: len(rows),
		Errors:    []domain.ImportRowError{},
	}

	seenCertIDs := make(map[string]int) // certID -> first row

	for idx, row := range rows {
		rowNumber := idx + 2 // 1-based, +1 for header
		hasError := false

		// 1. Full name validation (required)
		fullName := getCell(row, fullNameIdx)
		if strings.TrimSpace(fullName) == "" {
			summary.Errors = append(summary.Errors, domain.ImportRowError{
				RowNumber: rowNumber,
				Column:    mapping.FullNameColumn,
				Category:  "MISSING_REQUIRED",
				Message:   "Nama lengkap wajib diisi",
				RawValue:  "",
			})
			hasError = true
		}

		// 2. Email validation (if mapped and non-empty)
		if emailIdx != -1 {
			email := getCell(row, emailIdx)
			if email != "" {
				if err := domain.ValidateEmailString(email); err != nil {
					summary.Errors = append(summary.Errors, domain.ImportRowError{
						RowNumber: rowNumber,
						Column:    mapping.EmailColumn,
						Category:  "INVALID_EMAIL",
						Message:   "Format alamat email tidak valid",
						RawValue:  email,
					})
					hasError = true
				}
			}
		}

		// 3. Date validation (if mapped and non-empty)
		if dateIdx != -1 {
			dateStr := getCell(row, dateIdx)
			if dateStr != "" {
				if _, err := domain.NormalizeAndValidateDate(dateStr); err != nil {
					summary.Errors = append(summary.Errors, domain.ImportRowError{
						RowNumber: rowNumber,
						Column:    mapping.IssueDateColumn,
						Category:  "INVALID_DATE",
						Message:   "Format tanggal tidak valid",
						RawValue:  dateStr,
					})
					hasError = true
				}
			}
		}

		// 4. Duplicate Certificate ID within spreadsheet
		if certIDIdx != -1 {
			certID := getCell(row, certIDIdx)
			if certID != "" {
				if prevRow, exists := seenCertIDs[certID]; exists {
					summary.Errors = append(summary.Errors, domain.ImportRowError{
						RowNumber: rowNumber,
						Column:    mapping.CertificateIDColumn,
						Category:  "DUPLICATE_IN_FILE",
						Message:   fmt.Sprintf("Nomor sertifikat '%s' duplikat dengan baris %d", certID, prevRow),
						RawValue:  certID,
					})
					hasError = true
				} else {
					seenCertIDs[certID] = rowNumber
				}
			}
		}

		if hasError {
			summary.InvalidRows++
		} else {
			summary.ValidRows++
		}
	}

	_ = orgIdx
	_ = eventIdx

	return summary, nil
}

// ImportRecipients validates, converts, and inserts records into SQLite transactionally.
func (s *RecipientService) ImportRecipients(
	ctx context.Context,
	projectID string,
	filePath string,
	sheetName string,
	mapping domain.ColumnMapping,
	strategy domain.ConflictStrategy,
) (*domain.ImportSummary, error) {
	if err := filesystem.ValidateID(projectID); err != nil {
		return nil, domain.ErrInvalidProjectID
	}

	// Verify project exists
	if _, err := s.projectRepo.GetByID(ctx, projectID); err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	// Validate pre-import
	validationSummary, err := s.ValidateMapping(filePath, sheetName, mapping)
	if err != nil {
		return nil, err
	}

	// If strategy is fail and there are validation errors, return early with errors
	if strategy == domain.ConflictStrategyFail && validationSummary.InvalidRows > 0 {
		return validationSummary, fmt.Errorf("import aborted: %d invalid rows detected", validationSummary.InvalidRows)
	}

	headers, rows, err := s.readAllTable(filePath, sheetName)
	if err != nil {
		return nil, err
	}

	headerMap := make(map[string]int)
	for i, h := range headers {
		headerMap[strings.TrimSpace(h)] = i
	}

	fullNameIdx := headerMap[strings.TrimSpace(mapping.FullNameColumn)]
	emailIdx := getIndexOrMinusOne(headerMap, mapping.EmailColumn)
	orgIdx := getIndexOrMinusOne(headerMap, mapping.OrganizationColumn)
	eventIdx := getIndexOrMinusOne(headerMap, mapping.EventNameColumn)
	dateIdx := getIndexOrMinusOne(headerMap, mapping.IssueDateColumn)
	certIDIdx := getIndexOrMinusOne(headerMap, mapping.CertificateIDColumn)

	// Collect valid recipient entities
	var validRecipients []domain.Recipient
	now := time.Now().UTC()

	for _, row := range rows {
		fullName := getCell(row, fullNameIdx)
		if strings.TrimSpace(fullName) == "" {
			continue // skip invalid row without full name
		}

		email := ""
		if emailIdx != -1 {
			email = getCell(row, emailIdx)
		}

		org := ""
		if orgIdx != -1 {
			org = getCell(row, orgIdx)
		}

		eventName := ""
		if eventIdx != -1 {
			eventName = getCell(row, eventIdx)
		}

		issueDate := ""
		if dateIdx != -1 {
			dateStr := getCell(row, dateIdx)
			norm, err := domain.NormalizeAndValidateDate(dateStr)
			if err == nil && norm != "" {
				issueDate = norm
			} else {
				issueDate = dateStr
			}
		}

		certID := ""
		if certIDIdx != -1 {
			certID = getCell(row, certIDIdx)
		}

		// Handle Custom Fields mapping
		customFields := make(map[string]string)
		for fieldName, colName := range mapping.CustomFieldColumns {
			if idx, ok := headerMap[colName]; ok {
				val := getCell(row, idx)
				if val != "" {
					customFields[fieldName] = val
				}
			}
		}

		customFieldsJSON := ""
		if len(customFields) > 0 {
			bytes, _ := json.Marshal(customFields)
			customFieldsJSON = string(bytes)
		}

		rec := domain.Recipient{
			ID:               uuid.New().String(),
			ProjectID:        projectID,
			FullName:         fullName,
			Email:            email,
			Organization:     org,
			EventName:        eventName,
			IssueDate:        issueDate,
			CertificateID:    certID,
			CustomFieldsJSON: customFieldsJSON,
			CreatedAt:        now,
			UpdatedAt:        now,
		}

		validRecipients = append(validRecipients, rec)
	}

	// Insert transactionally
	summary, err := s.repo.BatchCreate(ctx, projectID, validRecipients, strategy)
	if err != nil {
		s.logger.Error("failed to batch insert recipients", "projectId", projectID, "error", err)
		return summary, err
	}

	// Preserve validation errors from pre-import check
	summary.Errors = append(validationSummary.Errors, summary.Errors...)
	summary.InvalidRows = validationSummary.InvalidRows
	summary.TotalRows = len(rows)

	s.logger.Info("recipients imported successfully",
		"projectId", projectID,
		"imported", summary.ImportedRows,
		"skipped", summary.SkippedRows,
		"updated", summary.UpdatedRows,
		"errors", len(summary.Errors),
	)

	return summary, nil
}

// Single Recipient CRUD Operations

func (s *RecipientService) CreateRecipient(ctx context.Context, rec domain.Recipient) (*domain.Recipient, error) {
	if err := filesystem.ValidateID(rec.ProjectID); err != nil {
		return nil, domain.ErrInvalidProjectID
	}
	if err := domain.ValidateRecipient(&rec); err != nil {
		return nil, err
	}

	rec.ID = uuid.New().String()
	rec.CreatedAt = time.Now().UTC()
	rec.UpdatedAt = rec.CreatedAt

	if err := s.repo.Create(ctx, &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

func (s *RecipientService) GetRecipient(ctx context.Context, id string) (*domain.Recipient, error) {
	if err := filesystem.ValidateID(id); err != nil {
		return nil, domain.ErrRecipientNotFound
	}
	return s.repo.GetByID(ctx, id)
}

func (s *RecipientService) UpdateRecipient(ctx context.Context, rec domain.Recipient) (*domain.Recipient, error) {
	if err := filesystem.ValidateID(rec.ID); err != nil {
		return nil, domain.ErrRecipientNotFound
	}
	if err := domain.ValidateRecipient(&rec); err != nil {
		return nil, err
	}
	rec.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

func (s *RecipientService) DeleteRecipient(ctx context.Context, id string) error {
	if err := filesystem.ValidateID(id); err != nil {
		return domain.ErrRecipientNotFound
	}
	return s.repo.Delete(ctx, id)
}

func (s *RecipientService) DeleteRecipients(ctx context.Context, ids []string) error {
	for _, id := range ids {
		if err := filesystem.ValidateID(id); err != nil {
			return fmt.Errorf("invalid recipient id in batch: %s", id)
		}
	}
	return s.repo.DeleteBatch(ctx, ids)
}

func (s *RecipientService) DeleteAllRecipients(ctx context.Context, projectID string) error {
	if err := filesystem.ValidateID(projectID); err != nil {
		return domain.ErrInvalidProjectID
	}
	return s.repo.DeleteAllByProject(ctx, projectID)
}

func (s *RecipientService) ListRecipients(ctx context.Context, filter domain.RecipientFilter) (*domain.PaginatedRecipients, error) {
	if err := filesystem.ValidateID(filter.ProjectID); err != nil {
		return nil, domain.ErrInvalidProjectID
	}
	return s.repo.List(ctx, filter)
}

func (s *RecipientService) CountRecipients(ctx context.Context, projectID string) (int, error) {
	if err := filesystem.ValidateID(projectID); err != nil {
		return 0, domain.ErrInvalidProjectID
	}
	return s.repo.Count(ctx, projectID)
}

// Helpers

func (s *RecipientService) readAllTable(filePath string, sheetName string) ([]string, [][]string, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".csv":
		return s.csvParser.ReadAll(filePath)
	case ".xlsx":
		return s.excelParser.ReadAll(filePath, sheetName)
	default:
		return nil, nil, importer.ErrUnsupportedFormat
	}
}

func getIndexOrMinusOne(m map[string]int, key string) int {
	clean := strings.TrimSpace(key)
	if clean == "" {
		return -1
	}
	if idx, ok := m[clean]; ok {
		return idx
	}
	return -1
}

func getCell(row []string, idx int) string {
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}
