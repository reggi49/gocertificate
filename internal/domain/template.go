package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrTemplateNotFound    = errors.New("template not found")
	ErrEmptyTemplateName   = errors.New("template name cannot be empty")
	ErrTemplateNameTooLong = errors.New("template name cannot exceed 100 characters")
	ErrInvalidTemplateID   = errors.New("invalid template id")
)

// Template represents a certificate template design associated with a project.
type Template struct {
	ID            string    `json:"id"`
	ProjectID     string    `json:"projectId"`
	Name          string    `json:"name"`
	DocumentJSON  string    `json:"documentJson"`
	SchemaVersion int       `json:"schemaVersion"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// CreateTemplateInput holds input parameters to create a new template.
type CreateTemplateInput struct {
	ProjectID    string `json:"projectId"`
	Name         string `json:"name"`
	DocumentJSON string `json:"documentJson"`
}

// UpdateTemplateInput holds input parameters to update an existing template.
type UpdateTemplateInput struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	DocumentJSON string `json:"documentJson"`
}

// ValidateTemplateName ensures template name conforms to rules.
func ValidateTemplateName(name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", ErrEmptyTemplateName
	}
	if len(trimmed) > 100 {
		return "", ErrTemplateNameTooLong
	}
	for _, r := range trimmed {
		if r < 32 {
			return "", fmt.Errorf("template name contains invalid control characters")
		}
	}
	return trimmed, nil
}

// NewDefaultA4LandscapeDocument creates a canonical blank A4 landscape certificate document.
func NewDefaultA4LandscapeDocument() *Document {
	return &Document{
		SchemaVersion: 1,
		Page: PageSettings{
			WidthMm:         297,
			HeightMm:        210,
			Orientation:     "landscape",
			BackgroundColor: "#FFFFFF",
		},
		Elements: []Element{},
	}
}
