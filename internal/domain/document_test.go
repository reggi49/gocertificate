package domain_test

import (
	"testing"

	"certificate-generator/internal/domain"
)

func TestDocumentValidation(t *testing.T) {
	doc := domain.Document{
		SchemaVersion: 1,
		Page: domain.PageSettings{
			WidthMm:         297,
			HeightMm:        210,
			Orientation:     "landscape",
			BackgroundColor: "#FFFFFF",
		},
		Elements: []domain.Element{
			{
				ID:          "el-1",
				Type:        domain.ElementTypeText,
				XMm:         20,
				YMm:         30,
				WidthMm:     100,
				HeightMm:    15,
				RotationDeg: 0,
				Opacity:     1.0,
				Visible:     true,
				ZIndex:      1,
				Properties: domain.ElementProperties{
					Text:       "Certificate of Excellence",
					FontFamily: "LiberationSerif",
					FontSizePt: 24,
					Color:      "#000000",
					Alignment:  "center",
				},
			},
		},
	}

	if err := doc.Validate(); err != nil {
		t.Fatalf("Expected valid document, got error: %v", err)
	}

	jsonStr, err := doc.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize to JSON: %v", err)
	}

	parsedDoc, err := domain.DocumentFromJSON(jsonStr)
	if err != nil {
		t.Fatalf("Failed to deserialize from JSON: %v", err)
	}
	if parsedDoc.Elements[0].ID != "el-1" {
		t.Errorf("Expected element ID el-1, got %s", parsedDoc.Elements[0].ID)
	}

	// Test duplicate ID detection
	docWithDuplicate := doc
	docWithDuplicate.Elements = append(docWithDuplicate.Elements, domain.Element{
		ID: "el-1", // duplicate
	})
	if err := docWithDuplicate.Validate(); err == nil {
		t.Errorf("Expected error for duplicate element ID, got nil")
	}

	// Test invalid schema version
	invalidSchema := doc
	invalidSchema.SchemaVersion = 0
	if err := invalidSchema.Validate(); err == nil {
		t.Errorf("Expected error for schemaVersion = 0, got nil")
	}

	// Test negative page width
	invalidPage := doc
	invalidPage.Page.WidthMm = -10
	if err := invalidPage.Validate(); err == nil {
		t.Errorf("Expected error for negative page width, got nil")
	}
}
