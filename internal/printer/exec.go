package printer

import (
	"context"
	"os/exec"
	"time"
)

const (
	// listPrintersTimeout bounds printer enumeration (PowerShell/CIM can hang on broken spoolers).
	listPrintersTimeout = 20 * time.Second
	// printTimeout bounds a single print job hand-off so a stuck spooler/driver never leaves
	// orphaned PowerShell/Acrobat processes consuming CPU and RAM in the background.
	printTimeout = 3 * time.Minute
	// acrobatPrintTimeout bounds a single Acrobat Reader hand-off (Reader often never exits).
	acrobatPrintTimeout = 45 * time.Second
)

// command builds a context-bound *exec.Cmd with platform specific attributes
// (on Windows: no console window flashing, see hideWindow).
func command(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	hideWindow(cmd)
	// Ensure the process is killed and pipes closed shortly after the context expires.
	cmd.WaitDelay = 5 * time.Second
	return cmd
}
