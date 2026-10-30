package domain

import "math"

const (
	// MillimetersPerInch is the standard exact international inch in millimeters
	MillimetersPerInch = 25.4

	// PointsPerInch is the PostScript / PDF point standard (72 pt / inch)
	PointsPerInch = 72.0

	// DefaultScreenDPI is standard web display pixel density (96 DPI)
	DefaultScreenDPI = 96.0
)

// MmToPt converts millimeters to PDF points.
// points = mm * 72.0 / 25.4
func MmToPt(mm float64) float64 {
	return mm * PointsPerInch / MillimetersPerInch
}

// PtToMm converts PDF points to millimeters.
// mm = points * 25.4 / 72.0
func PtToMm(pt float64) float64 {
	return pt * MillimetersPerInch / PointsPerInch
}

// MmToPx converts millimeters to display pixels given a specific DPI (default 96).
func MmToPx(mm float64, dpi float64) float64 {
	if dpi <= 0 {
		dpi = DefaultScreenDPI
	}
	return mm * dpi / MillimetersPerInch
}

// PxToMm converts display pixels to millimeters given a specific DPI.
func PxToMm(px float64, dpi float64) float64 {
	if dpi <= 0 {
		dpi = DefaultScreenDPI
	}
	return px * MillimetersPerInch / dpi
}

// FloatEquals tests equality within a tolerance epsilon.
func FloatEquals(a, b, epsilon float64) bool {
	return math.Abs(a-b) <= epsilon
}
