package font

import (
	"os"
	"path/filepath"
	"strings"

	"certificate-generator/assets"
)

// Resolver manages TrueType fonts from embedded assets and optional local directories.
type Resolver struct {
	fontDir        string
	fontPaths      map[string]string // normalized name -> absolute path (if on disk)
	fontBytes      map[string][]byte // normalized name -> embedded font bytes
	canonicalNames map[string]string // normalized name -> canonical family name
}

// NewResolver scans both embedded fonts and the specified font directory.
func NewResolver(fontDir string) *Resolver {
	r := &Resolver{
		fontDir:        fontDir,
		fontPaths:      make(map[string]string),
		fontBytes:      make(map[string][]byte),
		canonicalNames: make(map[string]string),
	}
	r.scanEmbedded()
	r.scanDisk()
	return r
}

func normalizeFontKey(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, "_", "")
	return s
}

func (r *Resolver) scanEmbedded() {
	names, err := assets.ListFonts()
	if err != nil {
		return
	}
	for _, name := range names {
		data, err := assets.ReadFont(name)
		if err != nil {
			continue
		}
		base := strings.TrimSuffix(name, filepath.Ext(name))
		cleanName := strings.ReplaceAll(base, "-Regular", "")

		r.registerFontBytes(base, cleanName, data)
		r.registerFontBytes(cleanName, cleanName, data)
	}
}

func (r *Resolver) registerFontBytes(alias, canonical string, data []byte) {
	lower := strings.ToLower(alias)
	norm := normalizeFontKey(alias)

	r.fontBytes[lower] = data
	r.canonicalNames[lower] = canonical

	r.fontBytes[alias] = data
	r.canonicalNames[alias] = canonical

	r.fontBytes[norm] = data
	r.canonicalNames[norm] = canonical
}

func (r *Resolver) scanDisk() {
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

			r.registerDiskPath(base, cleanName, fullPath)
			r.registerDiskPath(cleanName, cleanName, fullPath)
		}
	}
}

func (r *Resolver) registerDiskPath(alias, canonical, path string) {
	lower := strings.ToLower(alias)
	norm := normalizeFontKey(alias)

	r.fontPaths[lower] = path
	r.canonicalNames[lower] = canonical

	r.fontPaths[alias] = path
	r.canonicalNames[alias] = canonical

	r.fontPaths[norm] = path
	r.canonicalNames[norm] = canonical
}

// ResolveBytes returns canonical name and raw TTF bytes (embedded or read from disk).
func (r *Resolver) ResolveBytes(family string) (string, []byte, bool) {
	trimmed := strings.TrimSpace(family)
	lower := strings.ToLower(trimmed)
	norm := normalizeFontKey(trimmed)

	// Direct embedded match
	if data, ok := r.fontBytes[trimmed]; ok {
		return r.canonicalNames[trimmed], data, true
	}
	if data, ok := r.fontBytes[lower]; ok {
		return r.canonicalNames[lower], data, true
	}
	if data, ok := r.fontBytes[norm]; ok {
		return r.canonicalNames[norm], data, true
	}

	// Fallback mappings
	switch lower {
	case "sans-serif", "arial", "helvetica", "ubuntu":
		if data, ok := r.fontBytes["ubuntu"]; ok {
			return "Ubuntu", data, true
		}
	case "serif", "times", "times new roman", "liberationserif":
		if data, ok := r.fontBytes["liberationserif"]; ok {
			return "LiberationSerif", data, true
		}
	}

	// Prefix matching
	for k, data := range r.fontBytes {
		if strings.HasPrefix(k, norm) || strings.HasPrefix(norm, k) {
			return r.canonicalNames[k], data, true
		}
	}

	// Check disk if not found embedded
	if canon, path, ok := r.Resolve(family); ok && path != "" {
		if data, err := os.ReadFile(path); err == nil {
			return canon, data, true
		}
	}

	// Default fallback to LiberationSerif
	if data, ok := r.fontBytes["liberationserif"]; ok {
		return "LiberationSerif", data, true
	}

	return "", nil, false
}

// Resolve returns the canonical name and file path for a font family name.
func (r *Resolver) Resolve(family string) (string, string, bool) {
	trimmed := strings.TrimSpace(family)
	lower := strings.ToLower(trimmed)
	norm := normalizeFontKey(trimmed)

	// Check direct match on disk
	if path, ok := r.fontPaths[trimmed]; ok {
		canon := r.canonicalNames[trimmed]
		return canon, path, true
	}
	if path, ok := r.fontPaths[lower]; ok {
		canon := r.canonicalNames[lower]
		return canon, path, true
	}
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

	// Prefix matching on disk
	for k, path := range r.fontPaths {
		if strings.HasPrefix(k, norm) || strings.HasPrefix(norm, k) {
			canon := r.canonicalNames[k]
			return canon, path, true
		}
	}

	// If only available embedded (no disk path), return canonical name with empty path
	if canon, _, ok := r.ResolveBytes(family); ok {
		return canon, "", true
	}

	return "", "", false
}

// AllEmbeddedFonts returns a map of canonical name -> font bytes for all pre-embedded fonts.
func (r *Resolver) AllEmbeddedFonts() map[string][]byte {
	result := make(map[string][]byte)
	for name, data := range r.fontBytes {
		canon := r.canonicalNames[name]
		if _, exists := result[canon]; !exists {
			result[canon] = data
		}
	}
	return result
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
