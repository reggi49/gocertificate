//go:build darwin

package app

import (
	"os/exec"
	"path/filepath"
)

func openFileNative(filePath string) error {
	cmd := exec.Command("open", filepath.Clean(filePath))
	return cmd.Start()
}

func openDirectoryNative(dirPath string) error {
	cmd := exec.Command("open", filepath.Clean(dirPath))
	return cmd.Start()
}

