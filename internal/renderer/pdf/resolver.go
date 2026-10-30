package pdf

import (
	"regexp"
	"strings"
	"unicode"

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

// ToTitleCase capitalizes the first letter of each word in a string,
// making the rest of the letters lowercase, while preserving abbreviations and punctuation.
func ToTitleCase(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(s))

	capitalizeNext := true
	for i, r := range s {
		if unicode.IsLetter(r) {
			if capitalizeNext {
				b.WriteRune(unicode.ToUpper(r))
				capitalizeNext = false
			} else {
				// If immediately following a dot (like M.Sc. or Ph.D.), keep capital if it was already capital
				if i > 1 && s[i-1] == '.' {
					b.WriteRune(unicode.ToUpper(r))
				} else {
					b.WriteRune(unicode.ToLower(r))
				}
			}
		} else {
			b.WriteRune(r)
			if unicode.IsSpace(r) || r == '-' || r == '/' || r == '(' || r == '[' || r == '.' {
				capitalizeNext = true
			}
		}
	}

	return b.String()
}

// ApplyTextTransform applies text casing transformations: "uppercase", "capitalize", "lowercase", or "none".
func ApplyTextTransform(text string, transform string) string {
	switch strings.ToLower(strings.TrimSpace(transform)) {
	case "uppercase":
		return strings.ToUpper(text)
	case "lowercase":
		return strings.ToLower(text)
	case "capitalize":
		return ToTitleCase(text)
	default:
		return text
	}
}
