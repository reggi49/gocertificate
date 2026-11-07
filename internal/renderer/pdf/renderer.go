package pdf

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"certificate-generator/internal/domain"
	fontresolver "certificate-generator/internal/renderer/font"
	"certificate-generator/internal/storage/filesystem"

	"github.com/signintech/gopdf"
	"github.com/skip2/go-qrcode"
)

// Renderer defines the contract for certificate PDF generation.
type Renderer interface {
	Render(doc *domain.Document, variables map[string]string, assetRootDir string) ([]byte, error)
	RenderToFile(doc *domain.Document, variables map[string]string, assetRootDir string, outputPath string) error
}

// GoPDFRenderer implements Renderer using github.com/signintech/gopdf.
type GoPDFRenderer struct {
	FontDir      string
	fontResolver *fontresolver.Resolver
}

// NewGoPDFRenderer constructs a renderer with the specified font directory.
func NewGoPDFRenderer(fontDir string) *GoPDFRenderer {
	return &GoPDFRenderer{
		FontDir:      fontDir,
		fontResolver: fontresolver.NewResolver(fontDir),
	}
}

// ParseHexColor parses hex color strings (e.g. #RRGGBB or #RGB) into R, G, B components.
func ParseHexColor(hex string) (uint8, uint8, uint8, error) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) == 3 {
		r, err := strconv.ParseUint(string([]byte{hex[0], hex[0]}), 16, 8)
		if err != nil {
			return 0, 0, 0, err
		}
		g, err := strconv.ParseUint(string([]byte{hex[1], hex[1]}), 16, 8)
		if err != nil {
			return 0, 0, 0, err
		}
		b, err := strconv.ParseUint(string([]byte{hex[2], hex[2]}), 16, 8)
		if err != nil {
			return 0, 0, 0, err
		}
		return uint8(r), uint8(g), uint8(b), nil
	}
	if len(hex) == 6 {
		val, err := strconv.ParseUint(hex, 16, 32)
		if err != nil {
			return 0, 0, 0, err
		}
		return uint8(val >> 16), uint8((val >> 8) & 0xFF), uint8(val & 0xFF), nil
	}
	return 0, 0, 0, fmt.Errorf("invalid hex color: %s", hex)
}

// Render compiles the document into a professional vector PDF byte buffer.
func (r *GoPDFRenderer) Render(doc *domain.Document, variables map[string]string, assetRootDir string) ([]byte, error) {
	var buf bytes.Buffer
	if err := r.renderTo(&buf, doc, variables, assetRootDir); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// renderTo builds the PDF and streams it directly into w (no intermediate full-size copy).
func (r *GoPDFRenderer) renderTo(w io.Writer, doc *domain.Document, variables map[string]string, assetRootDir string) (retErr error) {
	// Guard against panic in third-party PDF subsetter/measurer so the app NEVER crashes
	defer func() {
		if rec := recover(); rec != nil {
			retErr = fmt.Errorf("pdf rendering panic recovered: %v", rec)
		}
	}()

	if err := doc.Validate(); err != nil {
		return fmt.Errorf("invalid document: %w", err)
	}

	pdf := &gopdf.GoPdf{}

	pageWidthPt := domain.MmToPt(doc.Page.WidthMm)
	pageHeightPt := domain.MmToPt(doc.Page.HeightMm)

	pdf.Start(gopdf.Config{
		PageSize: gopdf.Rect{
			W: pageWidthPt,
			H: pageHeightPt,
		},
	})

	pdf.AddPage()

	// Fonts are registered lazily (only the families actually used by visible text elements).
	// Previously ALL embedded fonts (~1.5 MB of TTF) were parsed, subset and flate-compressed
	// for every certificate, which dominated CPU/RAM usage during batch generation.
	registeredFonts := make(map[string]bool)
	defaultFontName := "LiberationSerif"
	defaultFontReady := false

	ensureDefaultFont := func() string {
		if defaultFontReady {
			return defaultFontName
		}
		defaultFontReady = true
		candidates := []string{"LiberationSerif", "Ubuntu"}
		for _, name := range candidates {
			if registeredFonts[name] {
				defaultFontName = name
				return defaultFontName
			}
			if canon, data, ok := r.fontResolver.ResolveBytes(name); ok {
				if err := pdf.AddTTFFontData(canon, data); err == nil {
					registeredFonts[canon] = true
					defaultFontName = canon
					return defaultFontName
				}
			}
		}
		diskFonts := map[string]string{
			"LiberationSerif": filepath.Join(r.FontDir, "LiberationSerif-Regular.ttf"),
			"Ubuntu":          filepath.Join(r.FontDir, "Ubuntu-Regular.ttf"),
		}
		for _, name := range candidates {
			p := diskFonts[name]
			if _, err := os.Stat(p); err == nil {
				if err := pdf.AddTTFFont(name, p); err == nil {
					registeredFonts[name] = true
					defaultFontName = name
					return defaultFontName
				}
			}
		}
		// Last resort: any embedded font
		for family, data := range r.fontResolver.AllEmbeddedFonts() {
			if err := pdf.AddTTFFontData(family, data); err == nil {
				registeredFonts[family] = true
				defaultFontName = family
				return defaultFontName
			}
		}
		return defaultFontName
	}

	// Render background color if defined
	if doc.Page.BackgroundColor != "" && doc.Page.BackgroundColor != "#FFFFFF" {
		if cr, cg, cb, err := ParseHexColor(doc.Page.BackgroundColor); err == nil {
			pdf.SetFillColor(cr, cg, cb)
			pdf.RectFromUpperLeftWithStyle(0, 0, pageWidthPt, pageHeightPt, "F")
		}
	}

	// Sort elements by zIndex ascending to preserve correct layer order
	elements := make([]domain.Element, len(doc.Elements))
	copy(elements, doc.Elements)
	sort.Slice(elements, func(i, j int) bool {
		return elements[i].ZIndex < elements[j].ZIndex
	})

	for _, el := range elements {
		if !el.Visible {
			continue
		}

		xPt := domain.MmToPt(el.XMm)
		yPt := domain.MmToPt(el.YMm)
		wPt := domain.MmToPt(el.WidthMm)
		hPt := domain.MmToPt(el.HeightMm)

		switch el.Type {
		case domain.ElementTypeBackground, domain.ElementTypeImage, domain.ElementTypeLogo, domain.ElementTypeSignature:
			imgRelPath := el.Properties.RelativePath
			if imgRelPath != "" {
				loaded := false
				candidates := []string{
					filepath.Join(assetRootDir, imgRelPath),
					imgRelPath,
					filepath.Join("assets", imgRelPath),
					filepath.Join("assets", "backgrounds", filepath.Base(imgRelPath)),
					filepath.Join("frontend", "public", imgRelPath),
					filepath.Join("frontend", "public", "assets", filepath.Base(imgRelPath)),
				}
				for _, candidate := range candidates {
					if img, err := loadDiskImage(candidate); err == nil {
						if err := pdf.ImageByHolder(img.holder(), xPt, yPt, &gopdf.Rect{W: wPt, H: hPt}); err == nil {
							loaded = true
							break
						}
					}
				}

				// Fallback to embedded backgrounds (guarantees standalone binary works anywhere)
				if !loaded {
					baseName := filepath.Base(imgRelPath)
					if img, err := loadEmbeddedBackground(baseName); err == nil {
						if err := pdf.ImageByHolder(img.holder(), xPt, yPt, &gopdf.Rect{W: wPt, H: hPt}); err == nil {
							loaded = true
						}
					}
				}
			}

		case domain.ElementTypeRectangle:
			style := ""
			if el.Properties.FillColor != "" {
				if rCol, gCol, bCol, err := ParseHexColor(el.Properties.FillColor); err == nil {
					pdf.SetFillColor(rCol, gCol, bCol)
					style += "F"
				}
			}
			if el.Properties.StrokeColor != "" && el.Properties.StrokeWidth > 0 {
				if rCol, gCol, bCol, err := ParseHexColor(el.Properties.StrokeColor); err == nil {
					pdf.SetStrokeColor(rCol, gCol, bCol)
					lineWidth := domain.MmToPt(el.Properties.StrokeWidth)
					if lineWidth <= 0 {
						lineWidth = 1.0
					}
					pdf.SetLineWidth(lineWidth)
					style += "D"
				}
			}
			if style != "" {
				pdf.RectFromUpperLeftWithStyle(xPt, yPt, wPt, hPt, style)
			}

		case domain.ElementTypeCircle:
			if el.Properties.FillColor != "" {
				if rCol, gCol, bCol, err := ParseHexColor(el.Properties.FillColor); err == nil {
					pdf.SetFillColor(rCol, gCol, bCol)
				}
			}
			if el.Properties.StrokeColor != "" && el.Properties.StrokeWidth > 0 {
				if rCol, gCol, bCol, err := ParseHexColor(el.Properties.StrokeColor); err == nil {
					pdf.SetStrokeColor(rCol, gCol, bCol)
					lineWidth := domain.MmToPt(el.Properties.StrokeWidth)
					if lineWidth <= 0 {
						lineWidth = 1.0
					}
					pdf.SetLineWidth(lineWidth)
				}
			}
			pdf.Oval(xPt, yPt, xPt+wPt, yPt+hPt)

		case domain.ElementTypeLine:
			if el.Properties.StrokeColor != "" {
				if rCol, gCol, bCol, err := ParseHexColor(el.Properties.StrokeColor); err == nil {
					pdf.SetStrokeColor(rCol, gCol, bCol)
					lineWidth := domain.MmToPt(el.Properties.StrokeWidth)
					if lineWidth <= 0 {
						lineWidth = 1.0
					}
					pdf.SetLineWidth(lineWidth)
					pdf.Line(xPt, yPt, xPt+wPt, yPt+hPt)
				}
			}

		case domain.ElementTypeQRCode:
			content := ReplacePlaceholders(el.Properties.QRContent, variables)
			if content != "" {
				qrLevel := qrcode.Medium
				switch el.Properties.RecoveryLevel {
				case "L":
					qrLevel = qrcode.Low
				case "Q":
					qrLevel = qrcode.High
				case "H":
					qrLevel = qrcode.Highest
				}

				qrPng, err := qrcode.Encode(content, qrLevel, 512)
				if err == nil {
					imgHolder, err := gopdf.ImageHolderByBytes(qrPng)
					if err == nil {
						_ = pdf.ImageByHolder(imgHolder, xPt, yPt, &gopdf.Rect{W: wPt, H: hPt})
					}
				}
			}

		case domain.ElementTypeText:
			fontFamily := el.Properties.FontFamily
			canonicalName, fontBytes, foundBytes := r.fontResolver.ResolveBytes(fontFamily)
			if foundBytes {
				if !registeredFonts[canonicalName] {
					if err := pdf.AddTTFFontData(canonicalName, fontBytes); err == nil {
						registeredFonts[canonicalName] = true
					}
				}
				fontFamily = canonicalName
			} else {
				resolvedName, fontPath, found := r.fontResolver.Resolve(fontFamily)
				if found {
					if !registeredFonts[resolvedName] && fontPath != "" {
						if err := pdf.AddTTFFont(resolvedName, fontPath); err == nil {
							registeredFonts[resolvedName] = true
						}
					}
					fontFamily = resolvedName
				} else {
					fontFamily = ensureDefaultFont()
				}
			}

			if !registeredFonts[fontFamily] {
				fontFamily = ensureDefaultFont()
			}

			fontSize := el.Properties.FontSizePt
			if fontSize <= 0 {
				fontSize = 12
			}

			// Activate font safely
			if err := pdf.SetFont(fontFamily, "", int(fontSize)); err != nil {
				if def := ensureDefaultFont(); def != "" && registeredFonts[def] {
					_ = pdf.SetFont(def, "", int(fontSize))
					fontFamily = def
				}
			}

			// Text color
			if el.Properties.Color != "" {
				if rCol, gCol, bCol, err := ParseHexColor(el.Properties.Color); err == nil {
					pdf.SetTextColor(rCol, gCol, bCol)
				}
			} else {
				pdf.SetTextColor(0, 0, 0)
			}

			rawText := ReplacePlaceholders(el.Properties.Text, variables)
			if el.Properties.TextTransform != "" {
				rawText = ApplyTextTransform(rawText, el.Properties.TextTransform)
			}
			lineHeightRatio := el.Properties.LineHeight
			if lineHeightRatio <= 0 {
				lineHeightRatio = 1.2
			}

			// Split by newline and filter out empty/whitespace lines (e.g. from optional placeholders like event_city)
			// so they do not distort vertical centering and line spacing.
			var paragraphs []string
			for _, p := range strings.Split(rawText, "\n") {
				tp := strings.TrimSpace(p)
				if tp != "" {
					paragraphs = append(paragraphs, tp)
				}
			}

			if len(paragraphs) == 0 {
				continue
			}

			// Handle multi-line wrapping or newline splitting
			var lines []string
			if el.Properties.FitMode == "wrap" && wPt > 0 {
				for _, para := range paragraphs {
					wrappedLines, err := pdf.SplitTextWithWordWrap(para, wPt)
					if err == nil && len(wrappedLines) > 0 {
						for _, wl := range wrappedLines {
							twl := strings.TrimSpace(wl)
							if twl != "" {
								lines = append(lines, twl)
							}
						}
					} else {
						lines = append(lines, para)
					}
				}
			} else {
				// Single line or explicit newlines
				if el.Properties.FitMode == "shrink" && wPt > 0 {
					// Shrink-to-fit calculation
					minFontSize := 8.0
					for fontSize > minFontSize {
						allFit := true
						for _, p := range paragraphs {
							textW, err := pdf.MeasureTextWidth(p)
							if err == nil && textW > wPt {
								allFit = false
								break
							}
						}
						if allFit {
							break
						}
						fontSize -= 1.0
						_ = pdf.SetFont(fontFamily, "", int(fontSize))
					}
				}
				lines = paragraphs
			}

			if len(lines) == 0 {
				continue
			}

			// Calculate vertical alignment offset within bounding box
			lineHeightPt := fontSize * lineHeightRatio
			totalTextHeight := float64(len(lines)) * lineHeightPt
			vertOffset := 0.0

			if hPt > totalTextHeight {
				switch el.Properties.VerticalAlignment {
				case "middle":
					vertOffset = (hPt - totalTextHeight) / 2
				case "bottom":
					vertOffset = hPt - totalTextHeight
				default: // "top"
					vertOffset = 0
				}
			}

			for lineIdx, line := range lines {
				lineY := yPt + vertOffset + (float64(lineIdx) * lineHeightPt)

				// Horizontal alignment offset
				alignOffset := 0.0
				textWidth, err := pdf.MeasureTextWidth(line)
				if err == nil && wPt > textWidth {
					switch el.Properties.Alignment {
					case "center":
						alignOffset = (wPt - textWidth) / 2
					case "right":
						alignOffset = wPt - textWidth
					}
				}

				pdf.SetX(xPt + alignOffset)
				pdf.SetY(lineY)
				_ = pdf.Cell(nil, line)
			}
		}
	}

	if _, err := pdf.WriteTo(w); err != nil {
		return fmt.Errorf("failed to generate PDF stream: %w", err)
	}
	return nil
}

// RenderToFile streams the rendered PDF atomically to the specified file path
// (temp file + fsync + rename) without buffering the whole document in memory first.
func (r *GoPDFRenderer) RenderToFile(doc *domain.Document, variables map[string]string, assetRootDir string, outputPath string) error {
	dir := filepath.Dir(outputPath)
	base := filepath.Base(outputPath)
	sanitizedBase := filesystem.SanitizeFilename(base)
	finalPath := filepath.Join(dir, sanitizedBase)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".render-*.pdf.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp output file: %w", err)
	}
	tmpPath := tmp.Name()
	committed := false
	defer func() {
		if !committed {
			_ = tmp.Close()
			_ = os.Remove(tmpPath)
		}
	}()

	bw := bufio.NewWriterSize(tmp, 256*1024)
	if err := r.renderTo(bw, doc, variables, assetRootDir); err != nil {
		return err
	}
	if err := bw.Flush(); err != nil {
		return fmt.Errorf("failed to write PDF: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("failed to sync PDF: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("failed to close PDF: %w", err)
	}
	if err := os.Chmod(tmpPath, 0644); err != nil {
		return err
	}
	// Windows: os.Rename fails if destination exists on some setups; remove first.
	if _, err := os.Stat(finalPath); err == nil {
		_ = os.Remove(finalPath)
	}
	if err := os.Rename(tmpPath, finalPath); err != nil {
		return fmt.Errorf("failed to finalize PDF: %w", err)
	}
	committed = true
	return nil
}

// Ensure interface implementation at compile time
var _ Renderer = (*GoPDFRenderer)(nil)
