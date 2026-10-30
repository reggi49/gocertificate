package importer_test

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"certificate-generator/internal/importer"
	csvparser "certificate-generator/internal/importer/csv"
	excelparser "certificate-generator/internal/importer/excel"

	"github.com/xuri/excelize/v2"
)

func ensure1000CSV(path string) error {
	if info, err := os.Stat(path); err == nil && info.Size() > 10000 {
		return nil
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	// Header
	_ = w.Write([]string{"Nama Lengkap", "Email", "Organisasi", "Nama Acara", "Tanggal Terbit", "Nomor Sertifikat", "Nilai Kelulusan"})

	for i := 1; i <= 1000; i++ {
		row := []string{
			fmt.Sprintf("Peserta Ke-%04d, S.Kom.", i),
			fmt.Sprintf("peserta%04d@instansi.id", i),
			"Universitas Indonesia",
			"Pelatihan Nasional AI 2026",
			"30 Oktober 2026",
			fmt.Sprintf("CERT/2026/%05d", i),
			fmt.Sprintf("%d", 80+(i%20)),
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}
	return nil
}

func ensure1000XLSX(path string) error {
	if info, err := os.Stat(path); err == nil && info.Size() > 10000 {
		return nil
	}
	f := excelize.NewFile()
	defer f.Close()

	sheet := "Daftar Peserta"
	index, _ := f.NewSheet(sheet)
	f.SetActiveSheet(index)
	_ = f.DeleteSheet("Sheet1")

	// Header
	headers := []string{"Nama Lengkap", "Email", "Organisasi", "Nama Acara", "Tanggal Terbit", "Nomor Sertifikat", "Predikat"}
	for colIdx, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}

	for i := 1; i <= 1000; i++ {
		rowIdx := i + 1
		vals := []any{
			fmt.Sprintf("Dr. Ir. Peserta %04d, M.Sc.", i),
			fmt.Sprintf("peserta%04d@kampus.id", i),
			"Institut Teknologi Bandung",
			"Konferensi Nasional 2026",
			"2026-10-30",
			fmt.Sprintf("CERT-ITB-%05d", i),
			"Sangat Memuaskan",
		}
		for colIdx, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx)
			_ = f.SetCellValue(sheet, cell, v)
		}
	}

	return f.SaveAs(path)
}

func ensureMultiSheetXLSX(path string) error {
	if info, err := os.Stat(path); err == nil && info.Size() > 1000 {
		return nil
	}
	f := excelize.NewFile()
	defer f.Close()

	sheet1 := "Peserta Gelombang 1"
	sheet2 := "Peserta Gelombang 2"

	f.SetSheetName("Sheet1", sheet1)
	_, _ = f.NewSheet(sheet2)

	_ = f.SetCellValue(sheet1, "A1", "Nama Lengkap")
	_ = f.SetCellValue(sheet1, "B1", "Email")
	_ = f.SetCellValue(sheet1, "A2", "Budi Wicaksono")
	_ = f.SetCellValue(sheet1, "B2", "budi@mail.com")

	_ = f.SetCellValue(sheet2, "A1", "Nama Lengkap")
	_ = f.SetCellValue(sheet2, "B1", "Email")
	_ = f.SetCellValue(sheet2, "A2", "Siti Aminah")
	_ = f.SetCellValue(sheet2, "B2", "siti@mail.com")

	return f.SaveAs(path)
}

func TestParsersWithFixtures(t *testing.T) {
	fixtureDir := filepath.Join("..", "..", "tests", "fixtures")
	_ = os.MkdirAll(fixtureDir, 0755)

	csvPath := filepath.Join(fixtureDir, "recipients_1000.csv")
	xlsxPath := filepath.Join(fixtureDir, "recipients_1000.xlsx")
	multiSheetPath := filepath.Join(fixtureDir, "recipients_multisheet.xlsx")

	if err := ensure1000CSV(csvPath); err != nil {
		t.Fatalf("Failed to generate 1000 CSV: %v", err)
	}
	if err := ensure1000XLSX(xlsxPath); err != nil {
		t.Fatalf("Failed to generate 1000 XLSX: %v", err)
	}
	if err := ensureMultiSheetXLSX(multiSheetPath); err != nil {
		t.Fatalf("Failed to generate multi-sheet XLSX: %v", err)
	}

	// 1. Test CSV Parser
	csvP := csvparser.NewCSVParser()
	preview, err := csvP.Preview(csvPath, 5)
	if err != nil {
		t.Fatalf("CSV Preview failed: %v", err)
	}
	if len(preview.Headers) != 7 {
		t.Errorf("Expected 7 headers, got %d", len(preview.Headers))
	}
	if preview.TotalRows != 1000 {
		t.Errorf("Expected 1000 total rows, got %d", preview.TotalRows)
	}
	if len(preview.Rows) != 5 {
		t.Errorf("Expected 5 preview rows, got %d", len(preview.Rows))
	}

	// Read all CSV
	headers, rows, err := csvP.ReadAll(csvPath)
	if err != nil {
		t.Fatalf("CSV ReadAll failed: %v", err)
	}
	if len(headers) != 7 || len(rows) != 1000 {
		t.Errorf("CSV ReadAll mismatch: %d headers, %d rows", len(headers), len(rows))
	}

	// 2. Test Excel Parser
	xlsxP := excelparser.NewExcelParser()
	meta, err := xlsxP.InspectWorkbook(multiSheetPath)
	if err != nil {
		t.Fatalf("InspectWorkbook failed: %v", err)
	}
	if len(meta.Worksheets) != 2 {
		t.Errorf("Expected 2 worksheets in multi-sheet file, got %d", len(meta.Worksheets))
	}

	// Test 1000 rows XLSX preview & read
	xlsxPreview, err := xlsxP.Preview(xlsxPath, "Daftar Peserta", 5)
	if err != nil {
		t.Fatalf("XLSX Preview failed: %v", err)
	}
	if xlsxPreview.TotalRows != 1000 {
		t.Errorf("Expected 1000 total rows in XLSX, got %d", xlsxPreview.TotalRows)
	}

	// Read all XLSX
	xHeaders, xRows, err := xlsxP.ReadAll(xlsxPath, "Daftar Peserta")
	if err != nil {
		t.Fatalf("XLSX ReadAll failed: %v", err)
	}
	if len(xHeaders) != 7 || len(xRows) != 1000 {
		t.Errorf("XLSX ReadAll mismatch: %d headers, %d rows", len(xHeaders), len(xRows))
	}

	// 3. Test Security check: Macro rejection
	xlsmPath := filepath.Join(t.TempDir(), "macro_file.xlsm")
	_ = os.WriteFile(xlsmPath, []byte("fake macro binary"), 0644)
	_, err = xlsxP.InspectWorkbook(xlsmPath)
	if err != importer.ErrMacroDetected {
		t.Errorf("Expected ErrMacroDetected for .xlsm file, got %v", err)
	}
}
