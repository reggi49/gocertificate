package assets

import (
	"embed"
	"fmt"
	"io/fs"
	"strings"
)

//go:embed all:fonts all:backgrounds
var FS embed.FS

// ReadFont returns raw TTF bytes for a given font filename (e.g. "LiberationSerif-Regular.ttf").
func ReadFont(name string) ([]byte, error) {
	name = strings.TrimPrefix(name, "fonts/")
	data, err := FS.ReadFile("fonts/" + name)
	if err != nil {
		return nil, fmt.Errorf("embedded font %s not found: %w", name, err)
	}
	return data, nil
}

// ListFonts returns list of all embedded font file names.
func ListFonts() ([]string, error) {
	entries, err := fs.ReadDir(FS, "fonts")
	if err != nil {
		return nil, err
	}
	var names []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".ttf") {
			names = append(names, entry.Name())
		}
	}
	return names, nil
}

// ReadBackground returns raw image bytes for a given background filename (e.g. "default_certificate_bg.jpg").
func ReadBackground(name string) ([]byte, error) {
	name = strings.TrimPrefix(name, "backgrounds/")
	data, err := FS.ReadFile("backgrounds/" + name)
	if err != nil {
		return nil, fmt.Errorf("embedded background %s not found: %w", name, err)
	}
	return data, nil
}

