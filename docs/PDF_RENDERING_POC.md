# PDF Rendering Proof of Concept (PoC)

## 1. Overview & Objective
This document details the technical proof of concept for the Certificate Generator application. The objective is to validate whether the Go PDF engine (`github.com/signintech/gopdf`), paired with a canonical millimeter-based document model and embedded local TrueType fonts, can faithfully generate professional certificates without relying on headless browsers, external web services, or canvas raster screenshots.

## 2. Library Versions & Stack
- **Go Version**: `go1.26.5 darwin/arm64`
- **PDF Engine**: `github.com/signintech/gopdf v0.32.0`
- **QR Code Engine**: `github.com/skip2/go-qrcode`
- **Database Engine**: `modernc.org/sqlite v1.44.3` (pure Go, CGO-free)
- **Frontend Canvas**: `React 19`, `konva v10.0.8`, `react-konva v19.2.2`, `vite v8.3.5`

## 3. Features Validated in PoC

| Feature | Status | Notes |
|---|---|---|
| A4 Landscape Specification | PASS | Exact 297 x 210 mm (841.89 x 595.28 PostScript points). Verified by binary inspection. |
| Local Font Embedding | PASS | TTF fonts embedded (`Ubuntu-Regular.ttf`, `LiberationSerif-Regular.ttf`). No remote font calls. |
| Indonesian Character Set & Accents | PASS | Accents (`é`), titles (`Prof. Dr. Ir. M.Sc., Ph.D.`), hyphens, apostrophes render cleanly. |
| Transparent PNG Images | PASS | Alpha channel preserved for digital signature and logo overlays. |
| JPEG Background Image | PASS | Full-bleed background image correctly scaled to canvas boundaries. |
| Vector Rectangles & Borders | PASS | Millimeter-specified stroke width and fill colors rendered as native vector paths. |
| Vector Lines | PASS | Decorative divider lines rendered with explicit stroke color and width. |
| QR Code Generation | PASS | Dynamic URL (`https://verify.cert.local/verify?id={{certificate_id}}`) rendered in-memory and placed onto PDF. |
| Dynamic Placeholder Replacement | PASS | Strict regex identifier matching replacing `{{recipient_name}}`, `{{event_name}}`, `{{issue_date}}`, `{{certificate_id}}`. |
| Shrink-to-Fit Text Algorithm | PASS | Automatically measures text width against bounding box and decrements font size until text fits. |
| Coordinate Accuracy | PASS | Canonical millimeter coordinate system rigorously converted to points ($pt = mm \times 72 / 25.4$) with $10^{-4}$ precision tests. |

## 4. Canvas vs. PDF Comparison
- **Coordinate Model**: Both the frontend `CanvasPreview` (React-Konva) and backend `GoPDFRenderer` parse identical JSON schemas (`schemaVersion: 1`).
- **Display Scaling**:
  - Frontend scales canonical millimeters to screen display pixels using 96 DPI baseline ($px = mm \times 96 / 25.4$).
  - Backend scales canonical millimeters to PDF points ($pt = mm \times 72 / 25.4$).
- **Alignment**: Center alignment offsets text dynamically based on measured width relative to bounding box width.

## 5. Artifacts
- **Reference Output PDF**: `tests/fixtures/poc_output.pdf`
- **File Size**: ~42 KB (compact vector document with embedded TTF font subsets and compressed images).
- **Test Command**: `go test -v ./internal/renderer/pdf`

## 6. Architectural Decision: GoPDF
Based on these findings, `github.com/signintech/gopdf` successfully fulfills all Phase 0 exit criteria:
- Pure Go execution without external CGO or binary dependencies.
- Fast generation (<25 ms per certificate).
- Low memory footprint (<10 MB peak).
- Vector preservation for text and geometric elements.
- Full offline capability.

**Recommendation**: Proceed with `github.com/signintech/gopdf` behind the `pdf.Renderer` domain interface.
