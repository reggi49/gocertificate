package filesystem

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// Windows illegal characters in filename
	illegalChars           = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)
	multiUnderscoreOrSpace = regexp.MustCompile(`[_\s]+`)
	// Reserved Windows file names
	reservedNames = map[string]bool{
		"CON": true, "PRN": true, "AUX": true, "NUL": true,
		"COM1": true, "COM2": true, "COM3": true, "COM4": true,
		"COM5": true, "COM6": true, "COM7": true, "COM8": true, "COM9": true,
		"LPT1": true, "LPT2": true, "LPT3": true, "LPT4": true,
		"LPT5": true, "LPT6": true, "LPT7": true, "LPT8": true, "LPT9": true,
	}
)

// SanitizeFilename cleans a filename to be safely used across Windows, macOS, and Linux.
func SanitizeFilename(name string) string {
	clean := strings.TrimSpace(name)
	if clean == "" {
		return "unnamed"
	}

	// First strip trailing dots and spaces
	clean = strings.TrimRight(clean, ". ")

	// Separate extension if present
	ext := filepathExt(clean)
	base := strings.TrimSuffix(clean, ext)

	// Replace illegal characters with underscore in base
	base = illegalChars.ReplaceAllString(base, "_")
	// Normalize underscores and spaces
	base = multiUnderscoreOrSpace.ReplaceAllString(base, "_")
	base = strings.Trim(base, "._ ")

	if base == "" {
		base = "unnamed"
	}

	// Sanitize extension as well
	if ext != "" {
		ext = illegalChars.ReplaceAllString(ext, "")
		ext = strings.TrimRight(ext, ". ")
	}

	clean = base + ext
	clean = strings.TrimRight(clean, ". ")
	if clean == "" {
		return "unnamed"
	}

	// Check against reserved Windows names (case insensitive)
	baseUpper := strings.ToUpper(base)
	if reservedNames[baseUpper] {
		clean = "_" + clean
	}

	// Bound max length to 150 characters
	if len(clean) > 150 {
		clean = clean[:150]
	}

	return clean
}

func filepathExt(path string) string {
	for i := len(path) - 1; i >= 0 && path[i] != '/' && path[i] != '\\'; i-- {
		if path[i] == '.' {
			return path[i:]
		}
	}
	return ""
}

// ValidateID ensures an identifier contains only alphanumeric characters, underscores, and hyphens.
func ValidateID(id string) error {
	trimmed := strings.TrimSpace(id)
	if trimmed == "" {
		return fmt.Errorf("id cannot be empty")
	}
	for _, r := range trimmed {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
			return fmt.Errorf("id contains invalid character: %c", r)
		}
	}
	return nil
}
