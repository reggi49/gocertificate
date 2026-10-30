package domain

import (
	"errors"
	"fmt"
	"runtime"
	"time"
)

// Generation Job Status Constants
const (
	JobStatusPending             = "pending"
	JobStatusValidating          = "validating"
	JobStatusRunning             = "running"
	JobStatusCompleted           = "completed"
	JobStatusCompletedWithErrors = "completed_with_errors"
	JobStatusCancelling          = "cancelling"
	JobStatusCancelled           = "cancelled"
	JobStatusFailed              = "failed"
)

// Job Item Status Constants
const (
	ItemStatusPending   = "pending"
	ItemStatusCompleted = "completed"
	ItemStatusFailed    = "failed"
	ItemStatusSkipped   = "skipped"
)

var (
	ErrJobNotFound             = errors.New("generation job not found")
	ErrInvalidStateTransition  = errors.New("invalid job state transition")
	ErrJobAlreadyFinished      = errors.New("generation job has already finished")
	ErrEmptyRecipientsForBatch = errors.New("cannot start batch generation without recipients")
)

// AllowedTransitions maps each state to its permissible target states.
var AllowedTransitions = map[string]map[string]bool{
	JobStatusPending: {
		JobStatusValidating: true,
		JobStatusFailed:     true,
		JobStatusCancelled:  true,
	},
	JobStatusValidating: {
		JobStatusRunning:   true,
		JobStatusFailed:    true,
		JobStatusCancelled: true,
	},
	JobStatusRunning: {
		JobStatusCompleted:           true,
		JobStatusCompletedWithErrors: true,
		JobStatusCancelling:          true,
		JobStatusCancelled:           true,
		JobStatusFailed:              true,
	},
	JobStatusCancelling: {
		JobStatusCancelled: true,
		JobStatusFailed:    true,
	},
	// Terminal states have no outward transitions (except retry creating a new job or re-running failed items)
	JobStatusCompleted:           {},
	JobStatusCompletedWithErrors: {},
	JobStatusCancelled:           {},
	JobStatusFailed:              {},
}

// CanTransition checks if moving from currentState to targetState is valid.
func CanTransition(current, target string) bool {
	if current == target {
		return true
	}
	allowed, ok := AllowedTransitions[current]
	if !ok {
		return false
	}
	return allowed[target]
}

// GenerationJob represents an ongoing or completed certificate batch production job.
type GenerationJob struct {
	ID             string    `json:"id"`
	ProjectID      string    `json:"projectId"`
	TemplateID     string    `json:"templateId"`
	Status         string    `json:"status"`
	TotalCount     int       `json:"totalCount"`
	CompletedCount int       `json:"completedCount"`
	FailedCount    int       `json:"failedCount"`
	OutputPath     string    `json:"outputPath,omitempty"`
	ErrorMessage   string    `json:"errorMessage,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// GenerationJobItem tracks the rendering result of one recipient in a job.
type GenerationJobItem struct {
	ID             string    `json:"id"`
	JobID          string    `json:"jobId"`
	RecipientID    string    `json:"recipientId"`
	RecipientName  string    `json:"recipientName"`
	CertificateID  string    `json:"certificateId,omitempty"`
	OutputFileName string    `json:"outputFileName,omitempty"`
	Status         string    `json:"status"`
	ErrorMessage   string    `json:"errorMessage,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// BatchConfig holds configuration parameters for executing a batch.
type BatchConfig struct {
	ProjectID           string   `json:"projectId"`
	TemplateID          string   `json:"templateId"`
	RecipientIDs        []string `json:"recipientIds,omitempty"` // empty means all recipients in project
	Concurrency         int      `json:"concurrency"`            // default 1, maximum 4
	OutputDir           string   `json:"outputDir,omitempty"`    // optional custom output directory
	ExportZip           bool     `json:"exportZip"`              // bundle into zip file
	ExportIndividualPDF bool     `json:"exportIndividualPdf"`    // keep individual files
}

// Validate validates and applies defaults to BatchConfig.
func (c *BatchConfig) Validate() error {
	if c.ProjectID == "" {
		return fmt.Errorf("project id is required")
	}
	// TemplateID is optional; when omitted, BatchGenerator falls back to project's default template
	if c.Concurrency <= 0 {
		cpus := runtime.NumCPU()
		calculated := int(float64(cpus) * 0.75)
		if calculated < 1 {
			calculated = 1
		}
		c.Concurrency = calculated
	}
	return nil
}

// ProgressEvent represents real-time progress dispatched to UI / Wails event bus.
type ProgressEvent struct {
	JobID                string `json:"jobId"`
	CurrentItem          int    `json:"currentItem"`
	Total                int    `json:"total"`
	Completed            int    `json:"completed"`
	Failed               int    `json:"failed"`
	Percentage           int    `json:"percentage"`
	Status               string `json:"status"`
	CurrentRecipientName string `json:"currentRecipientName,omitempty"`
	Message              string `json:"message,omitempty"`
}
