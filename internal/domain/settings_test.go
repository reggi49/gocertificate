package domain_test

import (
	"testing"

	"certificate-generator/internal/domain"
)

func TestSettingsValidation(t *testing.T) {
	s := domain.DefaultSettings()
	if err := domain.ValidateSettings(&s); err != nil {
		t.Fatalf("Default settings should be valid: %v", err)
	}

	invalidTheme := domain.AppSettings{Theme: "neon-blue"}
	if err := domain.ValidateSettings(&invalidTheme); err == nil {
		t.Errorf("Expected error for invalid theme, got nil")
	}

	invalidLang := domain.AppSettings{Language: "de"}
	if err := domain.ValidateSettings(&invalidLang); err == nil {
		t.Errorf("Expected error for invalid language, got nil")
	}

	// AutoSave bounds check
	clamped := domain.AppSettings{AutoSaveIntervalSec: 500}
	if err := domain.ValidateSettings(&clamped); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if clamped.AutoSaveIntervalSec != 300 {
		t.Errorf("Expected clamped interval 300, got %d", clamped.AutoSaveIntervalSec)
	}
}
