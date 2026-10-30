package domain_test

import (
	"strings"
	"testing"

	"certificate-generator/internal/domain"
)

func TestProjectValidation(t *testing.T) {
	// Empty name
	_, err := domain.ValidateCreateProjectInput(domain.CreateProjectInput{Name: "   "})
	if err != domain.ErrEmptyProjectName {
		t.Errorf("Expected ErrEmptyProjectName, got %v", err)
	}

	// Name too long
	longName := strings.Repeat("A", 101)
	_, err = domain.ValidateCreateProjectInput(domain.CreateProjectInput{Name: longName})
	if err != domain.ErrProjectNameTooLong {
		t.Errorf("Expected ErrProjectNameTooLong, got %v", err)
	}

	// Control character
	_, err = domain.ValidateCreateProjectInput(domain.CreateProjectInput{Name: "Project\x07Name"})
	if err == nil {
		t.Errorf("Expected error for control characters in project name, got nil")
	}

	// Valid name and event details
	clean, err := domain.ValidateCreateProjectInput(domain.CreateProjectInput{
		Name:          "  Pelatihan AI 2026  ",
		Description:   " Sertifikat peserta ",
		EventDate:     " 2026-11-15 ",
		EventLocation: " Gedung Sasana Budaya ",
		EventAddress:  " Jl. Ganesha No. 10, Bandung ",
	})
	if err != nil {
		t.Fatalf("Unexpected error for valid input: %v", err)
	}
	if clean.Name != "Pelatihan AI 2026" {
		t.Errorf("Expected trimmed name 'Pelatihan AI 2026', got '%s'", clean.Name)
	}
	if clean.Description != "Sertifikat peserta" {
		t.Errorf("Expected trimmed description, got '%s'", clean.Description)
	}
	if clean.EventDate != "2026-11-15" {
		t.Errorf("Expected trimmed event date, got '%s'", clean.EventDate)
	}
	if clean.EventLocation != "Gedung Sasana Budaya" {
		t.Errorf("Expected trimmed event location, got '%s'", clean.EventLocation)
	}
	if clean.EventAddress != "Jl. Ganesha No. 10, Bandung" {
		t.Errorf("Expected trimmed event address, got '%s'", clean.EventAddress)
	}
}
