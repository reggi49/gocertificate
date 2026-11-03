//go:build windows

package app

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"syscall"
	"unsafe"
)

var (
	shell32           = syscall.NewLazyDLL("shell32.dll")
	procShellExecuteW = shell32.NewProc("ShellExecuteW")
)

// openFileNative opens a file using the Windows ShellExecute API.
// This directly invokes the registered handler (e.g. Microsoft Excel) with the exact file path,
// avoiding cmd.exe quote-stripping issues that cause Excel to open blank without documents.
func openFileNative(filePath string) error {
	filePath = filepath.Clean(filePath)
	verbPtr, _ := syscall.UTF16PtrFromString("open")
	filePtr, _ := syscall.UTF16PtrFromString(filePath)

	ret, _, err := procShellExecuteW.Call(
		0,
		uintptr(unsafe.Pointer(verbPtr)),
		uintptr(unsafe.Pointer(filePtr)),
		0,
		0,
		1, // SW_SHOWNORMAL
	)
	if ret <= 32 {
		// Fallback to explorer.exe if ShellExecute fails
		cmd := exec.Command("explorer.exe", filePath)
		if cmdErr := cmd.Start(); cmdErr != nil {
			return fmt.Errorf("gagal membuka berkas dengan ShellExecute (%v) dan explorer (%v)", err, cmdErr)
		}
	}
	return nil
}

// openDirectoryNative opens a folder in Windows File Explorer.
func openDirectoryNative(dirPath string) error {
	dirPath = filepath.Clean(dirPath)
	verbPtr, _ := syscall.UTF16PtrFromString("open")
	dirPtr, _ := syscall.UTF16PtrFromString(dirPath)

	ret, _, _ := procShellExecuteW.Call(
		0,
		uintptr(unsafe.Pointer(verbPtr)),
		uintptr(unsafe.Pointer(dirPtr)),
		0,
		0,
		1, // SW_SHOWNORMAL
	)
	if ret <= 32 {
		cmd := exec.Command("explorer.exe", dirPath)
		return cmd.Start()
	}
	return nil
}

