package font

import (
	"os"
	"path/filepath"
	"strings"
)

// Resolver manages local TrueType fonts for PDF embedding.
type Resolver struct {
	fontDir        string
	fontPaths      map[string]string // normalized name -> absolute path
	canonicalNames map[string]string // normalized name -> canonical family name
}

// NewResolver scans the font directory and registers all available .ttf fonts.
func NewResolver(fontDir string) *Resolver {
	r := &Resolver{
		fontDir:        fontDir,
		fontPaths:      make(map[string]string),
		canonicalNames: make(map[string]string),
	}
	r.scan()
	return r
}

func normalizeFontKey(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, "_", "")
	return s
}

func (r *Resolver) scan() {
	if r.fontDir == "" {
		return
	}

	entries, err := os.ReadDir(r.fontDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(strings.ToLower(name), ".ttf") {
			base := strings.TrimSuffix(name, filepath.Ext(name))
			fullPath := filepath.Join(r.fontDir, name)

			cleanName := strings.ReplaceAll(base, "-Regular", "")

			r.fontPaths[strings.ToLower(base)] = fullPath
			r.canonicalNames[strings.ToLower(base)] = cleanName

			r.fontPaths[strings.ToLower(cleanName)] = fullPath
			r.canonicalNames[strings.ToLower(cleanName)] = cleanName

			r.fontPaths[cleanName] = fullPath
			r.canonicalNames[cleanName] = cleanName

			r.fontPaths[base] = fullPath
			r.canonicalNames[base] = cleanName

			normBase := normalizeFontKey(base)
			r.fontPaths[normBase] = fullPath
			r.canonicalNames[normBase] = cleanName

			normClean := normalizeFontKey(cleanName)
			r.fontPaths[normClean] = fullPath
			r.canonicalNames[normClean] = cleanName
		}
	}
}

// Resolve returns the canonical name and file path for a font family name.
func (r *Resolver) Resolve(family string) (string, string, bool) {
	if len(r.fontPaths) == 0 {
		return "", "", false
	}

	trimmed := strings.TrimSpace(family)
	lower := strings.ToLower(trimmed)
	norm := normalizeFontKey(trimmed)

	// Check direct match
	if path, ok := r.fontPaths[trimmed]; ok {
		canon := r.canonicalNames[trimmed]
		return canon, path, true
	}
	// Check case-insensitive match
	if path, ok := r.fontPaths[lower]; ok {
		canon := r.canonicalNames[lower]
		return canon, path, true
	}
	// Check normalized match (ignoring spaces and hyphens)
	if path, ok := r.fontPaths[norm]; ok {
		canon := r.canonicalNames[norm]
		return canon, path, true
	}

	// Fallback mappings
	switch lower {
	case "sans-serif", "arial", "helvetica", "ubuntu":
		if path, ok := r.fontPaths["ubuntu"]; ok {
			return "Ubuntu", path, true
		}
	case "serif", "times", "times new roman", "liberationserif":
		if path, ok := r.fontPaths["liberationserif"]; ok {
			return "LiberationSerif", path, true
		}
	}

	// Normalized prefix matching fallback (e.g. "sloop" -> SloopScriptOne, "trajan" -> TrajanPro)
	for k, path := range r.fontPaths {
		if strings.HasPrefix(k, norm) || strings.HasPrefix(norm, k) {
			canon := r.canonicalNames[k]
			return canon, path, true
		}
	}

	// Default to first available font
	for name, path := range r.fontPaths {
		canon := r.canonicalNames[name]
		return canon, path, true
	}

	return "", "", false
}

// AvailableFonts returns list of registered font families.
func (r *Resolver) AvailableFonts() []string {
	var list []string
	seen := make(map[string]bool)
	for _, canon := range r.canonicalNames {
		if !seen[canon] {
			list = append(list, canon)
			seen[canon] = true
		}
	}
	return list
}
