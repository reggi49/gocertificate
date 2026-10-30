package filesystem_test

import (
	"testing"

	"certificate-generator/internal/storage/filesystem"
)

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal name",
			input:    "Certificate_Budi.pdf",
			expected: "Certificate_Budi.pdf",
		},
		{
			name:     "Illegal characters in filename",
			input:    `Sertifikat: "Budi" <Pemenang> / 2026?.pdf`,
			expected: "Sertifikat_Budi_Pemenang_2026.pdf",
		},
		{
			name:     "Windows reserved name CON",
			input:    "CON.pdf",
			expected: "_CON.pdf",
		},
		{
			name:     "Windows reserved name aux without extension",
			input:    "aux",
			expected: "_aux",
		},
		{
			name:     "Trailing dots and spaces",
			input:    "document.pdf.  .",
			expected: "document.pdf",
		},
		{
			name:     "Empty name fallback",
			input:    "    ",
			expected: "unnamed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filesystem.SanitizeFilename(tt.input)
			if got != tt.expected {
				t.Errorf("SanitizeFilename(%q) = %q, expected %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestValidateID(t *testing.T) {
	validIDs := []string{"proj-123", "a1b2c3d4-e5f6-7890-abcd-ef1234567890", "PROJECT_2026"}
	for _, id := range validIDs {
		if err := filesystem.ValidateID(id); err != nil {
			t.Errorf("Expected valid id for %q, got error: %v", id, err)
		}
	}

	invalidIDs := []string{"", "../escaped", "proj/sub", "proj 123", "id;rm"}
	for _, id := range invalidIDs {
		if err := filesystem.ValidateID(id); err == nil {
			t.Errorf("Expected invalid id for %q, got nil error", id)
		}
	}
}
