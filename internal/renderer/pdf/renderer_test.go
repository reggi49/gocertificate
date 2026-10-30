package pdf_test

import (
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"certificate-generator/internal/domain"
	imagegen "certificate-generator/internal/renderer/image"
	"certificate-generator/internal/renderer/pdf"
	"certificate-generator/internal/storage/filesystem"
)

func setupTestAssets(t *testing.T, assetDir string) {
	t.Helper()
	_ = os.MkdirAll(assetDir, 0755)

	bgPath := filepath.Join(assetDir, "bg.jpg")
	_ = imagegen.GenerateTestJPEG(bgPath, 800, 600, color.RGBA{R: 245, G: 247, B: 250, A: 255})

	logoPath := filepath.Join(assetDir, "logo.png")
	_ = imagegen.GenerateTestPNGWithAlpha(logoPath, 200, 200, color.RGBA{R: 20, G: 80, B: 180, A: 220})

	sigPath := filepath.Join(assetDir, "signature.png")
	_ = imagegen.GenerateTestPNGWithAlpha(sigPath, 300, 100, color.RGBA{R: 10, G: 10, B: 30, A: 255})
}

func TestPDFRenderingA4Landscape(t *testing.T) {
	tempDir := t.TempDir()
	fontDir := filepath.Join("..", "..", "..", "assets", "fonts")
	assetDir := filepath.Join(tempDir, "assets")
	setupTestAssets(t, assetDir)

	outputDir := filepath.Join("..", "..", "..", "tests", "fixtures", "output")
	_ = os.MkdirAll(outputDir, 0755)

	doc := &domain.Document{
		SchemaVersion: 1,
		Page: domain.PageSettings{
			WidthMm:         297,
			HeightMm:        210,
			Orientation:     "landscape",
			BackgroundColor: "#FBFBFC",
		},
		Elements: []domain.Element{
			{
				ID:       "el-bg",
				Type:     domain.ElementTypeBackground,
				XMm:      0,
				YMm:      0,
				WidthMm:  297,
				HeightMm: 210,
				ZIndex:   0,
				Visible:  true,
				Properties: domain.ElementProperties{
					RelativePath: "bg.jpg",
				},
			},
			{
				ID:       "el-rect-border",
				Type:     domain.ElementTypeRectangle,
				XMm:      10,
				YMm:      10,
				WidthMm:  277,
				HeightMm: 190,
				ZIndex:   1,
				Visible:  true,
				Properties: domain.ElementProperties{
					StrokeColor: "#1A365D",
					StrokeWidth: 1.5,
				},
			},
			{
				ID:       "el-gold-circle",
				Type:     domain.ElementTypeCircle,
				XMm:      250,
				YMm:      20,
				WidthMm:  25,
				HeightMm: 25,
				ZIndex:   2,
				Visible:  true,
				Properties: domain.ElementProperties{
					FillColor:   "#FEFCBF",
					StrokeColor: "#D69E2E",
					StrokeWidth: 1.0,
				},
			},
			{
				ID:       "el-title",
				Type:     domain.ElementTypeText,
				XMm:      30,
				YMm:      35,
				WidthMm:  237,
				HeightMm: 20,
				ZIndex:   3,
				Visible:  true,
				Properties: domain.ElementProperties{
					Text:       "SERTIFIKAT PENGHARGAAN",
					FontFamily: "Ubuntu",
					FontSizePt: 28,
					Color:      "#1A365D",
					Alignment:  "center",
				},
			},
			{
				ID:       "el-recipient",
				Type:     domain.ElementTypeText,
				XMm:      25,
				YMm:      80,
				WidthMm:  247,
				HeightMm: 25,
				ZIndex:   4,
				Visible:  true,
				Properties: domain.ElementProperties{
					Text:              "{{recipient_name}}",
					FontFamily:        "LiberationSerif",
					FontSizePt:        30,
					Color:             "#2D3748",
					Alignment:         "center",
					VerticalAlignment: "middle",
					FitMode:           "shrink",
				},
			},
			{
				ID:       "el-qr",
				Type:     domain.ElementTypeQRCode,
				XMm:      30,
				YMm:      150,
				WidthMm:  30,
				HeightMm: 30,
				ZIndex:   5,
				Visible:  true,
				Properties: domain.ElementProperties{
					QRContent:     "https://verify.cert.local/verify?id={{certificate_id}}",
					RecoveryLevel: "M",
				},
			},
			{
				ID:       "el-sig",
				Type:     domain.ElementTypeSignature,
				XMm:      190,
				YMm:      150,
				WidthMm:  50,
				HeightMm: 25,
				ZIndex:   6,
				Visible:  true,
				Properties: domain.ElementProperties{
					RelativePath: "signature.png",
				},
			},
		},
	}

	variables := map[string]string{
		"recipient_name": "Prof. Dr. Ir. Budi Santoso-Wicaksono, M.Sc., Ph.D.",
		"certificate_id": "CERT-2026-A4L-001",
	}

	renderer := pdf.NewGoPDFRenderer(fontDir)

	// Render byte buffer
	pdfBytes, err := renderer.Render(doc, variables, assetDir)
	if err != nil {
		t.Fatalf("renderer.Render failed: %v", err)
	}

	// Validate structural PDF (297 x 210 mm = 841.89 x 595.28 pt)
	expectedWPt := domain.MmToPt(297)
	expectedHPt := domain.MmToPt(210)
	if err := pdf.ValidatePDF(pdfBytes, expectedWPt, expectedHPt, 0.5); err != nil {
		t.Fatalf("ValidatePDF failed: %v", err)
	}

	// Save artifact
	outputPath := filepath.Join(outputDir, "a4_landscape_certificate.pdf")
	if err := renderer.RenderToFile(doc, variables, assetDir, outputPath); err != nil {
		t.Fatalf("RenderToFile failed: %v", err)
	}
}

func TestPDFRenderingA4Portrait(t *testing.T) {
	fontDir := filepath.Join("..", "..", "..", "assets", "fonts")
	outputDir := filepath.Join("..", "..", "..", "tests", "fixtures", "output")
	_ = os.MkdirAll(outputDir, 0755)

	doc := &domain.Document{
		SchemaVersion: 1,
		Page: domain.PageSettings{
			WidthMm:         210,
			HeightMm:        297,
			Orientation:     "portrait",
			BackgroundColor: "#FFFFFF",
		},
		Elements: []domain.Element{
			{
				ID:       "el-title",
				Type:     domain.ElementTypeText,
				XMm:      20,
				YMm:      40,
				WidthMm:  170,
				HeightMm: 20,
				ZIndex:   1,
				Visible:  true,
				Properties: domain.ElementProperties{
					Text:       "PIAGAM KELULUSAN",
					FontFamily: "Ubuntu",
					FontSizePt: 24,
					Alignment:  "center",
				},
			},
		},
	}

	renderer := pdf.NewGoPDFRenderer(fontDir)
	pdfBytes, err := renderer.Render(doc, nil, "")
	if err != nil {
		t.Fatalf("Render A4 portrait failed: %v", err)
	}

	expectedWPt := domain.MmToPt(210)
	expectedHPt := domain.MmToPt(297)
	if err := pdf.ValidatePDF(pdfBytes, expectedWPt, expectedHPt, 0.5); err != nil {
		t.Fatalf("ValidatePDF A4 portrait failed: %v", err)
	}

	outputPath := filepath.Join(outputDir, "a4_portrait_certificate.pdf")
	_ = renderer.RenderToFile(doc, nil, "", outputPath)
}

func TestPDFRenderingLetterLandscapeAndPortrait(t *testing.T) {
	fontDir := filepath.Join("..", "..", "..", "assets", "fonts")
	outputDir := filepath.Join("..", "..", "..", "tests", "fixtures", "output")
	_ = os.MkdirAll(outputDir, 0755)
	renderer := pdf.NewGoPDFRenderer(fontDir)

	// 1. Letter Landscape: 279.4 x 215.9 mm = 792.00 x 612.00 pt
	docLandscape := &domain.Document{
		SchemaVersion: 1,
		Page: domain.PageSettings{
			WidthMm:         279.4,
			HeightMm:        215.9,
			Orientation:     "landscape",
			BackgroundColor: "#FFFFFF",
		},
		Elements: []domain.Element{
			{
				ID:       "el-t",
				Type:     domain.ElementTypeText,
				XMm:      20,
				YMm:      20,
				WidthMm:  239.4,
				HeightMm: 15,
				Visible:  true,
				Properties: domain.ElementProperties{
					Text:       "Letter Landscape Certificate",
					FontFamily: "LiberationSerif",
					FontSizePt: 20,
					Alignment:  "center",
				},
			},
		},
	}

	pdfBytesL, err := renderer.Render(docLandscape, nil, "")
	if err != nil {
		t.Fatalf("Render Letter Landscape failed: %v", err)
	}
	if err := pdf.ValidatePDF(pdfBytesL, domain.MmToPt(279.4), domain.MmToPt(215.9), 0.5); err != nil {
		t.Fatalf("ValidatePDF Letter Landscape failed: %v", err)
	}
	_ = renderer.RenderToFile(docLandscape, nil, "", filepath.Join(outputDir, "letter_landscape_certificate.pdf"))

	// 2. Letter Portrait: 215.9 x 279.4 mm = 612.00 x 792.00 pt
	docPortrait := &domain.Document{
		SchemaVersion: 1,
		Page: domain.PageSettings{
			WidthMm:         215.9,
			HeightMm:        279.4,
			Orientation:     "portrait",
			BackgroundColor: "#FFFFFF",
		},
		Elements: []domain.Element{
			{
				ID:       "el-t",
				Type:     domain.ElementTypeText,
				XMm:      20,
				YMm:      20,
				WidthMm:  175.9,
				HeightMm: 15,
				Visible:  true,
				Properties: domain.ElementProperties{
					Text:       "Letter Portrait Certificate",
					FontFamily: "LiberationSerif",
					FontSizePt: 20,
					Alignment:  "center",
				},
			},
		},
	}

	pdfBytesP, err := renderer.Render(docPortrait, nil, "")
	if err != nil {
		t.Fatalf("Render Letter Portrait failed: %v", err)
	}
	if err := pdf.ValidatePDF(pdfBytesP, domain.MmToPt(215.9), domain.MmToPt(279.4), 0.5); err != nil {
		t.Fatalf("ValidatePDF Letter Portrait failed: %v", err)
	}
	_ = renderer.RenderToFile(docPortrait, nil, "", filepath.Join(outputDir, "letter_portrait_certificate.pdf"))
}

func TestPDFRenderingCustomSizeAndWrapping(t *testing.T) {
	fontDir := filepath.Join("..", "..", "..", "assets", "fonts")
	outputDir := filepath.Join("..", "..", "..", "tests", "fixtures", "output")
	_ = os.MkdirAll(outputDir, 0755)
	renderer := pdf.NewGoPDFRenderer(fontDir)

	// Custom dimensions: 200 x 140 mm
	docCustom := &domain.Document{
		SchemaVersion: 1,
		Page: domain.PageSettings{
			WidthMm:         200,
			HeightMm:        140,
			Orientation:     "landscape",
			BackgroundColor: "#F7FAFC",
		},
		Elements: []domain.Element{
			// Multi-line wrapped text
			{
				ID:       "el-wrap",
				Type:     domain.ElementTypeText,
				XMm:      20,
				YMm:      40,
				WidthMm:  160,
				HeightMm: 60,
				Visible:  true,
				Properties: domain.ElementProperties{
					Text:              "Dengan ini menerangkan bahwa peserta telah menyelesaikan seluruh kurikulum pelatihan kecerdasan buatan dan lulus evaluasi akhir dengan predikat yang sangat memuaskan.",
					FontFamily:        "LiberationSerif",
					FontSizePt:        14,
					LineHeight:        1.4,
					Alignment:         "center",
					VerticalAlignment: "middle",
					FitMode:           "wrap",
				},
			},
		},
	}

	pdfBytes, err := renderer.Render(docCustom, nil, "")
	if err != nil {
		t.Fatalf("Render custom size failed: %v", err)
	}
	if err := pdf.ValidatePDF(pdfBytes, domain.MmToPt(200), domain.MmToPt(140), 0.5); err != nil {
		t.Fatalf("ValidatePDF custom size failed: %v", err)
	}

	_ = renderer.RenderToFile(docCustom, nil, "", filepath.Join(outputDir, "custom_size_certificate.pdf"))
	_ = renderer.RenderToFile(docCustom, nil, "", filepath.Join(outputDir, "wrapped_multiline_certificate.pdf"))
}

func TestPDFVariableValidationAndExtraction(t *testing.T) {
	doc := &domain.Document{
		SchemaVersion: 1,
		Page: domain.PageSettings{
			WidthMm:  297,
			HeightMm: 210,
		},
		Elements: []domain.Element{
			{
				ID:   "t1",
				Type: domain.ElementTypeText,
				Properties: domain.ElementProperties{
					Text: "Diberikan kepada {{recipient_name}} atas acara {{event_name}}",
				},
			},
			{
				ID:   "q1",
				Type: domain.ElementTypeQRCode,
				Properties: domain.ElementProperties{
					QRContent: "https://cert.id/{{certificate_id}}",
				},
			},
		},
	}

	placeholders := pdf.ExtractPlaceholders(doc)
	if len(placeholders) != 3 {
		t.Errorf("Expected 3 placeholders, got %d: %v", len(placeholders), placeholders)
	}

	// Missing variable detection
	missing := pdf.ValidateRequiredVariables(doc, map[string]string{
		"recipient_name": "Budi",
		// event_name and certificate_id are missing!
	})
	if len(missing) != 2 {
		t.Errorf("Expected 2 missing variables, got %d: %v", len(missing), missing)
	}

	// Case-insensitive replacement test
	textWithUpper := "BERTEMPAT DI {{EVENT_LOCATION}} PADA {{Event-Address}}"
	replaced := pdf.ReplacePlaceholders(textWithUpper, map[string]string{
		"event_location": "Janaiz",
		"event_address":  "Jl Kelapa Gading",
	})
	expectedReplaced := "BERTEMPAT DI Janaiz PADA Jl Kelapa Gading"
	if replaced != expectedReplaced {
		t.Errorf("Expected '%s', got '%s'", expectedReplaced, replaced)
	}
}

func TestSafeAtomicFileOutput(t *testing.T) {
	tempDir := t.TempDir()
	fontDir := filepath.Join("..", "..", "..", "assets", "fonts")
	renderer := pdf.NewGoPDFRenderer(fontDir)

	doc := &domain.Document{
		SchemaVersion: 1,
		Page: domain.PageSettings{
			WidthMm:  200,
			HeightMm: 150,
		},
		Elements: []domain.Element{},
	}

	// Output with illegal filename characters that must be sanitized
	rawFilename := `Sertifikat: "Budi" <Pemenang> ? * [2026].pdf`
	outputPath := filepath.Join(tempDir, rawFilename)
	err := renderer.RenderToFile(doc, nil, "", outputPath)
	if err != nil {
		t.Fatalf("RenderToFile with unsanitized path failed: %v", err)
	}

	// Verify file was written with clean sanitized name
	cleanBase := filesystem.SanitizeFilename(rawFilename)
	expectedCleanPath := filepath.Join(tempDir, cleanBase)
	info, err := os.Stat(expectedCleanPath)
	if err != nil || info.Size() == 0 {
		t.Errorf("Expected sanitized file at %s to exist with size > 0", expectedCleanPath)
	}
}

func TestDefaultJanaizTemplate(t *testing.T) {
	fontDir := filepath.Join("..", "..", "..", "assets", "fonts")
	renderer := pdf.NewGoPDFRenderer(fontDir)

	doc := &domain.Document{
		SchemaVersion: 1,
		Page: domain.PageSettings{
			WidthMm:         210,
			HeightMm:        297,
			Orientation:     "portrait",
			BackgroundColor: "#000000",
		},
		Elements: []domain.Element{
			{
				ID:       "bg-janaiz",
				Type:     domain.ElementTypeBackground,
				XMm:      0,
				YMm:      0,
				WidthMm:  210,
				HeightMm: 297,
				Visible:  true,
				Locked:   true,
				ZIndex:   0,
				Opacity:  1.0,
				Properties: domain.ElementProperties{
					RelativePath: "backgrounds/default_certificate_bg.jpg",
				},
			},
			{
				ID:       "el-name",
				Type:     domain.ElementTypeText,
				XMm:      15,
				YMm:      110,
				WidthMm:  180,
				HeightMm: 22,
				Visible:  true,
				Locked:   false,
				ZIndex:   1,
				Opacity:  1.0,
				Properties: domain.ElementProperties{
					Text:              "{{recipient_name}}",
					FontFamily:        "Sloop Scriptone",
					FontSizePt:        50,
					FontWeight:        400,
					Color:             "#CBBDA3",
					Alignment:         "center",
					VerticalAlignment: "middle",
					FitMode:           "shrink",
				},
			},
			{
				ID:       "el-event-details",
				Type:     domain.ElementTypeText,
				XMm:      15,
				YMm:      205,
				WidthMm:  180,
				HeightMm: 28,
				Visible:  true,
				Locked:   false,
				ZIndex:   2,
				Opacity:  1.0,
				Properties: domain.ElementProperties{
					Text:              "PADA TANGGAL {{issue_date}}\nBERTEMPAT DI {{event_location}}\n{{event_address}}",
					FontFamily:        "Trajan Pro Bold",
					FontSizePt:        12,
					FontWeight:        700,
					Color:             "#CBBDA3",
					Alignment:         "center",
					VerticalAlignment: "middle",
					FitMode:           "wrap",
					LineHeight:        1.4,
				},
			},
		},
	}

	variables := map[string]string{
		"recipient_name": "Maharani Syahputra",
		"issue_date":     "23 OKTOBER 2026",
		"event_location": "YAYASAN JANAIZ",
		"event_address":  "SEDAYU CITY BLOK C NO.03 • JAKARTA UTARA",
	}

	assetDir := filepath.Join("..", "..", "..", "assets")
	pdfBytes, err := renderer.Render(doc, variables, assetDir)
	if err != nil {
		t.Fatalf("Failed to render Janaiz template: %v", err)
	}

	if len(pdfBytes) < 5000 {
		t.Fatalf("Generated PDF is too small: %d bytes", len(pdfBytes))
	}

	outputDir := filepath.Join("..", "..", "..", "tests", "fixtures", "output")
	_ = os.MkdirAll(outputDir, 0755)
	_ = os.WriteFile(filepath.Join(outputDir, "default_janaiz_test.pdf"), pdfBytes, 0644)
}
