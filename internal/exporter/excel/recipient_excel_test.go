package excel_test

import (
	"path/filepath"
	"testing"

	"certificate-generator/internal/domain"
	excelexport "certificate-generator/internal/exporter/excel"

	"github.com/xuri/excelize/v2"
)

func TestExportRecipientsToExcelFile(t *testing.T) {
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "Data-Penerima-Test.xlsx")

	proj := &domain.Project{
		ID:            "proj-1",
		Name:          "Wisuda Akbar 2026",
		EventLocation: "Hotel Mulia",
		EventAddress:  "Jl. Senayan No. 1, Jakarta",
		EventCity:     "Kecamatan Tanah Abang • Jakarta Pusat",
		EventDate:     "2026-11-03",
	}

	recipients := []domain.Recipient{
		{
			ID:        "rec-1",
			ProjectID: "proj-1",
			FullName:  "Budi Santoso",
			IssueDate: "2026-11-03",
		},
		{
			ID:        "rec-2",
			ProjectID: "proj-1",
			FullName:  "Siti Aminah",
			IssueDate: "", // should fallback to proj.EventDate
		},
	}

	err := excelexport.ExportRecipientsToExcelFile(proj, recipients, outputPath)
	if err != nil {
		t.Fatalf("ExportRecipientsToExcelFile failed: %v", err)
	}

	// Verify generated excel file
	f, err := excelize.OpenFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to open exported excel file: %v", err)
	}
	defer f.Close()

	sheetName := "Data Penerima"
	rows, err := f.GetRows(sheetName)
	if err != nil {
		t.Fatalf("Failed to get rows from sheet %s: %v", sheetName, err)
	}

	if len(rows) != 3 { // 1 header + 2 data rows
		t.Fatalf("Expected 3 rows, got %d", len(rows))
	}

	// Verify headers: No, Nama Lengkap, Tanggal Acara, Tempat, Alamat, Kecamatan - Kota
	expectedHeaders := []string{"No", "Nama Lengkap", "Tanggal Acara", "Tempat", "Alamat", "Kecamatan - Kota"}
	for i, h := range expectedHeaders {
		if rows[0][i] != h {
			t.Errorf("Header %d: expected %s, got %s", i, h, rows[0][i])
		}
	}

	// Row 1
	if rows[1][0] != "1" || rows[1][1] != "Budi Santoso" || rows[1][2] != "2026-11-03" || rows[1][3] != "Hotel Mulia" || rows[1][4] != "Jl. Senayan No. 1, Jakarta" || rows[1][5] != "Kecamatan Tanah Abang • Jakarta Pusat" {
		t.Errorf("Row 1 content mismatch: %v", rows[1])
	}

	// Row 2
	if rows[2][0] != "2" || rows[2][1] != "Siti Aminah" || rows[2][2] != "2026-11-03" || rows[2][3] != "Hotel Mulia" || rows[2][4] != "Jl. Senayan No. 1, Jakarta" || rows[2][5] != "Kecamatan Tanah Abang • Jakarta Pusat" {
		t.Errorf("Row 2 content mismatch: %v", rows[2])
	}
}
