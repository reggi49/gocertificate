package domain

import (
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"
)

var (
	ErrRecipientNotFound      = errors.New("recipient not found")
	ErrEmptyRecipientFullName = errors.New("recipient full name cannot be empty")
	ErrInvalidEmail           = errors.New("invalid email address format")
	ErrDuplicateCertID        = errors.New("certificate ID already exists in this project")
	ErrMissingFullNameMapping = errors.New("column mapping for full name is required")
)

// ConflictStrategy defines re-import duplicate resolution behavior.
type ConflictStrategy string

const (
	ConflictStrategySkip   ConflictStrategy = "skip"   // Skip duplicate rows
	ConflictStrategyUpdate ConflictStrategy = "update" // Update existing recipient
	ConflictStrategyFail   ConflictStrategy = "fail"   // Abort transaction on duplicate
)

// Recipient represents an individual certificate recipient record.
type Recipient struct {
	ID               string    `json:"id"`
	ProjectID        string    `json:"projectId"`
	FullName         string    `json:"fullName"`
	Email            string    `json:"email,omitempty"`
	Organization     string    `json:"organization,omitempty"`
	EventName        string    `json:"eventName,omitempty"`
	IssueDate        string    `json:"issueDate,omitempty"`
	CertificateID    string    `json:"certificateId,omitempty"`
	CustomFieldsJSON string    `json:"customFieldsJson,omitempty"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// ColumnMapping defines how spreadsheet columns map to recipient fields.
type ColumnMapping struct {
	FullNameColumn      string            `json:"fullNameColumn"`
	EmailColumn         string            `json:"emailColumn,omitempty"`
	OrganizationColumn  string            `json:"organizationColumn,omitempty"`
	EventNameColumn     string            `json:"eventNameColumn,omitempty"`
	IssueDateColumn     string            `json:"issueDateColumn,omitempty"`
	CertificateIDColumn string            `json:"certificateIdColumn,omitempty"`
	CustomFieldColumns  map[string]string `json:"customFieldColumns,omitempty"` // field_name -> source_column
}

// ImportRowError details an issue in a specific row.
type ImportRowError struct {
	RowNumber int    `json:"rowNumber"`
	Column    string `json:"column"`
	Category  string `json:"category"` // e.g. "MISSING_REQUIRED", "INVALID_DATE", "DUPLICATE_ID", "INVALID_EMAIL"
	Message   string `json:"message"`
	RawValue  string `json:"rawValue"`
}

// ImportSummary contains results of an import operation.
type ImportSummary struct {
	TotalRows    int              `json:"totalRows"`
	ValidRows    int              `json:"validRows"`
	InvalidRows  int              `json:"invalidRows"`
	ImportedRows int              `json:"importedRows"`
	SkippedRows  int              `json:"skippedRows"`
	UpdatedRows  int              `json:"updatedRows"`
	Errors       []ImportRowError `json:"errors"`
}

// RecipientFilter holds search and pagination criteria.
type RecipientFilter struct {
	ProjectID string `json:"projectId"`
	Search    string `json:"search,omitempty"`
	Page      int    `json:"page"`
	PageSize  int    `json:"pageSize"`
}

// PaginatedRecipients wraps paginated results.
type PaginatedRecipients struct {
	Items      []Recipient `json:"items"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"pageSize"`
	TotalPages int         `json:"totalPages"`
}

// ValidateRecipient validates mandatory fields and formats.
func ValidateRecipient(r *Recipient) error {
	r.FullName = strings.TrimSpace(r.FullName)
	if r.FullName == "" {
		return ErrEmptyRecipientFullName
	}

	r.Email = strings.TrimSpace(r.Email)
	if r.Email != "" {
		if err := ValidateEmailString(r.Email); err != nil {
			return err
		}
	}

	r.CertificateID = strings.TrimSpace(r.CertificateID)
	r.Organization = strings.TrimSpace(r.Organization)
	r.EventName = strings.TrimSpace(r.EventName)
	r.IssueDate = strings.TrimSpace(r.IssueDate)

	return nil
}

// ValidateEmailString checks email format if present.
func ValidateEmailString(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return nil
	}
	_, err := mail.ParseAddress(email)
	if err != nil || !strings.Contains(email, ".") {
		return ErrInvalidEmail
	}
	return nil
}

// Common date formats supported across Indonesian and International conventions
var supportedDateFormats = []string{
	"2006-01-02",
	"02/01/2006",
	"2/1/2006",
	"02-01-2006",
	"2-1-2006",
	"2006/01/02",
	"02 Jan 2006",
	"2 Jan 2006",
	"02 January 2006",
	"2 January 2006",
}

// Indonesian month mapping for date normalization
var indonesianMonths = map[string]string{
	"januari": "January", "februari": "February", "maret": "March",
	"april": "April", "mei": "May", "juni": "June",
	"juli": "July", "agustus": "August", "september": "September",
	"oktober": "October", "november": "November", "desember": "December",
	"agu": "Aug", "okt": "Oct", "des": "Dec",
}

var wordRegex = regexp.MustCompile(`(?i)\b([a-z]+)\b`)

// NormalizeAndValidateDate attempts to parse date strings and returns normalized format or error.
func NormalizeAndValidateDate(rawDate string) (string, error) {
	clean := strings.TrimSpace(rawDate)
	if clean == "" {
		return "", nil
	}

	// Translate Indonesian month names to English for time.Parse
	translated := wordRegex.ReplaceAllStringFunc(clean, func(m string) string {
		lower := strings.ToLower(m)
		if en, ok := indonesianMonths[lower]; ok {
			return en
		}
		return m
	})

	for _, layout := range supportedDateFormats {
		t, err := time.Parse(layout, translated)
		if err == nil {
			return t.Format("2006-01-02"), nil
		}
	}

	// If no standard numeric format matched, return error
	return "", fmt.Errorf("invalid date format '%s'", rawDate)
}
