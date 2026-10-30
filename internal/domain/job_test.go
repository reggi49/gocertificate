package domain_test

import (
	"testing"

	"certificate-generator/internal/domain"
)

func TestJobStateMachineTransitions(t *testing.T) {
	// Valid flow 1: pending -> validating -> running -> completed
	if !domain.CanTransition(domain.JobStatusPending, domain.JobStatusValidating) {
		t.Errorf("Expected pending -> validating to be valid")
	}
	if !domain.CanTransition(domain.JobStatusValidating, domain.JobStatusRunning) {
		t.Errorf("Expected validating -> running to be valid")
	}
	if !domain.CanTransition(domain.JobStatusRunning, domain.JobStatusCompleted) {
		t.Errorf("Expected running -> completed to be valid")
	}

	// Valid flow 2: running -> completed_with_errors
	if !domain.CanTransition(domain.JobStatusRunning, domain.JobStatusCompletedWithErrors) {
		t.Errorf("Expected running -> completed_with_errors to be valid")
	}

	// Valid flow 3: running -> cancelling -> cancelled
	if !domain.CanTransition(domain.JobStatusRunning, domain.JobStatusCancelling) {
		t.Errorf("Expected running -> cancelling to be valid")
	}
	if !domain.CanTransition(domain.JobStatusCancelling, domain.JobStatusCancelled) {
		t.Errorf("Expected cancelling -> cancelled to be valid")
	}

	// Invalid flow: completed -> running (terminal state)
	if domain.CanTransition(domain.JobStatusCompleted, domain.JobStatusRunning) {
		t.Errorf("Expected completed -> running to be rejected")
	}

	// Invalid flow: pending -> completed (cannot skip running)
	if domain.CanTransition(domain.JobStatusPending, domain.JobStatusCompleted) {
		t.Errorf("Expected pending -> completed to be rejected")
	}

	// Self-transition is always allowed
	if !domain.CanTransition(domain.JobStatusRunning, domain.JobStatusRunning) {
		t.Errorf("Self-transition should be allowed")
	}
}

func TestBatchConfigValidation(t *testing.T) {
	cfg := &domain.BatchConfig{
		ProjectID:  "proj-1",
		TemplateID: "tmpl-1",
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("cfg.Validate failed: %v", err)
	}

	// Concurrency should default to 1
	if cfg.Concurrency != 1 {
		t.Errorf("Expected default concurrency 1, got %d", cfg.Concurrency)
	}

	// Concurrency capped at 4
	cfg.Concurrency = 10
	_ = cfg.Validate()
	if cfg.Concurrency != 4 {
		t.Errorf("Expected capped concurrency 4, got %d", cfg.Concurrency)
	}

	// Missing project id
	emptyProj := &domain.BatchConfig{TemplateID: "tmpl-1"}
	if err := emptyProj.Validate(); err == nil {
		t.Errorf("Expected error for missing project id, got nil")
	}
}
