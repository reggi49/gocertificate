package pdf

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
)

var (
	ErrInvalidPDFHeader = errors.New("invalid PDF header: missing %PDF- signature")
	ErrPDFTooSmall      = errors.New("invalid PDF: file size is suspiciously small (< 1000 bytes)")
	ErrInvalidMediaBox  = errors.New("PDF media box page dimensions do not match expected canonical size")
)

var mediaBoxRegex = regexp.MustCompile(`/MediaBox\s*\[\s*0\s+0\s+([\d\.]+)\s+([\d\.]+)\s*\]`)

// ValidatePDF performs structural validation on a generated PDF byte buffer.
func ValidatePDF(data []byte, expectedWidthPt, expectedHeightPt float64, tolerancePt float64) error {
	if len(data) < 1000 {
		return ErrPDFTooSmall
	}

	// 1. Check %PDF- header in the first 1024 bytes
	headerWindow := data
	if len(headerWindow) > 1024 {
		headerWindow = headerWindow[:1024]
	}
	if !bytes.Contains(headerWindow, []byte("%PDF-")) {
		return ErrInvalidPDFHeader
	}

	// 2. Extract and check MediaBox dimensions if expected dimensions are specified
	if expectedWidthPt > 0 && expectedHeightPt > 0 {
		matches := mediaBoxRegex.FindSubmatch(data)
		if len(matches) >= 3 {
			w, err1 := strconv.ParseFloat(string(matches[1]), 64)
			h, err2 := strconv.ParseFloat(string(matches[2]), 64)
			if err1 == nil && err2 == nil {
				if math.Abs(w-expectedWidthPt) > tolerancePt || math.Abs(h-expectedHeightPt) > tolerancePt {
					return fmt.Errorf("%w: got (%.2f x %.2f pt), expected (%.2f x %.2f pt)",
						ErrInvalidMediaBox, w, h, expectedWidthPt, expectedHeightPt)
				}
			}
		}
	}

	return nil
}
