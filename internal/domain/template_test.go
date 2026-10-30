package domain_test

import (
	"strings"
	"testing"

	"certificate-generator/internal/domain"
)

func TestTemplateValidation(t *testing.T) {
	// Empty name
	_, err := domain.ValidateTemplateName("   ")
	if err != domain.ErrEmptyTemplateName {
		t.Errorf("Expected ErrEmptyTemplateName, got %v", err)
	}

	// Name too long
	longName := strings.Repeat("T", 101)
	_, err = domain.ValidateTemplateName(longName)
	if err != domain.ErrTemplateNameTooLong {
		t.Errorf("Expected ErrTemplateNameTooLong, got %v", err)
	}

	// Valid name
	clean, err := domain.ValidateTemplateName("  Template Piagam Penghargaan  ")
	if err != nil {
		t.Fatalf("Unexpected error for valid name: %v", err)
	}
	if clean != "Template Piagam Penghargaan" {
		t.Errorf("Expected clean trimmed name, got %s", clean)
	}

	// Blank document defaults
	doc := domain.NewDefaultA4LandscapeDocument()
	if err := doc.Validate(); err != nil {
		t.Fatalf("Default blank document failed validation: %v", err)
	}
	if doc.Page.WidthMm != 297 || doc.Page.HeightMm != 210 || doc.Page.Orientation != "landscape" {
		t.Errorf("Unexpected default page properties: %+v", doc.Page)
	}
}
