package csv

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"certificate-generator/internal/importer"
)

// CSVParser parses CSV files with automatic delimiter detection and empty row filtering.
type CSVParser struct{}

// NewCSVParser constructs a new CSVParser.
func NewCSVParser() *CSVParser {
	return &CSVParser{}
}

// DetectDelimiter inspects the first non-empty lines to determine delimiter (comma, semicolon, or tab).
func DetectDelimiter(r io.Reader) (rune, error) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// Count occurrences
		commas := strings.Count(line, ",")
		semicolons := strings.Count(line, ";")
		tabs := strings.Count(line, "\t")

		if semicolons > commas && semicolons > tabs {
			return ';', nil
		}
		if tabs > commas && tabs > semicolons {
			return '\t', nil
		}
		return ',', nil
	}
	return ',', scanner.Err()
}

// Preview reads the headers and up to maxRows rows.
func (p *CSVParser) Preview(filePath string, maxRows int) (*importer.TablePreview, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}
	if fileInfo.Size() > importer.MaxImportFileSize {
		return nil, importer.ErrFileTooLarge
	}
	if fileInfo.Size() == 0 {
		return nil, importer.ErrEmptyFile
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	delimiter, _ := DetectDelimiter(bytes.NewReader(data))

	reader := csv.NewReader(bytes.NewReader(data))
	reader.Comma = delimiter
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	var headers []string
	var previewRows [][]string
	totalCount := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading CSV line: %w", err)
		}

		// Check if record is completely empty
		if isAllEmpty(record) {
			continue
		}

		if headers == nil {
			headers = trimRecord(record)
			continue
		}

		totalCount++
		if len(previewRows) < maxRows {
			previewRows = append(previewRows, trimRecord(record))
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

// ReadAll reads the full table into memory.
func (p *CSVParser) ReadAll(filePath string) ([]string, [][]string, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, nil, err
	}
	if fileInfo.Size() > importer.MaxImportFileSize {
		return nil, nil, importer.ErrFileTooLarge
	}
	if fileInfo.Size() == 0 {
		return nil, nil, importer.ErrEmptyFile
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, err
	}

	delimiter, _ := DetectDelimiter(bytes.NewReader(data))

	reader := csv.NewReader(bytes.NewReader(data))
	reader.Comma = delimiter
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	var headers []string
	var rows [][]string

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, fmt.Errorf("error reading CSV: %w", err)
		}

		if isAllEmpty(record) {
			continue
		}

		if headers == nil {
			headers = trimRecord(record)
			continue
		}

		rows = append(rows, trimRecord(record))
	}

	if headers == nil {
		return nil, nil, importer.ErrEmptyFile
	}

	return headers, rows, nil
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
