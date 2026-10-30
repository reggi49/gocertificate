package excel

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"certificate-generator/internal/importer"

	"github.com/xuri/excelize/v2"
)

// ExcelParser extracts recipient tabular data from XLSX workbooks.
type ExcelParser struct{}

// NewExcelParser constructs an ExcelParser.
func NewExcelParser() *ExcelParser {
	return &ExcelParser{}
}

// InspectWorkbook inspects worksheets and basic metadata from an Excel workbook.
func (p *ExcelParser) InspectWorkbook(filePath string) (*importer.FileMetadata, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}
	if fileInfo.Size() > importer.MaxImportFileSize {
		return nil, importer.ErrFileTooLarge
	}

	// Security check: Reject macro-enabled files
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext == ".xlsm" || ext == ".xltm" {
		return nil, importer.ErrMacroDetected
	}

	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel workbook: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, importer.ErrEmptyFile
	}

	return &importer.FileMetadata{
		FileName:   filepath.Base(filePath),
		FileSize:   fileInfo.Size(),
		FileType:   "xlsx",
		Worksheets: sheets,
	}, nil
}

// Preview reads headers and up to maxRows rows using streaming row iterator.
func (p *ExcelParser) Preview(filePath string, sheetName string, maxRows int) (*importer.TablePreview, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext == ".xlsm" {
		return nil, importer.ErrMacroDetected
	}

	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel workbook: %w", err)
	}
	defer f.Close()

	if sheetName == "" {
		sheets := f.GetSheetList()
		if len(sheets) == 0 {
			return nil, importer.ErrEmptyFile
		}
		sheetName = sheets[0]
	}

	rows, err := f.Rows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read worksheet %s: %w", sheetName, err)
	}
	defer rows.Close()

	var headers []string
	var previewRows [][]string
	totalCount := 0

	for rows.Next() {
		columns, err := rows.Columns()
		if err != nil {
			return nil, err
		}

		if isAllEmpty(columns) {
			continue
		}

		if headers == nil {
			headers = trimRecord(columns)
			continue
		}

		totalCount++
		if len(previewRows) < maxRows {
			previewRows = append(previewRows, trimRecord(columns))
		}
	}

	if headers == nil {
		return nil, importer.ErrEmptyFile
	}

	return &importer.TablePreview{
		Headers:   headers,
		Rows:      previewRows,
		TotalRows: totalCount,
	}, nil
}

// ReadAll reads all rows from the selected worksheet.
func (p *ExcelParser) ReadAll(filePath string, sheetName string) ([]string, [][]string, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext == ".xlsm" {
		return nil, nil, importer.ErrMacroDetected
	}

	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open Excel workbook: %w", err)
	}
	defer f.Close()

	if sheetName == "" {
		sheets := f.GetSheetList()
		if len(sheets) == 0 {
			return nil, nil, importer.ErrEmptyFile
		}
		sheetName = sheets[0]
	}

	rows, err := f.Rows(sheetName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read worksheet %s: %w", sheetName, err)
	}
	defer rows.Close()

	var headers []string
	var allRows [][]string

	for rows.Next() {
		columns, err := rows.Columns()
		if err != nil {
			return nil, nil, err
		}

		if isAllEmpty(columns) {
			continue
		}

		if headers == nil {
			headers = trimRecord(columns)
			continue
		}

		allRows = append(allRows, trimRecord(columns))
	}

	if headers == nil {
		return nil, nil, importer.ErrEmptyFile
	}

	return headers, allRows, nil
}

func isAllEmpty(record []string) bool {
	for _, f := range record {
		if strings.TrimSpace(f) != "" {
			return false
		}
	}
	return true
}

func trimRecord(record []string) []string {
	res := make([]string, len(record))
	for i, s := range record {
		res[i] = strings.TrimSpace(s)
	}
	return res
}
