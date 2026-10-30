# Execution Plan: Phase 6 — Printing, Backup, Restore, and Polish

## Phase Being Executed
**Phase 6 — Printing, Backup, Restore, and Polish**

## Scope
1. **Printing Subsystem**:
   - Dedicated print view component without headers, sidebars, or toolbars.
   - Native print dialog support (`window.print()`) with CSS `@media print` rules enforcing exact page sizing (A4 / Letter Landscape/Portrait, zero margins).
   - Backend PDF generation for standalone printing workflow with correct paper orientation.
2. **Backup Subsystem**:
   - `internal/storage/backup/backup_manager.go`:
     - Individual project backup (`BackupProject`) exporting project directory, manifest, database rows (project, templates, recipients), assets, and integrity checksum (`manifest.json` with SHA-256 hashes).
     - Full application backup (`BackupAll`) archiving entire data directory (`app.db`, `projects/`, `app_settings`).
     - Archive format: ZIP with versioned schema (`schemaVersion: 1`, `appVersion`, `createdAt`, `checksums`).
3. **Restore Subsystem & Security**:
   - `internal/storage/backup/restore_manager.go`:
     - Safe extraction with strict **ZIP Slip protection** (rejecting `../`, absolute paths, drive prefixes, symlinks).
     - Archive manifest & checksum integrity verification before modifying database.
     - Project ID conflict resolution policies: `create_new_id`, `overwrite`, or `abort`.
     - Transactional restore with automatic rollback on failure and complete temporary directory cleanup.
4. **Interrupted Generation Recovery**:
   - Automatic scanning on application startup (`CheckInterruptedJobs`) discovering jobs in non-terminal states and marking them failed/cancelled cleanly without data corruption.
5. **UI Polish & UX Enhancements**:
   - Unsaved changes confirmation dialog when switching templates/projects.
   - Global keyboard shortcuts dialog / legend.
   - Dedicated Print Modal in Designer & Recipients view.
   - Backup & Restore management UI in Settings with conflict options and progress indicators.
   - Accessible tooltips, focus rings, and sanitized error display.
6. **Automated Testing Suite**:
   - Corrupt archive test.
   - ZIP Slip malicious path traversal payload test (`../../evil.sh`).
   - Missing manifest test.
   - Project ID conflict resolution test (`create_new_id` vs `overwrite`).
   - Interrupted restore rollback test.
   - Interrupted generation recovery test.
   - Vitest tests for Print view and Backup/Restore modals.

## Out of Scope
- Cross-platform installer packaging & OS signing (Phase 7).

## Acceptance Criteria
- [ ] Dedicated print view renders certificate in isolation with print media CSS.
- [ ] Backup creates portable ZIP archives with manifest and SHA-256 integrity checksums.
- [ ] Restore rejects malicious ZIP Slip entries and corrupted archives.
- [ ] Restore handles ID conflicts deterministically (create new vs overwrite).
- [ ] Failed restore operations clean up all temporary files and roll back database changes.
- [ ] Interrupted jobs are recovered and cleanly logged on startup.
- [ ] All automated tests pass without warnings or race conditions.
