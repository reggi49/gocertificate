package pdf

import (
	"regexp"
	"strings"

	"certificate-generator/internal/domain"
)

var placeholderRegex = regexp.MustCompile(`\{\{([a-zA-Z0-9_-]+)\}\}`)

// ExtractPlaceholders finds all unique {{variable}} placeholders present in a document.
func ExtractPlaceholders(doc *domain.Document) []string {
	seen := make(map[string]bool)
	var result []string

	for _, el := range doc.Elements {
		var matches [][]string
		if el.Type == domain.ElementTypeText {
			matches = placeholderRegex.FindAllStringSubmatch(el.Properties.Text, -1)
		} else if el.Type == domain.ElementTypeQRCode {
			matches = placeholderRegex.FindAllStringSubmatch(el.Properties.QRContent, -1)
		}

		for _, m := range matches {
			if len(m) > 1 && !seen[m[1]] {
				seen[m[1]] = true
				result = append(result, m[1])
			}
		}
	}

	return result
}

func normalizePlaceholderKey(k string) string {
	k = strings.ToLower(k)
	k = strings.ReplaceAll(k, "_", "")
	k = strings.ReplaceAll(k, "-", "")
	return k
}

// ValidateRequiredVariables checks whether all required variables exist and are non-empty.
func ValidateRequiredVariables(doc *domain.Document, variables map[string]string) []string {
	placeholders := ExtractPlaceholders(doc)
	var missing []string

	normVars := make(map[string]string, len(variables))
	for k, v := range variables {
		normVars[normalizePlaceholderKey(k)] = v
	}

	for _, p := range placeholders {
		val, ok := variables[p]
		if !ok {
			val, ok = normVars[normalizePlaceholderKey(p)]
		}
		if !ok || strings.TrimSpace(val) == "" {
			missing = append(missing, p)
		}
	}

	return missing
}

// ReplacePlaceholders replaces {{key}} with values from the variables map without executing code.
// Lookup is case-insensitive and normalizes underscores/hyphens.
func ReplacePlaceholders(input string, variables map[string]string) string {
	normVars := make(map[string]string, len(variables))
	for k, v := range variables {
		normVars[normalizePlaceholderKey(k)] = v
	}

	return placeholderRegex.ReplaceAllStringFunc(input, func(m string) string {
		match := placeholderRegex.FindStringSubmatch(m)
		if len(match) > 1 {
			rawKey := match[1]
			if val, ok := variables[rawKey]; ok {
				return val
			}
			if val, ok := normVars[normalizePlaceholderKey(rawKey)]; ok {
				return val
			}
		}
		return m
	})
}
