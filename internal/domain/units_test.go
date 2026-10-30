package domain_test

import (
	"testing"

	"certificate-generator/internal/domain"
)

func TestUnitsConversion(t *testing.T) {
	epsilon := 1e-4

	// A4 dimensions: 210 x 297 mm
	// In PDF points:
	// 210 mm = 210 * 72 / 25.4 = 595.27559 pt
	// 297 mm = 297 * 72 / 25.4 = 841.88976 pt
	a4WidthPt := domain.MmToPt(297.0)
	expectedA4WidthPt := 841.88976
	if !domain.FloatEquals(a4WidthPt, expectedA4WidthPt, epsilon) {
		t.Errorf("MmToPt(297) = %f, expected %f", a4WidthPt, expectedA4WidthPt)
	}

	a4HeightPt := domain.MmToPt(210.0)
	expectedA4HeightPt := 595.27559
	if !domain.FloatEquals(a4HeightPt, expectedA4HeightPt, epsilon) {
		t.Errorf("MmToPt(210) = %f, expected %f", a4HeightPt, expectedA4HeightPt)
	}

	// Roundtrip: mm -> pt -> mm
	roundtripMm := domain.PtToMm(a4WidthPt)
	if !domain.FloatEquals(roundtripMm, 297.0, epsilon) {
		t.Errorf("PtToMm(MmToPt(297)) = %f, expected 297.0", roundtripMm)
	}

	// Screen display pixels at 96 DPI:
	// 25.4 mm = 1 inch = 96 px
	px := domain.MmToPx(25.4, 96.0)
	if !domain.FloatEquals(px, 96.0, epsilon) {
		t.Errorf("MmToPx(25.4, 96) = %f, expected 96.0", px)
	}

	// Roundtrip: mm -> px -> mm
	roundtripPxMm := domain.PxToMm(px, 96.0)
	if !domain.FloatEquals(roundtripPxMm, 25.4, epsilon) {
		t.Errorf("PxToMm(96, 96) = %f, expected 25.4", roundtripPxMm)
	}
}
