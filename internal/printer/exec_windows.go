//go:build windows

package printer

import (
	"os/exec"
	"syscall"
)

const createNoWindow = 0x08000000

// hideWindow prevents a console window from being created for child processes (PowerShell etc.).
func hideWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
}
