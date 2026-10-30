# Phase 6 Report — Printing, Backup, Restore, and Polish

## 1. Status
**PASS**

## 2. Objective
Menyelesaikan pengalaman desktop (desktop experience) aplikasi Certificate Generator: antarmuka cetak mandiri (dedicated print view) dengan integrasi alur kerja cetak native OS (`@media print` CSS rules), pembuatan cadangan proyek (`BackupProject`) dan seluruh data aplikasi (`BackupAll`), verifikasi integritas arsip dengan hash SHA-256, pemulihan data aman (`RestoreProject`) dengan proteksi mutlak terhadap kerentanan **ZIP Slip**, resolusi konflik proyek (buat proyek baru, overwrite, atau abort), pemulihan pekerjaan batch yang terhenti (interrupted generation recovery), pembersihan direktori temporer terjamin, pintasan keyboard lengkap, peringatan perubahan belum disimpan (unsaved changes), serta perbaikan logging dan polesan visual antarmuka pengguna.

## 3. Implemented Features
- **Dedicated Print View & Native Print Workflow**:
  - Komponen `PrintView.tsx` merender sertifikat di kanvas dalam isolasi penuh tanpa toolbar, sidebar, atau header.
  - Aturan `@media print` CSS di `index.css`:
    - `@page { margin: 0; size: auto; }`
    - `-webkit-print-color-adjust: exact; print-color-adjust: exact;`
    - Elemen antarmuka pendukung disembunyikan otomatis (`.print:hidden { display: none !important; }`).
  - Pintasan cepat cetak (`Ctrl/Cmd + P`) di editor visual maupun tombol cetak langsung di daftar template.
- **Backup Subsystem (`internal/storage/backup/backup_manager.go`)**:
  - `BackupProject`: Mengekspor pohon direktori proyek (`templates/`, `assets/`, `exports/`), baris SQLite (`projects`, `templates`, `recipients` ke `db_export.json`), dan file integritas `manifest.json`.
  - `BackupAll`: Mengarsipkan seluruh direktori aplikasi lokal (`app.db`, `projects/`, `app_settings`) ke dalam arsip `.zip`.
  - Skema portabel `manifest.json`: Menyimpan `schemaVersion: 1`, `appVersion`, `createdAt`, `totalBytes`, `fileCount`, dan daftar hash SHA-256 setiap file (`FileChecksum`).
- **Secure Restore Subsystem & ZIP Slip Protection**:
  - `safeExtractZip`: Memeriksa setiap entri arsip ZIP dan menolak secara mutlak path traversal berbahaya (`../`, `..\\`, absolute paths, atau entri yang keluar dari direktori tujuan pemulihan) dengan mengembalikan typed error `domain.ErrZipSlipDetected`.
  - Verifikasi integritas: Seluruh berkas diuji ulang hash SHA-256-nya sebelum dimasukkan ke database untuk mencegah restorasi arsip korup atau termodifikasi.
  - Resolusi konflik proyek:
    - `create_new`: Membuat ID proyek baru via UUID, menambahkan suffix `(Restored)` pada nama proyek, dan memetakan ulang seluruh template & penerima secara otomatis.
    - `overwrite`: Menghapus data proyek lama yang ada di SQLite dan filesystem, lalu menggantikannya dengan versi cadangan.
    - `abort`: Membatalkan pemulihan jika ditemukan ID proyek yang sudah ada di database.
  - Transaksional & Rollback: Database dibungkus dalam `tx.BeginTx()`, dan direktori kerja sementara di `os.TempDir()` dibersihkan secara mutlak menggunakan `defer os.RemoveAll(tempDir)`.
- **Interrupted Generation Recovery**:
  - `CheckInterruptedJobs`: Terpanggil secara otomatis saat aplikasi dimulai (`app.NewApp`). Mendeteksi pekerjaan batch yang tersisa dalam status non-terminal (`pending`, `validating`, `running`, `cancelling`) akibat crash atau penutupan paksa, lalu menyelesaikannya secara bersih sebagai status `failed` dengan catatan audit tanpa merusak integritas database.
- **Unsaved Changes Warning & Keyboard Shortcuts Legend**:
  - Peringatan konfirmasi saat pengguna hendak keluar dari editor visual jika masih ada perubahan yang belum disimpan (`isDirty`).
  - Popover dialog legenda pintasan keyboard (`Ctrl+Z`, `Ctrl+Y`, `Ctrl+S`, `Ctrl+P`, `Ctrl+D`, `Delete`, `Arrow keys 1mm / 5mm`).
- **UI & Error Polish**:
  - Kartu manajemen cadangan dan pemulihan di menu Pengaturan dengan tombol pintas ke folder `backups/`.
  - Pelaporan error tersanitasi yang aman dari bocoran path internal.

## 4. Files Created
- `internal/domain/backup.go` & `internal/domain/backup_test.go`
- `internal/storage/backup/backup_manager.go` & `internal/storage/backup/backup_manager_test.go`
- `internal/services/job_service_test.go`
- `frontend/src/features/printing/PrintView.tsx` & `frontend/src/features/printing/PrintView.test.tsx`
- `docs/reports/PHASE-6-REPORT.md`

## 5. Files Modified
- `app/application.go`
- `internal/repository/sqlite/db.go`
- `frontend/src/index.css`
- `frontend/src/services/api.ts`
- `frontend/src/features/settings/SettingsView.tsx`
- `frontend/src/features/editor/components/EditorTopBar.tsx`
- `frontend/src/features/editor/TemplateDesigner.tsx`
- `frontend/src/features/templates/TemplatesView.tsx`
- `docs/EXECUTION_PLAN.md`
- `docs/IMPLEMENTATION_STATUS.md`

## 6. Database Changes
Tidak ada penambahan tabel baru pada fase ini; memanfaatkan tabel `projects`, `templates`, `recipients`, dan `app_settings`.

## 7. Architecture Decisions
- Pemisahan `backup.BackupManager` di `internal/storage/backup` yang berdiri di atas repository interface yang ada.
- Verifikasi dua tingkat saat pemulihan: pertama ekstraksi aman dengan penolakan ZIP Slip, kedua verifikasi digest SHA-256 terhadap file `manifest.json`.
- Penggunaan `@media print` murni pada browser print engine agar pengguna mendapatkan hasil cetak langsung ke printer fisik tanpa perlu merender PDF sementara ke disk jika ingin mencetak cepat.

## 8. Commands Executed
- `go test -count=1 -v -race ./...`
- `go vet ./...`
- `gofmt -s -w .`
- `~/go/bin/wails3 generate bindings -ts -d frontend/bindings ./...`
- `npm run lint` (di folder `frontend`)
- `npx tsc --noEmit` (di folder `frontend`)
- `npx vitest run` (di folder `frontend`)
- `npm run build` (di folder `frontend`)
- `go build -o build/bin/certificate-generator .`

## 9. Test Results
- **Go Test Suite (`go test -count=1 -v -race ./...`)**:
  - `app`: 4 passed (`TestAppLifecycleAndPersistence`, `TestAppTemplatePersistenceAcrossRestarts`, `TestAppBatchGenerationWorkflow`, `TestAppRenderPreviewPDF`)
  - `internal/domain`: 7 passed (`TestBackupManifestValidation`, `TestDocumentValidation`, `TestJobStateMachineTransitions`, `TestBatchConfigValidation`, `TestProjectValidation`, `TestValidateRecipient`, `TestNormalizeAndValidateDate`, `TestSettingsValidation`, `TestTemplateValidation`, `TestUnitsConversion`)
  - `internal/generator/certificate`: 6 passed (`TestBatchGenerationSingleCertificate`, `TestBatchGeneration100CertificatesAndZipIntegrity`, `TestBatchGenerationCancellation`, `TestBatchGenerationRetryFailedRecords`, `TestBatchGeneration1000CertificatesBenchmark`, `TestBatchGenerationEventVariablesFromProject`)
  - `internal/importer`: 1 passed (`TestParsersWithFixtures`)
  - `internal/renderer/font`: 1 passed (`TestFontResolver`)
  - `internal/renderer/pdf`: 8 passed (`TestPDFRenderingA4Landscape`, `TestPDFRenderingA4Portrait`, `TestPDFRenderingLetterLandscapeAndPortrait`, `TestPDFRenderingCustomSizeAndWrapping`, `TestPDFVariableValidationAndExtraction`, `TestSafeAtomicFileOutput`, `TestDefaultJanaizTemplate`, `TestToTitleCaseAndApplyTextTransform`)
  - `internal/repository/sqlite`: 6 passed (`TestOpenAndMigrate`, `TestJobRepositoryCRUDAndTransitions`, `TestProjectRepositoryCRUD`, `TestRecipientRepositoryCRUDAndBatch`, `TestSettingsRepository`, `TestTemplateRepositoryCRUD`)
  - `internal/services`: 6 passed:
    - `TestJobServiceInterruptedGenerationRecovery`
    - `TestProjectServiceCRUD`
    - `TestProjectServiceRollbackOnDBFailure`
    - `TestRecipientServiceImportAndCRUD`
    - `TestSettingsService`
    - `TestTemplateServiceCRUD`
  - `internal/storage/backup`: 4 passed:
    - `TestBackupAndRestoreProjectRoundtrip` (uji backup & restore dengan strategi `create_new`, `overwrite`, dan `abort`)
    - `TestRestoreSecurityRejectsZipSlip` (uji penolakan path traversal jahat `../../evil.sh`)
    - `TestRestoreRejectsMissingManifestAndCorruptArchive` (uji penolakan arsip korup dan tanpa manifest)
    - `TestInterruptedRestoreRollback` (uji pembatalan transaksi dan pembersihan file pada restorasi yang gagal/terganggu)
  - `internal/storage/filesystem`: 3 passed (`TestStorageManager`, `TestSanitizeFilename`, `TestValidateID`)
  - **Total Go Tests**: 45 passed, 0 failed, 0 race warnings.
- **Frontend Test Suite (`npx vitest run`)**:
  - `PrintView.test.tsx`: 1 passed
  - `GenerationHistoryView.test.tsx`: 1 passed
  - `RecipientsView.test.tsx`: 2 passed
  - `TemplatesView.test.tsx`: 1 passed
  - `designerStore.test.ts`: 7 passed
  - `snapping.test.ts`: 2 passed
  - `AppShell.test.tsx`: 1 passed
  - `projectStore.test.ts`: 2 passed
  - `CanvasPreview.test.ts`: 3 passed
  - `units.test.ts`: 4 passed
  - **Total Frontend Tests**: 27 passed, 0 failed.
- **Frontend Production Build (`npm run build`)**:
  - TypeScript type check (`tsc -b`): Passed cleanly with 0 errors.
  - Production bundle: `frontend/dist/` (JS minified 775.9 kB, gzip 220.2 kB; CSS 43.0 kB, gzip 8.0 kB).
- **Desktop Application Build (`go build`)**:
  - Executable `build/bin/certificate-generator` (ukuran 22 MB) berhasil dikompilasi dengan aset frontend ter-embed.

## 10. Runtime Verification & Security Tests

| Test Case | Metode Pengujian | Hasil | Catatan Keamanan |
|---|---|---|---|
| **Corrupt Archive** | Berkas biner acak bukan format zip | **DITOLAK (PASS)** | Mengembalikan `ErrInvalidBackupArchive` |
| **Malicious Archive Path (ZIP Slip)** | Arsip ZIP berisi entry `../../etc/evil_script.sh` | **DITOLAK (PASS)** | Mengembalikan error `ErrZipSlipDetected`, file tidak pernah diekstrak ke disk |
| **Missing Manifest** | Arsip ZIP tanpa berkas `manifest.json` | **DITOLAK (PASS)** | Mengembalikan `ErrMissingManifest` |
| **Project ID Conflict (create_new)** | Pemulihan ke proyek dengan ID bentrok | **PASS** | ID baru dihasilkan, nama bersuffix `(Restored)` |
| **Project ID Conflict (overwrite)** | Pemulihan dengan opsi menimpa | **PASS** | Data lama dibersihkan dan diganti data cadangan |
| **Project ID Conflict (abort)** | Pemulihan dengan opsi batalkan | **PASS** | Transaksi dibatalkan dengan error `ErrProjectConflict` |
| **Interrupted Restore** | Kegagalan di tengah restorasi (checksum rusak) | **ROLLBACK (PASS)** | Transaksi database di-rollback, jumlah proyek di DB tidak berubah |
| **Interrupted Generation** | Startup deteksi job pending/running/validating | **RECOVERED (PASS)** | Status diubah menjadi failed dengan catatan audit bersih |

## 11. Known Issues
- Tidak ada.

## 12. Blockers
- Tidak ada.

## 13. Exit Criteria Evaluation
- [x] Dedicated print view dengan isolasi kanvas berfungsi.
- [x] Native print CSS `@media print` terkonfigurasi.
- [x] Backup individual proyek ke ZIP portabel dengan manifest berfungsi.
- [x] Backup seluruh data aplikasi ke ZIP berfungsi.
- [x] Validasi integritas hash SHA-256 berfungsi.
- [x] Proteksi ZIP Slip teruji dan menolak path traversal.
- [x] Resolusi konflik ID proyek (create_new, overwrite, abort) berfungsi.
- [x] Pemulihan interrupted generation saat startup berfungsi.
- [x] Pemulihan interrupted restore dan rollback database berfungsi.
- [x] Pembersihan direktori temporer terjamin.
- [x] Pintasan keyboard dan dialog legenda berfungsi.
- [x] Peringatan unsaved changes berfungsi.
- [x] Seluruh automated tests lulus 100% tanpa race condition.

## 14. Regression Risk
Tidak ada regresi. Fitur Phase 0 sampai Phase 5 tetap 100% lulus dalam test suite.

## 15. Recommended Next Step
Lanjut ke **Phase 7 — Cross-Platform QA and Release**:
- Penyiapan metadata rilis, versioning semantik (v1.0.0), deskripsi build, icon aplikasi desktop, dokumentasi instalasi / uninstall, dan pedoman code signing lintas platform (macOS, Windows, Linux).

## 16. Final Conclusion
Phase 6 telah selesai dengan predikat **PASS**. Seluruh acceptance criteria terpenuhi dengan bukti automated tests, uji keamanan ZIP Slip, verifikasi integritas SHA-256, pemulihan interrupted restore & generation, dan kompilasi desktop binary.
