package domain

import (
	"errors"
	"strings"
)

// AppSettings holds application-wide preferences.
type AppSettings struct {
	DefaultOutputDir    string `json:"defaultOutputDir"`
	Theme               string `json:"theme"`               // "system", "light", "dark"
	Language            string `json:"language"`            // "id", "en"
	AutoSaveIntervalSec int    `json:"autoSaveIntervalSec"` // default 10 seconds
}

// DefaultSettings returns sensible defaults for new installations.
func DefaultSettings() AppSettings {
	return AppSettings{
		DefaultOutputDir:    "",
		Theme:               "system",
		Language:            "id",
		AutoSaveIntervalSec: 10,
	}
}

// ValidateSettings validates settings values.
func ValidateSettings(s *AppSettings) error {
	switch s.Theme {
	case "system", "light", "dark":
	case "":
		s.Theme = "system"
	default:
		return errors.New("invalid theme: must be 'system', 'light', or 'dark'")
	}

	switch s.Language {
	case "id", "en":
	case "":
		s.Language = "id"
	default:
		return errors.New("invalid language: must be 'id' or 'en'")
	}

	if s.AutoSaveIntervalSec <= 0 {
		s.AutoSaveIntervalSec = 10
	} else if s.AutoSaveIntervalSec > 300 {
		s.AutoSaveIntervalSec = 300
	}

	s.DefaultOutputDir = strings.TrimSpace(s.DefaultOutputDir)
	return nil
}
