//go:build !windows

package printer

import "os/exec"

func hideWindow(_ *exec.Cmd) {}
