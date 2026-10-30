# Implementation Status

## Repository Overview
The repository is set up with Go backend and React + TypeScript + Vite frontend.
- Backend module: `certificate-generator`
- Git initialized: Yes
- Standalone desktop binary: `build/bin/certificate-generator` (22 MB)

## Detected Technology Versions
- Operating System: macOS Darwin arm64 (Apple Silicon)
- Go Version: `go1.26.5 darwin/arm64`
- Node.js Version: `v24.11.0`
- npm Version: `11.6.1`
- Git Version: `2.51.0`
- Desktop Runtime: `Wails v3` (`github.com/wailsapp/wails/v3@v3.0.0`)
- SQLite Driver: `modernc.org/sqlite v1.44.3` (pure Go, CGO-free, embedded)
- PDF Engine: `github.com/signintech/gopdf v0.32.0`
- QR Code Engine: `github.com/skip2/go-qrcode`
- Excel Engine: `github.com/xuri/excelize/v2 v2.11.0`
- Archive Engine: standard `archive/zip`
- Frontend: `React 19`, `konva v10.9.0`, `react-konva v19.3.2`, `vite v8.3.5`, `vitest v5.0.5`, `tailwindcss v4`, `zustand v5.0.16`

## Existing Architecture
- `app/`: Desktop application lifecycle, service exposure, database connection management, and rendering coordination.
- `internal/domain/`:
  - Canonical shared document model (`schemaVersion: 1`, millimeter coordinates).
  - Unit conversions (`mm` <-> `pt`, `mm` <-> `px`).
  - Project entity, validation, and manifest schema (`project.json`).
  - Template entity, default paper presets, and validation.
  - Recipient entity, validation, column mapping, and conflict strategies.
  - Generation job entity, state machine transitions, job item results, batch config, and progress event schemas.
  - Application settings model and storage info.
- `internal/importer/`:
  - `CSVParser`: CSV parser with automatic delimiter detection (comma, semicolon, tab), multiline quotes support, and empty line filtering.
  - `ExcelParser`: Streaming XLSX parser with worksheet selection, row iterator, macro file rejection (`.xlsm`), and file size limits.
- `internal/storage/filesystem/`:
  - Standardized application data directory (`CertificateGenerator/`).
  - Scaffolding of project directories (`projects/{id}/templates`, `assets/...`, `exports`).
  - Path traversal protection (`SafeResolve`).
  - Cross-platform filename sanitization.
  - Atomic write operations.
- `internal/repository/sqlite/`:
  - Pure-Go SQLite connection handling (WAL mode, busy timeout).
  - Embedded SQL migration runner (`001_initial_schema.sql`, `002_add_event_details_to_projects.sql`, `003_generation_job_items.sql`).
  - `ProjectRepository`: CRUD, count, recent projects query, archive toggle.
  - `TemplateRepository`: CRUD and project templates query.
  - `RecipientRepository`: CRUD, pagination, search, bulk deletion, and transactional batch insertion with conflict resolution (skip, update, fail).
  - `JobRepository`: CRUD, strict state transitions, item results tracking, failed items query, and interrupted job discovery.
  - `SettingsRepository`: Key-value app preferences persistence.
- `internal/services/`:
  - `ProjectService`: Transactional coordination between SQLite and Filesystem with automatic rollback to avoid orphan rows/folders.
  - `TemplateService`: Template validation, SQLite persistence, and disk synchronization.
  - `RecipientService`: File metadata inspection, preview, mapping validation, and transactional import execution.
  - `JobService`: Batch generation execution, status queries, cancellation, and restart recovery.
  - `SettingsService`: App settings management.
- `internal/renderer/`:
  - `font.Resolver`: Local TTF font discovery, family resolution (`LiberationSerif`, `Ubuntu`, `sans-serif`, `serif`, `monospace`), and safe fallbacks.
  - `pdf.GoPDFRenderer`: Production GoPDF vector renderer supporting:
    - Precise millimeter-to-point conversion.
    - TrueType font embedding and Unicode Indonesian accents.
    - Dynamic variable resolver (regex identifier whitelist).
    - Horizontal alignment (left, center, right) and vertical alignment (top, middle, bottom).
    - Shrink-to-fit (`fitMode: shrink`) and multi-line paragraph wrapping (`fitMode: wrap`).
    - Vector shapes (rectangles, circles via `pdf.Oval`, lines) with Z-Index layer ordering.
    - Dynamic QR code generation (512x512) with error correction levels.
    - PNG alpha transparency and JPEG backgrounds.
    - Safe atomic writes with temporary file move and filename sanitization.
    - Structural PDF validator (`ValidatePDF`) checking `%PDF-` header, MediaBox dimensions, and minimum size.
- `internal/generator/certificate/`:
  - `BatchGenerator`:
    - Bounded worker pool with default concurrency 1 (configurable up to 4).
    - Context-based cancellation checking `jobCtx.Done()` between items.
    - One-by-one PDF rendering avoiding high RAM consumption.
    - Streaming ZIP archive generation with Deflate compression.
    - Automatic `export-summary.csv` generation inside ZIP.
    - Guaranteed cleanup of temporary working directories.
    - Retry logic for failed records.
- `frontend/`:
  - App Shell layout with sidebar navigation, active project badge, and global error boundary.
  - Lightweight Dashboard with summary metrics, recent projects, and storage disk usage.
  - Projects view with search, status filters (All, Active, Archived), create, rename, duplicate, archive, and delete modals.
  - Templates view with template listing, duplication, deletion, and visual designer launcher.
  - Template Designer:
    - React-Konva visual canvas editor with millimeter canonical coordinates.
    - TopBar with paper presets, undo/redo, zoom, grid toggle, snap toggle, and save/preview.
    - LeftPanel with elements catalog and Layer manager.
    - RightPanel property inspector.
    - Bounded undo/redo history (max 50 snapshots).
    - Grid and alignment guides snapping.
    - Global keyboard shortcuts and debounced autosave.
    - Interactive preview modal with Indonesian placeholder data and test PDF vector rendering.
  - Recipients view:
    - Table with search, pagination (10, 25, 50, 100), bulk selection, manual CRUD, and delete all.
    - Import Wizard Modal: 5-step wizard.
  - Generation History view:
    - List of batch generation jobs with status indicators.
    - Output location badge with file path.
    - Rendered certificate items summary table.
    - Trigger modal for new batch generation.
    - Retry failed records button.
  - Batch Generation Modal:
    - Template selector and recipient count indicator.
    - Concurrency worker slider (1-4).
    - Output options (Streaming ZIP vs Individual PDFs).
    - Live progress bar with real-time percentage and cancel button.
  - Settings view for user preferences.
  - Zustand state stores (`useProjectStore`, `useDesignerStore`) with type-safe Wails v3 bindings and mock testing fallback.

## Completed Phases
- **Phase 0 — Foundation & Technical Proof of Concept**: Completed (**PASS**). Reference artifact in `tests/fixtures/poc_output.pdf`.
- **Phase 1 — Application Foundation**: Completed (**PASS**). Full app shell, project CRUD, filesystem rollback, settings, restart persistence, and production desktop binary.
- **Phase 2 — Template Designer**: Completed (**PASS**). Interactive visual canvas editor React-Konva, millimeter canonical coordinate model, elements, transformer, layer manager, properties inspector, undo/redo, alignment guides, autosave, and preview modal.
- **Phase 3 — Recipient Management**: Completed (**PASS**). Recipient CRUD, pagination, search, bulk selection, CSV & XLSX import with worksheet selection, column mapping, date/email validation, duplicate detection, transactional batch storage, and 1,000-row fixture import.
- **Phase 4 — Certificate Rendering**: Completed (**PASS**). Production GoPDF vector renderer, font resolver, Unicode Indonesian support, text measurement, horizontal/vertical alignment, shrink-to-fit, word wrapping, vector shapes, QR code embedding, safe atomic writes, filename sanitization, structural PDF validation, and 6 visual PDF fixture artifacts.
- **Phase 5 — Batch Generation**: Completed (**PASS**). Bounded worker pool (concurrency 1-4), context cancellation, streaming ZIP creation, `export-summary.csv`, temporary directory cleanup, retry failed items, interrupted job recovery, and benchmark 1.000 certificates (~5.9 ms/cert).

## Missing Features / Roadmap
- **Phase 6 — Printing, Backup, Restore, and Polish**: Native printing, project backup/restore with ZIP slip protection, recovery from interrupted jobs.
- **Phase 7 — Cross-Platform QA and Release**: Packaging, signing guides, and release notes.

## Technical Risks & Mitigations
- *Risk*: Memory exhaustion when rendering thousands of certificates simultaneously.
  *Mitigation*: Bounded worker pool (concurrency 1 default) renders certificates sequentially to temporary storage and streams directly to ZIP, keeping memory allocated under 100 MB RAM.
- *Risk*: Data inconsistency when application crashes during a batch generation.
  *Mitigation*: Transactional job item logging and startup recovery (`CheckInterruptedJobs`) clean up orphaned state cleanly.

## Current Test Status
- Backend Go tests (`go test -count=1 -v -race ./...`): **38 passed, 0 failed, 0 race warnings**.
- Go static checks (`go vet ./...`): **PASS**.
- Frontend unit tests (`npx vitest run`): **24 passed, 0 failed**.
- Frontend production build (`npm run build`): **PASS** (minified bundle 759 kB).
- Go desktop binary build (`go build`): **PASS** (`build/bin/certificate-generator` 22 MB).

## Recommended Next Phase
**Phase 6 — Printing, Backup, Restore, and Polish**
