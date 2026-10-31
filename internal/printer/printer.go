package printer

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"unicode/utf16"

	"certificate-generator/internal/domain"
)

// ListPrinters returns the list of printers connected to the host system.
func ListPrinters() ([]domain.PrinterInfo, error) {
	switch runtime.GOOS {
	case "darwin":
		return listPrintersDarwin()
	case "windows":
		return listPrintersWindows()
	default:
		return listPrintersUnix()
	}
}

// PrintPDF sends a PDF document directly to the specified printer or default printer.
func PrintPDF(filePath string, printerName string, copies int) error {
	filePath = filepath.Clean(filePath)
	if _, err := os.Stat(filePath); err != nil {
		return fmt.Errorf("berkas PDF tidak ditemukan di '%s': file mungkin telah dipindahkan atau dihapus", filePath)
	}

	if copies <= 0 {
		copies = 1
	}

	switch runtime.GOOS {
	case "darwin":
		return printPDFDarwin(filePath, printerName, copies)
	case "windows":
		return printPDFWindows(filePath, printerName, copies)
	default:
		return printPDFUnix(filePath, printerName, copies)
	}
}

func listPrintersDarwin() ([]domain.PrinterInfo, error) {
	defaultPrinter := ""
	outDef, err := exec.Command("lpstat", "-d").Output()
	if err == nil {
		str := strings.TrimSpace(string(outDef))
		if idx := strings.Index(str, ":"); idx != -1 {
			defaultPrinter = strings.TrimSpace(str[idx+1:])
		}
	}

	out, err := exec.Command("lpstat", "-e").Output()
	if err != nil {
		if defaultPrinter != "" {
			return []domain.PrinterInfo{{Name: defaultPrinter, IsDefault: true}}, nil
		}
		return nil, fmt.Errorf("gagal membaca daftar printer: %w", err)
	}

	var list []domain.PrinterInfo
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		name := strings.TrimSpace(line)
		if name == "" {
			continue
		}
		isDef := (name == defaultPrinter)
		list = append(list, domain.PrinterInfo{
			Name:      name,
			IsDefault: isDef,
		})
	}

	if len(list) == 0 && defaultPrinter != "" {
		list = append(list, domain.PrinterInfo{Name: defaultPrinter, IsDefault: true})
	}

	return list, nil
}

func printPDFDarwin(filePath string, printerName string, copies int) error {
	args := []string{"-n", strconv.Itoa(copies), "-o", "fit-to-page"}
	if printerName != "" {
		args = append(args, "-d", printerName)
	}
	args = append(args, filePath)

	cmd := exec.Command("lp", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gagal mencetak dokumen: %s (%w)", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func listPrintersWindows() ([]domain.PrinterInfo, error) {
	psCmd := `Get-CimInstance Win32_Printer | Select-Object -Property Name, Default | ConvertTo-Json`
	cmd := exec.Command("powershell", "-NoProfile", "-Command", psCmd)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan daftar printer dari Windows: %w", err)
	}

	type psPrinter struct {
		Name    string `json:"Name"`
		Default bool   `json:"Default"`
	}

	raw := strings.TrimSpace(string(out))
	if strings.HasPrefix(raw, "[") {
		var items []psPrinter
		if err := json.Unmarshal([]byte(raw), &items); err == nil {
			var res []domain.PrinterInfo
			for _, it := range items {
				res = append(res, domain.PrinterInfo{
					Name:      it.Name,
					IsDefault: it.Default,
				})
			}
			return res, nil
		}
	} else if strings.HasPrefix(raw, "{") {
		var it psPrinter
		if err := json.Unmarshal([]byte(raw), &it); err == nil {
			return []domain.PrinterInfo{{Name: it.Name, IsDefault: it.Default}}, nil
		}
	}

	return nil, fmt.Errorf("tidak ada printer terdeteksi di Windows")
}

func printPDFWindows(filePath string, printerName string, copies int) error {
	absPath, err := filepath.Abs(filePath)
	if err == nil {
		filePath = absPath
	}
	filePath = filepath.Clean(filePath)

	if copies <= 0 {
		copies = 1
	}

	// 1. Check for dedicated silent printing utilities (SumatraPDF)
	sumatraPaths := []string{
		os.Getenv("ProgramFiles") + `\SumatraPDF\SumatraPDF.exe`,
		os.Getenv("ProgramFiles(x86)") + `\SumatraPDF\SumatraPDF.exe`,
		os.Getenv("LocalAppData") + `\SumatraPDF\SumatraPDF.exe`,
	}
	for _, sp := range sumatraPaths {
		if _, err := os.Stat(sp); err == nil {
			args := []string{"-silent", "-exit-on-print"}
			if printerName != "" {
				args = append(args, "-print-to", printerName)
			} else {
				args = append(args, "-print-to-default")
			}
			args = append(args, filePath)
			for i := 0; i < copies; i++ {
				if err := exec.Command(sp, args...).Run(); err != nil {
					break
				}
			}
			return nil
		}
	}

	// 2. Check for Acrobat Reader (AcroRd32 / Acrobat)
	acroPaths := []string{
		`C:\Program Files\Adobe\Acrobat DC\Acrobat\Acrobat.exe`,
		`C:\Program Files (x86)\Adobe\Acrobat Reader DC\Reader\AcroRd32.exe`,
		`C:\Program Files\Adobe\Acrobat Reader DC\Reader\AcroRd32.exe`,
	}
	for _, ap := range acroPaths {
		if _, err := os.Stat(ap); err == nil {
			args := []string{"/n", "/s", "/h"}
			if printerName != "" {
				args = append(args, "/t", filePath, printerName)
			} else {
				args = append(args, "/p", filePath)
			}
			for i := 0; i < copies; i++ {
				_ = exec.Command(ap, args...).Run()
			}
			return nil
		}
	}

	// 3. Native Windows 10 & 11 WinRT + .NET PrintDocument Engine (built into 100% of modern Windows)
	escapedFilePath := strings.ReplaceAll(filePath, "'", "''")
	escapedPrinter := strings.ReplaceAll(printerName, "'", "''")

	psWinRTScript := fmt.Sprintf(`
$ErrorActionPreference = 'Stop'
Add-Type -AssemblyName System.Drawing
Add-Type -AssemblyName System.Windows.Forms

[Windows.Data.Pdf.PdfDocument, Windows.Data.Pdf, ContentType = WindowsRuntime] > $null
[Windows.Storage.StorageFile, Windows.Storage, ContentType = WindowsRuntime] > $null
[Windows.Storage.Streams.InMemoryRandomAccessStream, Windows.Storage.Streams, ContentType = WindowsRuntime] > $null

$filePath = '%s'
$targetPrinter = '%s'
$copies = %d

if (!(Test-Path -LiteralPath $filePath)) {
    throw "Berkas tidak ditemukan: $filePath"
}

$storageFile = [Windows.Storage.StorageFile]::GetFileFromPathAsync($filePath).GetAwaiter().GetResult()
$pdfDoc = [Windows.Data.Pdf.PdfDocument]::LoadFromFileAsync($storageFile).GetAwaiter().GetResult()

$pageIndex = 0
$totalPages = $pdfDoc.PageCount
if ($totalPages -le 0) {
    throw "Dokumen PDF tidak memuat halaman"
}

$printDoc = New-Object System.Drawing.Printing.PrintDocument
$printDoc.DocumentName = [System.IO.Path]::GetFileNameWithoutExtension($filePath)

if ($targetPrinter -ne '') {
    $printDoc.PrinterSettings.PrinterName = $targetPrinter
}

if ($copies -gt 1) {
    $printDoc.PrinterSettings.Copies = [int]$copies
}

# Auto-detect orientation from first page
$firstPage = $pdfDoc.GetPage(0)
$firstSize = $firstPage.Size
$firstPage.Dispose()

if ($firstSize.Width -gt $firstSize.Height) {
    $printDoc.DefaultPageSettings.Landscape = $true
} else {
    $printDoc.DefaultPageSettings.Landscape = $false
}

$printDoc.OriginAtMargins = $false

$printDoc.add_PrintPage({
    param($sender, $e)
    if ($pageIndex -lt $totalPages) {
        $page = $pdfDoc.GetPage($pageIndex)
        $stream = New-Object Windows.Storage.Streams.InMemoryRandomAccessStream
        $page.RenderToStreamAsync($stream).GetAwaiter().GetResult()
        $netStream = $stream.AsStreamForRead()
        $bitmap = [System.Drawing.Image]::FromStream($netStream)

        # Scale to paper bounds maintaining aspect ratio
        $destWidth = $e.PageBounds.Width
        $destHeight = $e.PageBounds.Height
        $scaleX = $destWidth / $bitmap.Width
        $scaleY = $destHeight / $bitmap.Height
        $scale = [Math]::Min($scaleX, $scaleY)
        $drawWidth = $bitmap.Width * $scale
        $drawHeight = $bitmap.Height * $scale
        $drawX = ($destWidth - $drawWidth) / 2
        $drawY = ($destHeight - $drawHeight) / 2

        $e.Graphics.DrawImage($bitmap, [float]$drawX, [float]$drawY, [float]$drawWidth, [float]$drawHeight)

        $bitmap.Dispose()
        $netStream.Dispose()
        $stream.Dispose()
        $page.Dispose()

        $pageIndex++
        if ($pageIndex -lt $totalPages) {
            $e.HasMorePages = $true
        } else {
            $e.HasMorePages = $false
            $pageIndex = 0
        }
    }
})

$printDoc.Print()
$printDoc.Dispose()
`, escapedFilePath, escapedPrinter, copies)

	encodedCmd := buildPowerShellEncodedCommand(psWinRTScript)
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-EncodedCommand", encodedCmd)
	out, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}

	// 4. Fallback: ShellExecute Print verb (without prematurely killing the process)
	outStr := strings.TrimSpace(string(out))
	psFallback := fmt.Sprintf(`Start-Process -FilePath '%s' -Verb Print`, escapedFilePath)
	fallbackCmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", psFallback)
	if fbOut, fbErr := fallbackCmd.CombinedOutput(); fbErr != nil {
		return fmt.Errorf("gagal mencetak dokumen ke printer di Windows: %s (fallback: %s)", outStr, strings.TrimSpace(string(fbOut)))
	}

	return nil
}

func buildPowerShellEncodedCommand(script string) string {
	utf16Units := utf16.Encode([]rune(script))
	rawBytes := make([]byte, len(utf16Units)*2)
	for i, u := range utf16Units {
		binary.LittleEndian.PutUint16(rawBytes[i*2:], u)
	}
	return base64.StdEncoding.EncodeToString(rawBytes)
}

func listPrintersUnix() ([]domain.PrinterInfo, error) {
	return listPrintersDarwin()
}

func printPDFUnix(filePath string, printerName string, copies int) error {
	return printPDFDarwin(filePath, printerName, copies)
}

