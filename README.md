# Certificate Generator

Aplikasi desktop modern, ringan, offline-first, dan privacy-first untuk mendesain template serta menghasilkan ribuan sertifikat penghargaan/kelulusan berkualitas profesional dalam format PDF vektor dan ekspor ZIP.

Aplikasi ini dibangun menggunakan arsitektur **Wails v3** (Go backend) dan **React + TypeScript + Vite + Tailwind CSS + React-Konva** (Frontend), dengan basis data lokal **SQLite** (pure Go, CGO-free via `modernc.org/sqlite`).

---

## 🚀 Fitur Utama

- **Offline-First & Local-First**: Seluruh template, data penerima, aset gambar, basis data, dan hasil PDF disimpan di perangkat pengguna tanpa ketergantungan cloud atau koneksi internet.
- **Project Management**:
  - Manajemen multi-proyek dengan direktori penyimpanan lokal terisolasi (`projects/{id}/`).
  - Operasi proyek lengkap: Buat, Buka, Ubah Nama, Duplikasi, Arsipkan, dan Hapus (dengan konfirmasi).
  - Mekanisme kompensasi rollback otomatis (mencegah folder orphan atau baris database korup).
- **Visual Canvas Template Designer (React-Konva)**:
  - Satuan kanonikal dokumen murni dalam milimeter (`mm`) dengan versi skema (`schemaVersion: 1`).
  - Pilihan ukuran kertas standar: A4 Landscape (297×210 mm), A4 Portrait (210×297 mm), Letter Landscape (279.4×215.9 mm), Letter Portrait (215.9×279.4 mm), dan ukuran Kustom.
  - Elemen lengkap: Teks, Persegi, Lingkaran, Garis, QR Code dinamis, dan Gambar/Logo/Tanda Tangan transparan.
  - Interaksi visual: Drag, Resize, Rotate, Kunci elemen (Lock protection), dan Visibilitas (Hide/Show).
  - Manajemen Lapisan (Layers): Pengurutan Z-Index (Bring to Front, Send to Back, Forward, Backward).
  - Snapping ke Grid milimeter dan panduan garis tengah otomatis (Alignment Guides).
  - Bounded Undo/Redo history (dibatasi 50 langkah untuk menjaga efisiensi RAM).
  - Autosave ter-debounce (2.5 detik) dan pintasan keyboard (`Ctrl+Z`, `Ctrl+Y`, `Ctrl+S`, `Ctrl+D`, `Delete`, `Arrow keys`).
  - Modal Pratinjau (Preview) interaktif dengan data variabel Indonesia dan uji render PDF backend langsung.
- **Recipient Management & Spreadsheet Import**:
  - Manajemen penerima masal (CRUD, pencarian instan, pemilihan centang masal, dan paginasi 10/25/50/100).
  - Import CSV: Deteksi otomatis delimiter (koma `,`, titik-koma `;`, tab `\t`) dan penanganan kutipan multiline.
  - Import Excel XLSX: Pemilihan lembar kerja (worksheet), pratinjau data baris terbatas, penolakan file makro berisiko (`.xlsm`), dan streaming row reader via `excelize`.
  - Pemetaan kolom interaktif (Column Mapping) dan pemetaan field kustom (Custom Fields).
  - Validasi ketat: Nama lengkap wajib, format email, normalisasi tanggal Indonesia ("30 Oktober 2026" / "30/10/2026"), dan deteksi duplikasi nomor sertifikat.
  - Transparansi error: Setiap baris yang bermasalah dicatat lengkap dengan nomor baris dan alasannya (tidak ada data yang dibuang diam-diam).
  - Strategi konflik re-impor: `skip` (lewati duplikat), `update` (perbarui data), dan `fail` (batalkan transaksi).
  - Teruji dengan dataset besar minimal **1.000 penerima** dalam waktu < 0.5 detik.
- **Vector-First PDF Rendering Pipeline**:
  - Renderer Go murni berbasis `gopdf` di balik abstraksi `pdf.Renderer`.
  - Konversi matematis presisi tinggi dari milimeter ke PDF points ($pt = mm \times 72.0 / 25.4$).
  - Font embedding lokal TrueType (`LiberationSerif`, `Ubuntu`) dan karakter aksen Indonesia (`é`, gelar akademis panjang).
  - Algoritma **Shrink-to-Fit** bertahap untuk nama penerima agar tidak terpotong, serta **Text Wrapping** untuk paragraf panjang.
  - QR Code beresolusi tinggi (512×512) dengan penggantian variabel URL dinamis per penerima.
  - Transparansi PNG murni untuk tanda tangan dan logo tanpa artifact kotak putih.
  - Penulisan file aman melalui **Atomic Write** (menulis ke file temporer lalu me-rename) dan sanitasi nama file lintas OS.
  - Validasi struktural PDF otomatis (memeriksa signature `%PDF-`, MediaBox dimensi halaman, dan integritas file).

---

## 🛠️ Arsitektur & Teknologi

| Komponen | Teknologi |
|---|---|
| **Desktop Runtime** | [Wails v3](https://v3.wails.io) |
| **Backend Language** | Go (1.24+ / 1.26) |
| **Database** | SQLite via [modernc.org/sqlite](https://gitlab.com/cznic/sqlite) (Pure Go, CGO-Free, WAL Mode) |
| **PDF Engine** | [github.com/signintech/gopdf](https://github.com/signintech/gopdf) |
| **Spreadsheet Engine** | [github.com/xuri/excelize/v2](https://github.com/xuri/excelize) & `encoding/csv` |
| **QR Code Engine** | [github.com/skip2/go-qrcode](https://github.com/skip2/go-qrcode) |
| **Frontend Framework** | React 19, TypeScript (Strict Mode), Vite |
| **Canvas Editor** | [React-Konva](https://konvajs.org/docs/react/) & Konva |
| **State Management** | [Zustand](https://github.com/pmndrs/zustand) |
| **Styling** | Tailwind CSS v4, Lucide React Icons |
| **Testing** | Go Test (dengan race detector), Vitest (JSDOM) |

---

## 📋 Prasyarat Sistem (Prerequisites)

Sebelum menjalankan atau membangun aplikasi, pastikan perangkat telah terpasang:

1. **Go Toolchain**: Versi 1.24 atau lebih baru ([Unduh Go](https://go.dev/dl/)).
2. **Node.js & npm**: Node.js v20+ atau v24+ ([Unduh Node.js](https://nodejs.org/)).
3. **Git**: Versi 2.30+ ([Unduh Git](https://git-scm.com/)).
4. **(Opsional) Path Go Binary di Shell**:
   Jika ingin memanggil CLI Go (seperti `wails3` untuk update binding TypeScript), pastikan `$HOME/go/bin` sudah ada di file `~/.zshrc` atau `~/.bashrc`:
   ```bash
   echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.zshrc
   source ~/.zshrc
   ```

---

## ⚡ Panduan Menjalankan Aplikasi (Development)

### 1. Kloning Repository & Instalasi Dependency

```bash
# Masuk ke direktori proyek
cd /path/to/gocertificate

# Download Go backend dependencies
go mod download

# Masuk ke folder frontend dan install dependensi npm
cd frontend
npm install
cd ..
```

### 2. Menjalankan Aplikasi

Aplikasi ini menggunakan arsitektur bersih Go murni yang mem-bundle aset frontend Vite secara mandiri (embedded FS), sehingga **tidak bergantung pada konfigurasi template CLI Wails (`build/config.yml`)**.

#### Opsi A: Menjalankan Aplikasi Desktop Langsung (Rekomendasi Utama)

Cukup build aset frontend sekali, lalu jalankan binary Go desktop:

```bash
# Build frontend lalu jalankan Go desktop runtime
cd frontend && npm run build && cd ..
go run .
```

Atau jika ingin mengompilasi binary mandiri ke folder `build/bin`:
```bash
# Kompilasi binary
mkdir -p build/bin
go build -o build/bin/certificate-generator .

# Jalankan aplikasi desktop
./build/bin/certificate-generator
```

> [!NOTE]
> **Mengapa tidak menggunakan `wails3 dev`?**
> Perintah `wails3 dev` membutuhkan file template boilerplate `./build/config.yml` dan `Taskfile.yml`. Karena proyek ini mengadopsi arsitektur *standalone Go embed*, menjalankan `wails3 dev` akan menghasilkan pesan error `open ./build/config.yml: no such file or directory`. Gunakan perintah `go run .` atau jalankan binary `./build/bin/certificate-generator` di atas.

#### Opsi B: Menjalankan Frontend Web Server Saja (UI Development di Browser)

Untuk iterasi cepat saat mendesain kanvas Konva, layout CSS Tailwind, atau interaksi wizard:

```bash
cd frontend
npm run dev
```
Buka browser di `http://localhost:5173`. Frontend sudah dilengkapi otomatis dengan `MockApiClient` bawaan sehingga UI, kanvas milimeter, dan alur kerja import dapat diuji dan dikembangkan secara instan dengan hot-reload Vite tanpa perlu compile Go setiap saat.

#### Opsi C: Kompilasi Executable untuk Windows (.exe)

Karena arsitektur backend Go dan database SQLite menggunakan pure-Go (`modernc.org/sqlite`) tanpa dependensi CGO, binary Windows `.exe` dapat di-cross-compile langsung dari macOS maupun Linux:

```bash
# 1. Pastikan frontend sudah di-build
cd frontend && npm run build && cd ..

# 2. Generate resource Windows (ikon dan manifest DPI)
~/go/bin/wails3 generate syso -arch amd64 -icon windows/icon.ico -manifest windows/wails.exe.manifest -info windows/info.json -out wails_windows_amd64.syso

# 3. Kompilasi binary Windows .exe (64-bit GUI mode)
mkdir -p build/bin
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -tags production -trimpath -ldflags="-w -s -H windowsgui" -o "build/bin/certificate-generator.exe" .

# 4. Bersihkan file syso sementara
rm -f wails_windows_amd64.syso
```

Hasil executable Windows berada di: `build/bin/certificate-generator.exe`. Berkas ini dapat langsung didistribusikan dan dijalankan pada sistem operasi Windows 10/11 (64-bit).

---

## 🧪 Menjalankan Pengujian Otomatis (Automated Tests)

### Pengujian Backend (Go Tests dengan Race Detector)
Menjalankan seluruh unit test, parser test, repository test, SQLite migration test, service rollback test, dan verifikasi rendering PDF:
```bash
go test -count=1 -v -race ./...
```

### Pengujian Frontend (Vitest)
Menjalankan pengujian komponen React, unit konversi milimeter/piksel, state store Zustand, dan alur kerja import wizard:
```bash
cd frontend
npx vitest run
```

### Static Analysis & Quality Gates
```bash
# Format kode Go
gofmt -s -w .

# Analisis statis Go
go vet ./...

# Typecheck & Lint Frontend
cd frontend
npm run lint
npx tsc --noEmit
cd ..
```

---

## 📦 Panduan Build Menjadi Aplikasi & Installer

### 1. Build Standalone Desktop Binary

Untuk mengompilasi aplikasi desktop mandiri di mana aset frontend Vite langsung di-embed ke dalam binary Go:

```bash
# Langkah 1: Build aset frontend produksi
cd frontend
npm run build
cd ..

# Langkah 2: Kompilasi binary desktop Go
mkdir -p build/bin
go build -o build/bin/certificate-generator .
```

Binary yang dihasilkan berada di `build/bin/certificate-generator` (ukuran sekitar ~22 MB, mandiri tanpa dependensi runtime eksternal).

Anda dapat langsung menjalankannya:
```bash
./build/bin/certificate-generator
```

---

### 2. Build Menjadi Bundle Installer Lintas Platform (Wails v3 Package)

Wails v3 menyediakan perintah packaging untuk menghasilkan bundle aplikasi resmi dan installer OS:

#### Build untuk macOS (.app / .dmg):
```bash
# Build production bundle macOS
wails3 package -platform darwin/arm64 -o build/bin/CertificateGenerator.app

# Membuat disk image installer (.dmg)
hdiutil create -volname "Certificate Generator" -srcfolder build/bin/CertificateGenerator.app -ov -format UDZO build/bin/CertificateGenerator.dmg
```

#### Build untuk Windows (.exe / installer):
```bash
# Build target Windows x64
wails3 package -platform windows/amd64 -o build/bin/CertificateGenerator.exe
```

#### Build untuk Linux (AppImage / Binary):
```bash
# Pastikan library GTK3/WebKit2GTK tersedia pada sistem Linux
wails3 package -platform linux/amd64 -o build/bin/certificate-generator-linux
```

---

## 🔄 Pembaruan Binding Wails (TypeScript Bindings)

Jika Anda menambahkan method baru pada backend Go (`app/application.go` atau `internal/domain/`):

1. Ubah source code method publik pada struct `App` di `app/application.go`.
2. Jalankan perintah binding generator resmi (Dilarang mengedit file di `frontend/bindings/` secara manual):
   ```bash
   wails3 generate bindings -ts -d frontend/bindings ./...
   ```
3. Lakukan verifikasi typecheck frontend:
   ```bash
   cd frontend && npx tsc --noEmit
   ```

---

## 📁 Struktur Direktori Repository

```
gocertificate/
├── main.go                                  # Titik masuk Wails v3 desktop runtime
├── go.mod / go.sum                          # Dependensi Go backend
├── app/
│   ├── application.go                       # Lifecycle aplikasi, database init, service exposure
│   └── application_test.go                  # Uji persistensi restart dan render PDF
├── assets/
│   └── fonts/                               # Font TrueType lokal (LiberationSerif, Ubuntu)
├── internal/
│   ├── domain/                              # Domain entities (Document, Project, Template, Recipient)
│   ├── storage/filesystem/                  # Storage manager, path traversal protection, atomic write
│   ├── repository/sqlite/                   # Pure-Go SQLite repositories dan SQL migrations
│   ├── services/                            # Application services (Project, Template, Recipient, Settings)
│   ├── renderer/
│   │   ├── pdf/                             # Production GoPDF vector renderer & structural validator
│   │   ├── font/                            # Local TTF font discovery & fallback resolver
│   │   └── image/                           # Image utilities (JPEG, PNG alpha transparency)
│   └── importer/
│       ├── csv/                             # Delimiter auto-detection CSV parser
│       └── excel/                           # Streaming XLSX parser via Excelize
├── frontend/
│   ├── package.json                         # Dependensi React, Tailwind, Konva, Vitest
│   ├── vite.config.ts / vitest.config.ts    # Konfigurasi bundler dan test runner
│   ├── bindings/                            # Generated Wails v3 TypeScript bindings (DO NOT EDIT)
│   └── src/
│       ├── components/                      # Reusable UI components & Error boundary
│       ├── features/
│       │   ├── dashboard/                   # Ringkasan proyek dan storage information
│       │   ├── projects/                    # CRUD proyek, duplikasi, arsip, konfirmasi hapus
│       │   ├── templates/                   # Manajemen template & daftar desain
│       │   ├── editor/                      # Visual Canvas Editor (React-Konva, Toolbar, Inspector)
│       │   ├── recipients/                  # Tabel penerima & Import Wizard 5-langkah
│       │   └── settings/                    # Preferensi tema, bahasa, dan interval autosave
│       ├── stores/                          # Zustand stores (useProjectStore, useDesignerStore)
│       └── lib/                             # Konversi satuan kanonikal (mm <-> pt <-> px)
├── tests/fixtures/                          # Test fixtures (1000 CSV, 1000 XLSX, multi-sheet)
│   └── output/                              # Visual PDF reference artifacts (A4, Letter, Custom, Wrap)
└── docs/                                    # Dokumentasi teknis & Laporan fase implementasi
```

---

## 📚 Dokumentasi Teknis Terkait

Untuk pemahaman mendalam tentang arsitektur dan keputusan teknis, silakan baca dokumen di folder `docs/`:

- [docs/PDF_RENDERING.md](docs/PDF_RENDERING.md) — Dokumentasi teknis engine PDF vector, font embedding, dan toleransi pengukuran.
- [docs/PDF_RENDERING_POC.md](docs/PDF_RENDERING_POC.md) — Hasil evaluasi Proof of Concept PDF Fase 0.
- [docs/IMPLEMENTATION_STATUS.md](docs/IMPLEMENTATION_STATUS.md) — Status aktual modul repository dan roadmap fitur.
- [docs/EXECUTION_PLAN.md](docs/EXECUTION_PLAN.md) — Rencana eksekusi dan kriteria penerimaan per fase.
- [docs/reports/PHASE-0-REPORT.md](docs/reports/PHASE-0-REPORT.md) — Laporan penyelesaian Fase 0 (Foundation & PoC).
- [docs/reports/PHASE-1-REPORT.md](docs/reports/PHASE-1-REPORT.md) — Laporan penyelesaian Fase 1 (App Shell & Project CRUD).
- [docs/reports/PHASE-2-REPORT.md](docs/reports/PHASE-2-REPORT.md) — Laporan penyelesaian Fase 2 (Visual Template Designer).
- [docs/reports/PHASE-3-REPORT.md](docs/reports/PHASE-3-REPORT.md) — Laporan penyelesaian Fase 3 (Recipient Management & 1000 Data Import).
- [docs/reports/PHASE-4-REPORT.md](docs/reports/PHASE-4-REPORT.md) — Laporan penyelesaian Fase 4 (Production PDF Rendering Pipeline).

---

## 📄 Lisensi & Privasi

Aplikasi ini bersifat **Privacy-First**. Seluruh data penerima sertifikat, file Excel/CSV, tanda tangan digital, dan file PDF yang dihasilkan berada 100% di perangkat lokal pengguna dan tidak pernah dikirimkan ke server eksternal mana pun.
