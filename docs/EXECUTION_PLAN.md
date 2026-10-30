# Execution Plan: Phase 5 — Batch Generation

## Phase Being Executed
**Phase 5 — Batch Generation**

## Scope
1. **Domain Layer**:
   - `internal/domain/job.go`:
     - Job status constants & state transitions: `pending` -> `validating` -> `running` -> `completed` / `completed_with_errors` / `failed` / `cancelling` -> `cancelled`.
     - Generation Job model (`ID`, `ProjectID`, `TemplateID`, `Status`, `TotalCount`, `CompletedCount`, `FailedCount`, `OutputPath`, `ErrorMessage`, `CreatedAt`, `UpdatedAt`).
     - Batch configuration (concurrency limit, selected recipient IDs, output directory, export ZIP toggle).
     - Progress event schema (`JobID`, `CurrentItem`, `Total`, `Completed`, `Failed`, `Status`, `Percentage`, `CurrentRecipientName`, `Message`).
     - Job Item result entity for individual tracking & retry support.
2. **Repository Layer**:
   - `internal/repository/sqlite/job_repo.go`:
     - Generation job CRUD.
     - State machine transition enforcement (invalid transitions rejected).
     - Job items status tracking (`generation_job_items`).
     - Query pending/interrupted jobs for startup recovery.
3. **Generator Pipeline & Worker Pool**:
   - `internal/generator/certificate/batch_generator.go`:
     - Bounded worker pool with default concurrency 1 (safe memory footprint for 4 GB RAM).
     - Context cancellation (`context.WithCancel`) checking cancellation between items.
     - Individual PDF generation using `pdf.Renderer` without holding all PDFs in memory.
     - Streaming ZIP archive creation via `archive/zip` directly into output file or temporary stream.
     - `export-summary.csv` generation (Recipient ID, Certificate ID, Recipient Name, Output File, Status, Error Message).
     - Temporary directory creation & cleanup (`defer os.RemoveAll(tempDir)`).
     - Retry logic for failed records in a job.
4. **Desktop & Wails Integration**:
   - Progress event broadcasting via Wails event system or callback channels.
   - Job service exposure in `app/application.go`.
   - Binding generation via `wails3 generate bindings`.
5. **Frontend Experience**:
   - `BatchGenerationModal.tsx` & `GenerationHistoryView.tsx`:
     - Recipient selection & sample preview.
     - Concurrency slider (1 to 4, default 1).
     - Real-time progress bar (percentage, completed, failed, current item name).
     - Cancel job button.
     - Completed state with download / open folder button and summary table.
     - Retry failed records button.
6. **Testing & Benchmarks**:
   - 1 certificate test.
   - 100 certificates batch test.
   - 1.000 certificates batch test (with memory and duration recording).
   - Cancellation test.
   - Retry failed records test.
   - ZIP archive integrity check.
   - Temporary file cleanup verification.
   - Invalid state machine transition test.

## Out of Scope
- Native OS printing dialogs and backup/restore archives (Phase 6).

## Acceptance Criteria
- [ ] Bounded worker pool generates certificates one-by-one without holding all PDFs in RAM.
- [ ] Concurrency defaults to 1.
- [ ] Context cancellation halts generation cleanly and transitions job state to `cancelled`.
- [ ] Temporary working directories are cleaned up on success, failure, and cancellation.
- [ ] Generated ZIP archive contains individual PDFs and `export-summary.csv` with valid ZIP integrity.
- [ ] Failed records are tracked individually and can be retried.
- [ ] 1.000 certificate benchmark completes within memory limits (< 100 MB RAM peak).
- [ ] All automated tests pass without warnings or race conditions.
