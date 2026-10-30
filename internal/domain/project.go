package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrProjectNotFound      = errors.New("project not found")
	ErrDuplicateProjectName = errors.New("a project with this name already exists")
	ErrInvalidProjectID     = errors.New("invalid project id")
	ErrEmptyProjectName     = errors.New("project name cannot be empty")
	ErrProjectNameTooLong   = errors.New("project name cannot exceed 100 characters")
)

// Project represents the core metadata entity of a certificate project.
type Project struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	EventDate     string    `json:"eventDate"`
	EventLocation string    `json:"eventLocation"`
	EventAddress  string    `json:"eventAddress"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	IsArchived    bool      `json:"isArchived"`
}

// ProjectManifest represents the portable project.json file persisted in the project directory.
type ProjectManifest struct {
	SchemaVersion int    `json:"schemaVersion"`
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	EventDate     string `json:"eventDate"`
	EventLocation string `json:"eventLocation"`
	EventAddress  string `json:"eventAddress"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

// CreateProjectInput holds validated input for creating a new project.
type CreateProjectInput struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	EventDate     string `json:"eventDate"`
	EventLocation string `json:"eventLocation"`
	EventAddress  string `json:"eventAddress"`
}

// UpdateProjectInput holds validated input for updating a project.
type UpdateProjectInput struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	EventDate     string `json:"eventDate"`
	EventLocation string `json:"eventLocation"`
	EventAddress  string `json:"eventAddress"`
}

// DuplicateProjectInput holds input for cloning an existing project.
type DuplicateProjectInput struct {
	SourceProjectID string `json:"sourceProjectId"`
	NewName         string `json:"newName"`
}

// ValidateProjectName checks that project name conforms to safety and length rules.
func ValidateProjectName(name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", ErrEmptyProjectName
	}
	if len(trimmed) > 100 {
		return "", ErrProjectNameTooLong
	}
	// Check for prohibited control characters
	for _, r := range trimmed {
		if r < 32 {
			return "", fmt.Errorf("project name contains invalid control characters")
		}
	}
	return trimmed, nil
}

// ValidateCreateProjectInput validates input before creation.
func ValidateCreateProjectInput(input CreateProjectInput) (CreateProjectInput, error) {
	cleanName, err := ValidateProjectName(input.Name)
	if err != nil {
		return input, err
	}
	return CreateProjectInput{
		Name:          cleanName,
		Description:   strings.TrimSpace(input.Description),
		EventDate:     strings.TrimSpace(input.EventDate),
		EventLocation: strings.TrimSpace(input.EventLocation),
		EventAddress:  strings.TrimSpace(input.EventAddress),
	}, nil
}

// ValidateUpdateProjectInput validates input before update.
func ValidateUpdateProjectInput(input UpdateProjectInput) (UpdateProjectInput, error) {
	cleanName, err := ValidateProjectName(input.Name)
	if err != nil {
		return input, err
	}
	return UpdateProjectInput{
		Name:          cleanName,
		Description:   strings.TrimSpace(input.Description),
		EventDate:     strings.TrimSpace(input.EventDate),
		EventLocation: strings.TrimSpace(input.EventLocation),
		EventAddress:  strings.TrimSpace(input.EventAddress),
	}, nil
}
