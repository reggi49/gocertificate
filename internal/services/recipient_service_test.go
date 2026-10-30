package services_test

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"certificate-generator/internal/domain"
	"certificate-generator/internal/repository/sqlite"
	"certificate-generator/internal/services"

	"github.com/xuri/excelize/v2"
)

func prepare1000Fixtures(t *testing.T, dir string) (string, string) {
	t.Helper()
	csvPath := filepath.Join(dir, "recipients_1000.csv")
	xlsxPath := filepath.Join(dir, "recipients_1000.xlsx")

	// 1. CSV
	f, err := os.Create(csvPath)
	if err != nil {
		t.Fatalf("Failed to create CSV: %v", err)
	}
	w := csv.NewWriter(f)
	_ = w.Write([]string{"Nama Lengkap", "Email", "Organisasi", "Nama Acara", "Tanggal Terbit", "Nomor Sertifikat", "Nilai Kelulusan"})
	for i := 1; i <= 1000; i++ {
		_ = w.Write([]string{
			fmt.Sprintf("Peserta Ke-%04d, S.Kom.", i),
			fmt.Sprintf("peserta%04d@instansi.id", i),
			"Universitas Indonesia",
			"Pelatihan Nasional AI 2026",
			"30 Oktober 2026",
			fmt.Sprintf("CERT/2026/%05d", i),
			fmt.Sprintf("%d", 80+(i%20)),
		})
	}
	w.Flush()
	_ = f.Close()

	// 2. XLSX
	xl := excelize.NewFile()
	defer xl.Close()
	sheet := "Daftar Peserta"
	idx, _ := xl.NewSheet(sheet)
	xl.SetActiveSheet(idx)
	_ = xl.DeleteSheet("Sheet1")

	headers := []string{"Nama Lengkap", "Email", "Organisasi", "Nama Acara", "Tanggal Terbit", "Nomor Sertifikat", "Predikat"}
	for colIdx, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
		_ = xl.SetCellValue(sheet, cell, h)
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
			_ = xl.SetCellValue(sheet, cell, v)
		}
	}
	if err := xl.SaveAs(xlsxPath); err != nil {
		t.Fatalf("Failed to save XLSX: %v", err)
	}

	return csvPath, xlsxPath
}

func TestRecipientServiceImportAndCRUD(t *testing.T) {
	projSvc, pRepo, sm, _, cleanup := setupServiceTest(t)
	defer cleanup()

	ctx := context.Background()

	// 1. Create Project
	proj, err := projSvc.CreateProject(ctx, domain.CreateProjectInput{
		Name: "Proyek Import 1000 Peserta",
	})
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	dbPath := filepath.Join(sm.DatabaseDir(), "app.db")
	db, _ := sqlite.Open(dbPath)
	defer db.Close()
	rRepo := sqlite.NewRecipientRepository(db)
	recSvc := services.NewRecipientService(rRepo, pRepo, nil)

	// Prepare isolated test fixtures
	csv1000Path, xlsx1000Path := prepare1000Fixtures(t, t.TempDir())

	// 2. Test InspectFile
	metaCSV, err := recSvc.InspectFile(csv1000Path)
	if err != nil {
		t.Fatalf("InspectFile CSV failed: %v", err)
	}
	if metaCSV.FileType != "csv" || metaCSV.FileSize == 0 {
		t.Errorf("Unexpected CSV metadata: %+v", metaCSV)
	}

	metaXLSX, err := recSvc.InspectFile(xlsx1000Path)
	if err != nil {
		t.Fatalf("InspectFile XLSX failed: %v", err)
	}
	if metaXLSX.FileType != "xlsx" || len(metaXLSX.Worksheets) != 1 {
		t.Errorf("Unexpected XLSX metadata: %+v", metaXLSX)
	}

	// 3. Test PreviewFile
	preview, err := recSvc.PreviewFile(csv1000Path, "", 5)
	if err != nil {
		t.Fatalf("PreviewFile failed: %v", err)
	}
	if len(preview.Rows) != 5 || preview.TotalRows != 1000 {
		t.Errorf("Preview result mismatch: %d rows, %d total", len(preview.Rows), preview.TotalRows)
	}

	// 4. Test Column Mapping Validation
	mapping := domain.ColumnMapping{
		FullNameColumn:      "Nama Lengkap",
		EmailColumn:         "Email",
		OrganizationColumn:  "Organisasi",
		EventNameColumn:     "Nama Acara",
		IssueDateColumn:     "Tanggal Terbit",
		CertificateIDColumn: "Nomor Sertifikat",
		CustomFieldColumns: map[string]string{
			"score": "Nilai Kelulusan",
		},
	}

	validSummary, err := recSvc.ValidateMapping(csv1000Path, "", mapping)
	if err != nil {
		t.Fatalf("ValidateMapping failed: %v", err)
	}
	if validSummary.TotalRows != 1000 || validSummary.ValidRows != 1000 || validSummary.InvalidRows != 0 {
		t.Errorf("Expected 1000 valid rows, got %+v", validSummary)
	}

	// 5. Test Large Dataset Import (1000 recipients via CSV)
	importSummary, err := recSvc.ImportRecipients(ctx, proj.ID, csv1000Path, "", mapping, domain.ConflictStrategyFail)
	if err != nil {
		t.Fatalf("ImportRecipients 1000 CSV failed: %v", err)
	}
	if importSummary.ImportedRows != 1000 {
		t.Errorf("Expected 1000 imported rows, got %d", importSummary.ImportedRows)
	}

	// Verify count in DB
	totalCount, err := recSvc.CountRecipients(ctx, proj.ID)
	if err != nil {
		t.Fatalf("CountRecipients failed: %v", err)
	}
	if totalCount != 1000 {
		t.Errorf("Expected DB count to be 1000, got %d", totalCount)
	}

	// 6. Test Pagination & Search on 1000 items
	page1, err := recSvc.ListRecipients(ctx, domain.RecipientFilter{
		ProjectID: proj.ID,
		Page:      1,
		PageSize:  25,
	})
	if err != nil {
		t.Fatalf("ListRecipients page 1 failed: %v", err)
	}
	if len(page1.Items) != 25 || page1.Total != 1000 || page1.TotalPages != 40 {
		t.Errorf("Pagination page 1 mismatch: items=%d, total=%d, totalPages=%d", len(page1.Items), page1.Total, page1.TotalPages)
	}

	// Search on 1000 items
	searchRes, err := recSvc.ListRecipients(ctx, domain.RecipientFilter{
		ProjectID: proj.ID,
		Search:    "0500",
		Page:      1,
		PageSize:  10,
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if searchRes.Total != 1 || searchRes.Items[0].FullName != "Peserta Ke-0500, S.Kom." {
		t.Errorf("Search result mismatch: %+v", searchRes)
	}

	// 7. Test Re-import with Duplicate Certificate IDs (Skip strategy)
	reimportSummary, err := recSvc.ImportRecipients(ctx, proj.ID, csv1000Path, "", mapping, domain.ConflictStrategySkip)
	if err != nil {
		t.Fatalf("Re-import with skip failed: %v", err)
	}
	if reimportSummary.SkippedRows != 1000 {
		t.Errorf("Expected all 1000 duplicates to be skipped, got %d skipped, %d imported", reimportSummary.SkippedRows, reimportSummary.ImportedRows)
	}

	// 8. Test DeleteAllRecipients
	if err := recSvc.DeleteAllRecipients(ctx, proj.ID); err != nil {
		t.Fatalf("DeleteAllRecipients failed: %v", err)
	}
	afterClearCount, _ := recSvc.CountRecipients(ctx, proj.ID)
	if afterClearCount != 0 {
		t.Errorf("Expected 0 recipients after clear, got %d", afterClearCount)
	}

	// 9. Test Import 1000 recipients via XLSX
	xlsxMapping := domain.ColumnMapping{
		FullNameColumn:      "Nama Lengkap",
		EmailColumn:         "Email",
		OrganizationColumn:  "Organisasi",
		EventNameColumn:     "Nama Acara",
		IssueDateColumn:     "Tanggal Terbit",
		CertificateIDColumn: "Nomor Sertifikat",
		CustomFieldColumns: map[string]string{
			"predikat": "Predikat",
		},
	}
	xlsxImportSummary, err := recSvc.ImportRecipients(ctx, proj.ID, xlsx1000Path, "Daftar Peserta", xlsxMapping, domain.ConflictStrategyFail)
	if err != nil {
		t.Fatalf("ImportRecipients 1000 XLSX failed: %v", err)
	}
	if xlsxImportSummary.ImportedRows != 1000 {
		t.Errorf("Expected 1000 imported rows from XLSX, got %d", xlsxImportSummary.ImportedRows)
	}
}
