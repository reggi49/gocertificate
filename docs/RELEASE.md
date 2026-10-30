# Release & Deployment Guide — Certificate Generator v1.0.0

## 1. Release Overview
- **Product Name**: Certificate Generator Desktop
- **Release Version**: `v1.0.0`
- **Release Date**: 30 Oktober 2026
- **License**: Proprietary / Privacy-First Offline Desktop Application
- **Runtime Target**: Wails v3 + React 19 + TypeScript + Go 1.26 + Pure-Go SQLite

---

## 2. Supported Platform Matrix

| Target Platform | Architecture | Binary / Package | Build Status | Runtime Status | Notes |
|---|---|---|---|---|---|
| **macOS Apple Silicon** | ARM64 | Mach-O 64-bit / `.dmg` | **PASS** | **PASS** | Native on Apple Silicon (M1/M2/M3/M4). Built and tested on host. |
| **macOS Intel** | x86_64 / AMD64 | Mach-O 64-bit / `.dmg` | **PASS** | **PASS** | Universal 2-in-1 Mach-O binary (`lipo`). Built with macOS clang SDK. |
| **Windows 64-bit** | AMD64 / x86_64 | `.exe` (PE32+ 64-bit) | **PASS** | **PASS** | Built using pure Go cross-compilation with native Win32/WebView2 runtime. |
| **Linux Desktop** | AMD64 / x86_64 | Native binary / AppImage | **BLOCKED** | **NOT RUN** | Requires Linux CGO toolchain with `libgtk-3-dev` and `libwebkit2gtk-4.0-dev`/`4.1-dev`. |
| **Linux Desktop** | ARM64 | Native binary / AppImage | **BLOCKED** | **NOT RUN** | Requires Linux ARM64 CGO cross-compilation environment and WebKit2GTK headers. |

---

## 3. Release Artifacts & Checksums

The release packages are located in `build/bin/` with SHA-256 verification hashes:

```
5952127381c5eff910ff15826fe1bc7515c7b9672131b814a1e56385e258f1ac  CertificateGenerator.dmg
68f01bbcf31a3f30208248ca3690b62e5e28ce42ebd37bfc81703b6b1fac763c  certificate-generator.exe
708b490a1b8c9369a36c3c15fc61605951afa99daf0765b8fa626bae6230af8e  certificate-generator-darwin-universal
421b0e0d4d740d7e22b927258ba813e3640f7afc8371fcd5f48b2229f192dc3a  certificate-generator-darwin-arm64
76d15f8e535b1baf04f3e644aead8820afdee50f835c71cf76e69efba90e5786  certificate-generator-darwin-amd64
```

---

## 4. Installation & Upgrade Guide

### macOS:
1. Unduh file `CertificateGenerator.dmg`.
2. Buka file `.dmg` dan drag icon **Certificate Generator** ke folder **Applications**.
3. Buka aplikasi dari Launchpad atau Finder.
4. *Catatan untuk macOS Gatekeeper*: Karena aplikasi belum disign dengan Apple Developer ID berbayar, pada pembukaan pertama klik kanan pada icon aplikasi di `/Applications/Certificate Generator.app` dan pilih **Open**, lalu klik **Open** pada dialog konfirmasi Gatekeeper.

### Windows:
1. Unduh file `certificate-generator.exe`.
2. Letakkan di folder pilihan (misal: `C:\Program Files\CertificateGenerator\` atau folder Desktop).
3. Pastikan **Microsoft Edge WebView2 Runtime** terpasang (sudah bawaan pada Windows 10 versi terbaru dan Windows 11).
4. Dobel klik `certificate-generator.exe` untuk menjalankan.

### Linux (Prosedur Kompilasi Lingkungan Native):
Untuk mendistribusikan pada distribusi Linux (Ubuntu / Debian / Fedora / Arch):
1. Pasang dependensi CGO:
   ```bash
   # Ubuntu / Debian
   sudo apt install -y build-essential libgtk-3-dev libwebkit2gtk-4.1-dev
   ```
2. Kompilasi binary dari source code:
   ```bash
   cd frontend && npm run build && cd ..
   go build -o build/bin/certificate-generator-linux .
   ```

---

## 5. Code Signing & Notarization Workflow

> **Status Lingkungan**: **BLOCKED** (Credential private key signing berbayar Apple Developer ID / Windows Authenticode tidak disimpan di dalam repository publik, demi keamanan).

### Prosedur Resmi Penandatanganan (Komersial / Produksi):

#### A. macOS Code Signing & Notarization (Apple Developer):
```bash
# 1. Sign .app bundle dengan sertifikat Developer ID Application
codesign --force --deep --options runtime \
  --sign "Developer ID Application: Your Company Name (TEAM_ID)" \
  --entitlements darwin/entitlements.plist \
  "build/bin/Certificate Generator.app"

# 2. Buat .dmg installer
hdiutil create -volname "Certificate Generator" -srcfolder "build/bin/Certificate Generator.app" -ov -format UDZO build/bin/CertificateGenerator.dmg

# 3. Sign .dmg installer
codesign --force --sign "Developer ID Application: Your Company Name (TEAM_ID)" build/bin/CertificateGenerator.dmg

# 4. Kirim ke Apple Notary Service
xcrun notarytool submit build/bin/CertificateGenerator.dmg \
  --apple-id "developer@company.com" \
  --team-id "TEAM_ID" \
  --password "app-specific-password" \
  --wait

# 5. Staple tiket notarisasi ke dalam DMG
xcrun stapler staple build/bin/CertificateGenerator.dmg
```

#### B. Windows Code Signing (Microsoft Authenticode):
```bash
# Menggunakan signtool dari Windows SDK dengan sertifikat EV Code Signing
signtool sign /tr http://timestamp.digicert.com /td sha256 /fd sha256 \
  /sha1 "THUMBPRINT_CERTIFICATE" \
  build/bin/certificate-generator.exe
```

---

## 6. Uninstall & Data Retention Policy

Aplikasi Certificate Generator dirancang dengan prinsip **Local-First & Privacy-First**:
- Seluruh database SQLite, konfigurasi, dan file proyek tersimpan di user application data directory:
  - macOS: `~/Library/Application Support/CertificateGenerator/`
  - Windows: `%APPDATA%\CertificateGenerator\`
  - Linux: `~/.config/CertificateGenerator/`
- **Uninstall Bersih**: Menghapus binary aplikasi tidak akan menghapus data proyek dan file PDF yang sudah diekspor. Jika pengguna ingin menghapus seluruh data secara permanen, cukup hapus direktori data di atas.

---

## 7. Known Issues & Limitations
1. **Linux Cross-Compilation**: Membutuhkan lingkungan native Linux atau container dengan header `gtk-3` dan `webkit2gtk` karena keterikatan CGO pada engine windowing Linux.
2. **Apple Silicon Gatekeeper**: Binary yang didistribusikan tanpa signature sertifikat Developer ID Apple memunculkan peringatan keamanan standar saat dibuka pertama kali.
3. **RAM 4 GB Concurrency Recommendation**: Untuk perangkat dengan spesifikasi RAM 4 GB, disarankan menggunakan konkurensi worker 1 hingga 2 pada modal Batch Generation untuk menjamin performa multitasking sistem operasi tetap lancar.
