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
	"github.com/xuri/excelize/v2"
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

// ExportRecipientsToExcel exports all recipients for a project into a formatted Excel (.xlsx) file.
func (s *RecipientService) ExportRecipientsToExcel(ctx context.Context, projectID string, outputPath string) (string, error) {
	if err := filesystem.ValidateID(projectID); err != nil {
		return "", domain.ErrInvalidProjectID
	}

	proj, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return "", fmt.Errorf("gagal memuat data acara: %w", err)
	}

	recipients, err := s.repo.GetAllByProject(ctx, projectID)
	if err != nil {
		return "", fmt.Errorf("gagal memuat daftar penerima: %w", err)
	}

	f := excelize.NewFile()
	defer func() {
		_ = f.Close()
	}()

	sheetName := "Data Penerima"
	f.SetSheetName("Sheet1", sheetName)

	headers := []string{
		"No",
		"Nama Lengkap",
		"No. Sertifikat",
		"Tanggal Acara",
		"Tempat",
		"Alamat",
		"Organisasi / Instansi",
		"Email",
		"Tanggal Ditambahkan",
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Color: "#FFFFFF",
			Size:  11,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#1E3A8A"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#D1D5DB", Style: 1},
			{Type: "top", Color: "#D1D5DB", Style: 1},
			{Type: "bottom", Color: "#D1D5DB", Style: 1},
			{Type: "right", Color: "#D1D5DB", Style: 1},
		},
	})

	dataStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "#E5E7EB", Style: 1},
			{Type: "top", Color: "#E5E7EB", Style: 1},
			{Type: "bottom", Color: "#E5E7EB", Style: 1},
			{Type: "right", Color: "#E5E7EB", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Vertical: "center",
		},
	})

	centerStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "#E5E7EB", Style: 1},
			{Type: "top", Color: "#E5E7EB", Style: 1},
			{Type: "bottom", Color: "#E5E7EB", Style: 1},
			{Type: "right", Color: "#E5E7EB", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})

	_ = f.SetRowHeight(sheetName, 1, 28)

	for colIdx, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
		_ = f.SetCellValue(sheetName, cell, header)
		_ = f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	for i, r := range recipients {
		row := i + 2
		_ = f.SetRowHeight(sheetName, row, 22)

		eventDate := r.IssueDate
		if eventDate == "" {
			eventDate = proj.EventDate
		}

		_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), i+1)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), r.FullName)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), r.CertificateID)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), eventDate)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), proj.EventLocation)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), proj.EventAddress)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), r.Organization)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), r.Email)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), r.CreatedAt.Format("2006-01-02 15:04"))

		_ = f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), centerStyle)
		_ = f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), dataStyle)
		_ = f.SetCellStyle(sheetName, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), centerStyle)
		_ = f.SetCellStyle(sheetName, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), centerStyle)
		_ = f.SetCellStyle(sheetName, fmt.Sprintf("E%d", row), fmt.Sprintf("E%d", row), dataStyle)
		_ = f.SetCellStyle(sheetName, fmt.Sprintf("F%d", row), fmt.Sprintf("F%d", row), dataStyle)
		_ = f.SetCellStyle(sheetName, fmt.Sprintf("G%d", row), fmt.Sprintf("G%d", row), dataStyle)
		_ = f.SetCellStyle(sheetName, fmt.Sprintf("H%d", row), fmt.Sprintf("H%d", row), dataStyle)
		_ = f.SetCellStyle(sheetName, fmt.Sprintf("I%d", row), fmt.Sprintf("I%d", row), centerStyle)
	}

	_ = f.SetColWidth(sheetName, "A", "A", 6)
	_ = f.SetColWidth(sheetName, "B", "B", 30)
	_ = f.SetColWidth(sheetName, "C", "C", 20)
	_ = f.SetColWidth(sheetName, "D", "D", 16)
	_ = f.SetColWidth(sheetName, "E", "E", 25)
	_ = f.SetColWidth(sheetName, "F", "F", 35)
	_ = f.SetColWidth(sheetName, "G", "G", 25)
	_ = f.SetColWidth(sheetName, "H", "H", 25)
	_ = f.SetColWidth(sheetName, "I", "I", 20)

	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("gagal membuat folder output: %w", err)
	}

	if err := f.SaveAs(outputPath); err != nil {
		return "", fmt.Errorf("gagal menyimpan berkas excel: %w", err)
	}

	return outputPath, nil
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
