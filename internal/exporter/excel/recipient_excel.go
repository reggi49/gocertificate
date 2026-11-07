package excel

import (
	"fmt"
	"os"
	"path/filepath"

	"certificate-generator/internal/domain"

	"github.com/xuri/excelize/v2"
)

// ExportRecipientsToExcelFile writes recipients for a project into a beautifully styled Excel (.xlsx) file.
// Format: 5 columns (No, Nama Lengkap, Tanggal Acara, Tempat, Alamat)
func ExportRecipientsToExcelFile(proj *domain.Project, recipients []domain.Recipient, outputPath string) error {
	f := excelize.NewFile()
	defer func() {
		_ = f.Close()
	}()

	sheetName := "Data Penerima"
	f.SetSheetName("Sheet1", sheetName)
	if idx, err := f.GetSheetIndex(sheetName); err == nil {
		f.SetActiveSheet(idx)
	}

	trueVal := true
	viewNormal := "normal"
	cellA1 := "A1"
	_ = f.SetSheetView(sheetName, 0, &excelize.ViewOptions{
		DefaultGridColor:  &trueVal,
		ShowGridLines:     &trueVal,
		ShowRowColHeaders: &trueVal,
		ShowZeros:         &trueVal,
		TopLeftCell:       &cellA1,
		View:              &viewNormal,
	})

	headers := []string{
		"No",
		"Nama Lengkap",
		"Tanggal Acara",
		"Tempat",
		"Alamat",
		"Kecamatan - Kota",
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Color: "FFFFFF",
			Size:  11,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"1E3A8A"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "D1D5DB", Style: 1},
			{Type: "top", Color: "D1D5DB", Style: 1},
			{Type: "bottom", Color: "D1D5DB", Style: 1},
			{Type: "right", Color: "D1D5DB", Style: 1},
		},
	})

	dataStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "E5E7EB", Style: 1},
			{Type: "top", Color: "E5E7EB", Style: 1},
			{Type: "bottom", Color: "E5E7EB", Style: 1},
			{Type: "right", Color: "E5E7EB", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Vertical: "center",
		},
	})

	centerStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "E5E7EB", Style: 1},
			{Type: "top", Color: "E5E7EB", Style: 1},
			{Type: "bottom", Color: "E5E7EB", Style: 1},
			{Type: "right", Color: "E5E7EB", Style: 1},
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

	var projDate, projLocation, projAddress, projCity string
	if proj != nil {
		projDate = proj.EventDate
		projLocation = proj.EventLocation
		projAddress = proj.EventAddress
		projCity = proj.EventCity
	}

	for i, r := range recipients {
		row := i + 2
		_ = f.SetRowHeight(sheetName, row, 22)

		eventDate := r.IssueDate
		if eventDate == "" {
			eventDate = projDate
		}

		_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), i+1)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), r.FullName)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), eventDate)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), projLocation)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), projAddress)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), projCity)

		_ = f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), centerStyle)
		_ = f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), dataStyle)
		_ = f.SetCellStyle(sheetName, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), centerStyle)
		_ = f.SetCellStyle(sheetName, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), dataStyle)
		_ = f.SetCellStyle(sheetName, fmt.Sprintf("E%d", row), fmt.Sprintf("E%d", row), dataStyle)
		_ = f.SetCellStyle(sheetName, fmt.Sprintf("F%d", row), fmt.Sprintf("F%d", row), dataStyle)
	}

	totalRows := len(recipients) + 1
	if totalRows < 1 {
		totalRows = 1
	}
	_ = f.SetSheetDimension(sheetName, fmt.Sprintf("A1:F%d", totalRows))

	_ = f.SetColWidth(sheetName, "A", "A", 6)
	_ = f.SetColWidth(sheetName, "B", "B", 30)
	_ = f.SetColWidth(sheetName, "C", "C", 18)
	_ = f.SetColWidth(sheetName, "D", "D", 28)
	_ = f.SetColWidth(sheetName, "E", "E", 40)
	_ = f.SetColWidth(sheetName, "F", "F", 40)

	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("gagal membuat folder output: %w", err)
	}

	if err := f.SaveAs(outputPath); err != nil {
		return fmt.Errorf("gagal menyimpan file excel ke %s: %w", outputPath, err)
	}

	return nil
}
