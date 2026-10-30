package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
)

// Supported canonical element types
const (
	ElementTypeText       = "text"
	ElementTypeImage      = "image"
	ElementTypeRectangle  = "rectangle"
	ElementTypeCircle     = "circle"
	ElementTypeLine       = "line"
	ElementTypeBackground = "background"
	ElementTypeSignature  = "signature"
	ElementTypeLogo       = "logo"
	ElementTypeQRCode     = "qrcode"
)

// PageSettings represents the page size and layout configuration in canonical millimeters.
type PageSettings struct {
	WidthMm         float64 `json:"widthMm"`
	HeightMm        float64 `json:"heightMm"`
	Orientation     string  `json:"orientation"` // "landscape" or "portrait"
	BackgroundColor string  `json:"backgroundColor"`
}

// ElementProperties represents polymorphic attributes for certificate elements.
type ElementProperties struct {
	// Text attributes
	Text              string  `json:"text,omitempty"`
	FontFamily        string  `json:"fontFamily,omitempty"`
	FontSizePt        float64 `json:"fontSizePt,omitempty"`
	FontWeight        int     `json:"fontWeight,omitempty"` // 400 = regular, 700 = bold
	Color             string  `json:"color,omitempty"`
	Alignment         string  `json:"alignment,omitempty"`         // "left", "center", "right"
	VerticalAlignment string  `json:"verticalAlignment,omitempty"` // "top", "middle", "bottom"
	LineHeight        float64 `json:"lineHeight,omitempty"`
	LetterSpacing     float64 `json:"letterSpacing,omitempty"`
	FitMode           string  `json:"fitMode,omitempty"` // "shrink", "wrap", "fixed"

	// Shape / Line attributes
	FillColor    string  `json:"fillColor,omitempty"`
	StrokeColor  string  `json:"strokeColor,omitempty"`
	StrokeWidth  float64 `json:"strokeWidthMm,omitempty"`
	CornerRadius float64 `json:"cornerRadiusMm,omitempty"`

	// Image / Asset attributes
	AssetID      string `json:"assetId,omitempty"`
	RelativePath string `json:"relativePath,omitempty"` // relative to project asset root

	// QR Code attributes
	QRContent     string `json:"qrContent,omitempty"`
	RecoveryLevel string `json:"recoveryLevel,omitempty"` // "L", "M", "Q", "H"
}

// Element represents an individual design element positioned in millimeters.
type Element struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`
	XMm         float64           `json:"xMm"`
	YMm         float64           `json:"yMm"`
	WidthMm     float64           `json:"widthMm"`
	HeightMm    float64           `json:"heightMm"`
	RotationDeg float64           `json:"rotationDeg"`
	Opacity     float64           `json:"opacity"` // 0.0 to 1.0
	Visible     bool              `json:"visible"`
	Locked      bool              `json:"locked"`
	ZIndex      int               `json:"zIndex"`
	Properties  ElementProperties `json:"properties"`
}

// Document represents the canonical versioned certificate template document model.
type Document struct {
	SchemaVersion int          `json:"schemaVersion"`
	Page          PageSettings `json:"page"`
	Elements      []Element    `json:"elements"`
}

// Validate validates the document against canonical rules.
func (d *Document) Validate() error {
	if d.SchemaVersion < 1 {
		return errors.New("invalid schemaVersion: must be >= 1")
	}
	if d.Page.WidthMm <= 0 || math.IsNaN(d.Page.WidthMm) || math.IsInf(d.Page.WidthMm, 0) {
		return errors.New("invalid page width: must be a positive number")
	}
	if d.Page.HeightMm <= 0 || math.IsNaN(d.Page.HeightMm) || math.IsInf(d.Page.HeightMm, 0) {
		return errors.New("invalid page height: must be a positive number")
	}

	seenIDs := make(map[string]bool)
	for i, el := range d.Elements {
		if el.ID == "" {
			return fmt.Errorf("element at index %d has empty ID", i)
		}
		if seenIDs[el.ID] {
			return fmt.Errorf("duplicate element ID: %s", el.ID)
		}
		seenIDs[el.ID] = true

		if math.IsNaN(el.XMm) || math.IsInf(el.XMm, 0) || math.IsNaN(el.YMm) || math.IsInf(el.YMm, 0) {
			return fmt.Errorf("element %s has invalid position coordinates", el.ID)
		}
		if el.WidthMm < 0 || math.IsNaN(el.WidthMm) || math.IsInf(el.WidthMm, 0) {
			return fmt.Errorf("element %s has invalid width", el.ID)
		}
		if el.HeightMm < 0 || math.IsNaN(el.HeightMm) || math.IsInf(el.HeightMm, 0) {
			return fmt.Errorf("element %s has invalid height", el.ID)
		}
		if el.Opacity < 0 || el.Opacity > 1.0 || math.IsNaN(el.Opacity) {
			return fmt.Errorf("element %s opacity must be between 0.0 and 1.0", el.ID)
		}
	}
	return nil
}

// ToJSON serializes the document to JSON string.
func (d *Document) ToJSON() (string, error) {
	bytes, err := json.Marshal(d)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// DocumentFromJSON deserializes and validates a document from JSON string.
func DocumentFromJSON(jsonStr string) (*Document, error) {
	var doc Document
	if err := json.Unmarshal([]byte(jsonStr), &doc); err != nil {
		return nil, fmt.Errorf("failed to parse document JSON: %w", err)
	}
	if err := doc.Validate(); err != nil {
		return nil, fmt.Errorf("document validation failed: %w", err)
	}
	return &doc, nil
}
