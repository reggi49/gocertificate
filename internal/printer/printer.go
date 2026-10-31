package printer

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

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
	// Try Edge headless first if installed
	edgePath := os.Getenv("ProgramFiles(x86)") + `\Microsoft\Edge\Application\msedge.exe`
	if _, err := os.Stat(edgePath); os.IsNotExist(err) {
		edgePath = os.Getenv("ProgramFiles") + `\Microsoft\Edge\Application\msedge.exe`
	}

	if _, err := os.Stat(edgePath); err == nil {
		args := []string{"--headless", "--print-to-printer"}
		if printerName != "" {
			args = append(args, fmt.Sprintf("--print-to-printer-name=%s", printerName))
		}
		args = append(args, filePath)

		success := true
		for i := 0; i < copies; i++ {
			cmd := exec.Command(edgePath, args...)
			if err := cmd.Run(); err != nil {
				success = false
				break
			}
		}
		if success {
			return nil
		}
	}

	// Standard PowerShell Start-Process with Print or PrintTo verb
	for i := 0; i < copies; i++ {
		var psScript string
		if printerName != "" {
			psScript = fmt.Sprintf(`Start-Process -FilePath '%s' -Verb PrintTo -ArgumentList '"%s"' -PassThru | ForEach-Object { Start-Sleep -Milliseconds 1500; if (!$_.HasExited) { $_.CloseMainWindow(); Start-Sleep -Milliseconds 500; if (!$_.HasExited) { Stop-Process -Id $_.Id -Force } } }`, filePath, printerName)
		} else {
			psScript = fmt.Sprintf(`Start-Process -FilePath '%s' -Verb Print -PassThru | ForEach-Object { Start-Sleep -Milliseconds 1500; if (!$_.HasExited) { $_.CloseMainWindow(); Start-Sleep -Milliseconds 500; if (!$_.HasExited) { Stop-Process -Id $_.Id -Force } } }`, filePath)
		}
		cmd := exec.Command("powershell", "-NoProfile", "-Command", psScript)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("gagal mengirim berkas ke printer: %s (%w)", strings.TrimSpace(string(out)), err)
		}
	}
	return nil
}

func listPrintersUnix() ([]domain.PrinterInfo, error) {
	return listPrintersDarwin()
}

func printPDFUnix(filePath string, printerName string, copies int) error {
	return printPDFDarwin(filePath, printerName, copies)
}

