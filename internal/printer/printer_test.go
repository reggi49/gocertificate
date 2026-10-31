package printer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListPrinters(t *testing.T) {
	// Should not panic and return printers if any are configured
	printers, err := ListPrinters()
	if err != nil {
		t.Logf("ListPrinters returned error (acceptable on headless environments): %v", err)
	} else {
		t.Logf("Found %d printers", len(printers))
		for _, p := range printers {
			t.Logf("- %s (Default: %v)", p.Name, p.IsDefault)
		}
	}
}

func TestPrintPDF_MissingFile(t *testing.T) {
	nonExistent := filepath.Join(os.TempDir(), "non_existent_certificate.pdf")
	err := PrintPDF(nonExistent, "", 1)
	if err == nil {
		t.Fatalf("expected error for non-existent file, got nil")
	}
}

func TestBuildPowerShellEncodedCommand(t *testing.T) {
	script := "Write-Host 'Hello World'"
	encoded := buildPowerShellEncodedCommand(script)
	if encoded == "" {
		t.Fatalf("expected non-empty base64 string")
	}
	// PowerShell encodes "Write-Host 'Hello World'" in UTF-16LE + Base64
	// In UTF-16LE, 'W' is 0x57 0x00, 'r' is 0x72 0x00, etc.
	expectedPrefix := "VwByAGkAdABlAC0ASABvAHMAdA"
	if len(encoded) < len(expectedPrefix) || encoded[:len(expectedPrefix)] != expectedPrefix {
		t.Fatalf("unexpected encoded output: %s", encoded)
	}
}

