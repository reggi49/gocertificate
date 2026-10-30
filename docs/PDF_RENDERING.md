# PDF Rendering Engine Documentation

## 1. Overview
The certificate PDF rendering engine in Certificate Generator is built natively in Go using `github.com/signintech/gopdf`. It operates behind the abstracted `pdf.Renderer` domain interface and renders vector-first PDF documents directly from the canonical millimeter-based document model (`domain.Document` schemaVersion: 1) without relying on headless web browsers, external cloud services, or raster canvas screenshots.

## 2. Canonical Coordinate System & Unit Conversions
All persistent positions, margins, stroke widths, and dimensions in templates are strictly expressed in millimeters (`mm`).
The engine converts millimeters directly to PostScript / PDF points (`pt`):
$$\text{points} = \text{mm} \times \frac{72.0}{25.4}$$

### Precision Standards:
- Standard PostScript resolution: 72 points per inch (1 inch = 25.4 mm).
- Web display canvas conversion: 96 display pixels per inch (1 inch = 25.4 mm).
- Floating point conversions are covered by automated unit tests with $10^{-4}$ precision tolerance.

## 3. Supported Page Formats
The rendering pipeline supports dynamic page configurations:
- **A4 Landscape**: 297 × 210 mm (841.89 × 595.28 pt)
- **A4 Portrait**: 210 × 297 mm (595.28 × 841.89 pt)
- **Letter Landscape**: 279.4 × 215.9 mm (792.00 × 612.00 pt)
- **Letter Portrait**: 215.9 × 279.4 mm (612.00 × 792.00 pt)
- **Custom Dimensions**: Any user-defined width and height in millimeters.

## 4. Local TrueType Font Embedding
- **No Remote Font Calls**: All fonts are loaded directly from local application assets (`assets/fonts/`).
- **Resolver**: The `font.Resolver` maps requested families (e.g., `LiberationSerif`, `Ubuntu`, `sans-serif`, `serif`, `monospace`) to registered local TrueType (`.ttf`) files.
- **Fallbacks**: When a requested font is unavailable, the resolver gracefully falls back to `LiberationSerif` without interrupting generation or crashing.
- **Unicode Support**: Supports Indonesian accent characters (`é`), titles (`Prof. Dr. Ir. M.Sc., Ph.D.`), hyphens, apostrophes, and standard punctuation.

## 5. Text Measurement, Alignment & Formatting
### Horizontal Alignment:
Calculates text width dynamically via font metrics:
- **Left**: $X = X_{\text{base}}$
- **Center**: $X = X_{\text{base}} + \frac{W_{\text{box}} - W_{\text{text}}}{2}$
- **Right**: $X = X_{\text{base}} + (W_{\text{box}} - W_{\text{text}})$

### Vertical Alignment:
Calculates bounding box height and total text height ($H_{\text{text}} = \text{lines} \times \text{fontSize} \times \text{lineHeight}$):
- **Top**: $Y = Y_{\text{base}}$
- **Middle**: $Y = Y_{\text{base}} + \frac{H_{\text{box}} - H_{\text{text}}}{2}$
- **Bottom**: $Y = Y_{\text{base}} + (H_{\text{box}} - H_{\text{text}})$

### Text Sizing Modes:
1. **Shrink-to-Fit (`fitMode: shrink`)**:
   - Default for recipient names.
   - Measures text against bounding box width.
   - Incrementally steps down font size from configured size to minimum 8pt.
   - Guarantees long names are never silently clipped.
2. **Text Wrapping (`fitMode: wrap`)**:
   - Wraps multi-line paragraphs using word boundaries into lines that fit within $W_{\text{box}}$.
   - Renders each line with configurable line height.
3. **Fixed (`fitMode: fixed`)**:
   - Renders text at exact configured point size.

## 6. Vector Shapes & Graphics
- **Layer Order (Z-Index)**: Elements are sorted ascending by `ZIndex` prior to rendering. Background is drawn first, followed by borders, shapes, QR codes, logos, text, and signatures.
- **Rectangles**: Supports fill color, stroke color, stroke width (mm), and corner radius.
- **Circles**: Rendered as native vector paths via `pdf.Oval`.
- **Lines**: Rendered as vector dividers with customizable millimeter stroke width and color.
- **Hidden Elements**: Skipped when `Visible: false`.

## 7. Dynamic QR Code Generation
- Generated on-the-fly using `github.com/skip2/go-qrcode` at 512×512 resolution.
- Placeholders in `qrContent` (e.g. `https://verify.cert.local/verify?id={{certificate_id}}`) are dynamically resolved per recipient.
- Placed directly into target bounding box as high-resolution image stream.

## 8. Images & Digital Signatures
- **Alpha Transparency**: Full PNG alpha transparency support for signature overlays and logos.
- **Backgrounds**: High-resolution JPEG or PNG scaled to page boundaries.

## 9. Safe Output Handling & Atomic Writes
- All output writes use `filesystem.AtomicWriteFile`:
  1. PDF stream is written to a temporary file (`temp-*.pdf`).
  2. Temporary file is flushed and synced to disk.
  3. Atomically moved/renamed to the target destination path.
- **Filename Sanitization**: Target filenames are sanitized against Windows reserved names (`CON`, `PRN`, `AUX`, `NUL`), illegal characters (`<>:"/\|?*`), and path traversal attempts.

## 10. Automated Validation & Test Artifacts
Every generated PDF is structurally verified:
- Checks `%PDF-` header signature.
- Verifies MediaBox page dimensions against canonical millimeter specifications within 0.5 pt tolerance.
- Validates non-empty file size (> 1000 bytes).

### Reference Test Artifacts (in `tests/fixtures/output/`):
- `a4_landscape_certificate.pdf`
- `a4_portrait_certificate.pdf`
- `letter_landscape_certificate.pdf`
- `letter_portrait_certificate.pdf`
- `custom_size_certificate.pdf`
- `wrapped_multiline_certificate.pdf`
