package sqlite_test

import (
	"context"
	"testing"
	"time"

	"certificate-generator/internal/domain"
	"certificate-generator/internal/repository/sqlite"
)

func TestRecipientRepositoryCRUDAndBatch(t *testing.T) {
	pRepo, _, db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().UTC()

	// 1. Create parent project
	proj := &domain.Project{
		ID:         "proj-rec-test-1",
		Name:       "Proyek Recipient Test",
		CreatedAt:  now,
		UpdatedAt:  now,
		IsArchived: false,
	}
	if err := pRepo.Create(ctx, proj); err != nil {
		t.Fatalf("pRepo.Create failed: %v", err)
	}

	rRepo := sqlite.NewRecipientRepository(db)

	// 2. Create single recipient
	rec1 := &domain.Recipient{
		ID:            "rec-1",
		ProjectID:     proj.ID,
		FullName:      "Budi Santoso, S.T.",
		Email:         "budi@mail.com",
		Organization:  "PT Maju Jaya",
		EventName:     "Seminar AI",
		IssueDate:     "2026-10-30",
		CertificateID: "CERT-001",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := rRepo.Create(ctx, rec1); err != nil {
		t.Fatalf("rRepo.Create failed: %v", err)
	}

	// 3. GetByID
	fetched, err := rRepo.GetByID(ctx, rec1.ID)
	if err != nil {
		t.Fatalf("rRepo.GetByID failed: %v", err)
	}
	if fetched.FullName != rec1.FullName || fetched.Email != rec1.Email {
		t.Errorf("Fetched recipient mismatch: %+v vs %+v", fetched, rec1)
	}

	// 4. Update
	rec1.FullName = "Budi Santoso, M.T."
	rec1.UpdatedAt = time.Now().UTC()
	if err := rRepo.Update(ctx, rec1); err != nil {
		t.Fatalf("rRepo.Update failed: %v", err)
	}
	updated, _ := rRepo.GetByID(ctx, rec1.ID)
	if updated.FullName != "Budi Santoso, M.T." {
		t.Errorf("Expected updated name, got %s", updated.FullName)
	}

	// 5. Test Batch Insertion with Conflict Strategies
	recList := []domain.Recipient{
		{
			ID:            "rec-2",
			ProjectID:     proj.ID,
			FullName:      "Siti Aminah, M.Kom.",
			Email:         "siti@univ.id",
			CertificateID: "CERT-002",
		},
		{
			ID:            "rec-3",
			ProjectID:     proj.ID,
			FullName:      "Ahmad Fauzi",
			Email:         "ahmad@mail.com",
			CertificateID: "CERT-003",
		},
	}

	summary, err := rRepo.BatchCreate(ctx, proj.ID, recList, domain.ConflictStrategyFail)
	if err != nil {
		t.Fatalf("BatchCreate failed: %v", err)
	}
	if summary.ImportedRows != 2 {
		t.Errorf("Expected 2 imported rows, got %d", summary.ImportedRows)
	}

	// Test Duplicate Skip strategy
	dupList := []domain.Recipient{
		{
			ID:            "rec-4",
			ProjectID:     proj.ID,
			FullName:      "Duplicate Budi",
			CertificateID: "CERT-001", // Duplicate!
		},
		{
			ID:            "rec-5",
			ProjectID:     proj.ID,
			FullName:      "New Person",
			CertificateID: "CERT-004", // New
		},
	}

	skipSummary, err := rRepo.BatchCreate(ctx, proj.ID, dupList, domain.ConflictStrategySkip)
	if err != nil {
		t.Fatalf("BatchCreate with skip failed: %v", err)
	}
	if skipSummary.SkippedRows != 1 || skipSummary.ImportedRows != 1 {
		t.Errorf("Expected 1 skipped and 1 imported, got %+v", skipSummary)
	}

	// 6. Test Pagination & Search
	filter := domain.RecipientFilter{
		ProjectID: proj.ID,
		Search:    "Siti",
		Page:      1,
		PageSize:  10,
	}
	searchRes, err := rRepo.List(ctx, filter)
	if err != nil {
		t.Fatalf("rRepo.List search failed: %v", err)
	}
	if searchRes.Total != 1 || searchRes.Items[0].FullName != "Siti Aminah, M.Kom." {
		t.Errorf("Search result mismatch: %+v", searchRes)
	}

	// 7. DeleteBatch
	if err := rRepo.DeleteBatch(ctx, []string{"rec-2", "rec-3"}); err != nil {
		t.Fatalf("DeleteBatch failed: %v", err)
	}
	count, _ := rRepo.Count(ctx, proj.ID)
	if count != 2 { // rec1 and rec5 remain
		t.Errorf("Expected 2 remaining recipients after batch delete, got %d", count)
	}
}
