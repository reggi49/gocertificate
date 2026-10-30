package domain_test

import (
	"testing"

	"certificate-generator/internal/domain"
)

func TestValidateRecipient(t *testing.T) {
	// Empty full name should fail
	r1 := &domain.Recipient{FullName: "   "}
	if err := domain.ValidateRecipient(r1); err != domain.ErrEmptyRecipientFullName {
		t.Errorf("Expected ErrEmptyRecipientFullName, got %v", err)
	}

	// Invalid email should fail
	r2 := &domain.Recipient{
		FullName: "Budi Santoso",
		Email:    "invalid-email-without-at",
	}
	if err := domain.ValidateRecipient(r2); err != domain.ErrInvalidEmail {
		t.Errorf("Expected ErrInvalidEmail, got %v", err)
	}

	// Valid recipient
	r3 := &domain.Recipient{
		FullName:      "  Dr. Siti Aminah, M.Kom.  ",
		Email:         "siti.aminah@univ.ac.id",
		CertificateID: "CERT-001",
	}
	if err := domain.ValidateRecipient(r3); err != nil {
		t.Fatalf("Unexpected error for valid recipient: %v", err)
	}
	if r3.FullName != "Dr. Siti Aminah, M.Kom." {
		t.Errorf("Expected trimmed full name, got %s", r3.FullName)
	}
}

func TestNormalizeAndValidateDate(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		isErr    bool
	}{
		{"2026-10-30", "2026-10-30", false},
		{"30/10/2026", "2026-10-30", false},
		{"30-10-2026", "2026-10-30", false},
		{"30 Oktober 2026", "2026-10-30", false},
		{"2 Mei 2026", "2026-05-02", false},
		{"   ", "", false}, // empty is allowed (optional field)
		{"invalid-date-xyz", "", true},
	}

	for _, tt := range tests {
		normalized, err := domain.NormalizeAndValidateDate(tt.input)
		if tt.isErr && err == nil {
			t.Errorf("Expected error for %q, got nil", tt.input)
		}
		if !tt.isErr && err != nil {
			t.Errorf("Unexpected error for %q: %v", tt.input, err)
		}
		if normalized != tt.expected {
			t.Errorf("NormalizeAndValidateDate(%q) = %q, expected %q", tt.input, normalized, tt.expected)
		}
	}
}
