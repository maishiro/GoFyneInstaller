package admin

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

// IsAdmin returns true if the current process is running with admin privileges
func IsAdmin() bool {
	_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	if err != nil {
		return false
	}
	return true
}

// IsAdminAlt is an alternative method using Windows API
func IsAdminAlt() bool {
	shell32 := syscall.NewLazyDLL("shell32.dll")

	// IsUserAnAdmin from shell32.dll
	isUserAnAdmin := shell32.NewProc("IsUserAnAdmin")
	ret, _, _ := isUserAnAdmin.Call()

	return ret != 0
}

// RelaunchAsAdmin restarts the current process with admin privileges
func RelaunchAsAdmin() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// 元のコマンドラインを保持
	args := os.Args[1:]

	// ShellExecute wrapper を使用する必要があるため、shell32.ShellExecute を呼び出す
	shell32 := syscall.NewLazyDLL("shell32.dll")
	shellExecute := shell32.NewProc("ShellExecuteW")

	// ShellExecute(hwnd, lpVerb, lpFile, lpParameters, lpDirectory, nShowCmd)
	exePtr, _ := syscall.UTF16PtrFromString(exe)
	verbPtr, _ := syscall.UTF16PtrFromString("runas")
	paramsStr := ""
	for i, arg := range args {
		if i > 0 {
			paramsStr += " "
		}
		paramsStr += `"` + arg + `"`
	}
	paramsPtr, _ := syscall.UTF16PtrFromString(paramsStr)
	dirPtr, _ := syscall.UTF16PtrFromString("")

	// nShowCmd = 1 (SW_SHOWNORMAL)
	shellExecute.Call(
		uintptr(0), // hwnd = NULL
		uintptr(unsafe.Pointer(verbPtr)),
		uintptr(unsafe.Pointer(exePtr)),
		uintptr(unsafe.Pointer(paramsPtr)),
		uintptr(unsafe.Pointer(dirPtr)),
		uintptr(1), // SW_SHOWNORMAL
	)

	// 元のプロセスを終了
	os.Exit(0)
	return nil
}

// MarkFileForDeletionOnReboot はファイルを PC 再起動後に削除されるようにマーク
// Windows API の MoveFileEx を使用してファイルを削除予定管理者権限が必要です
func MarkFileForDeletionOnReboot(filePath string) error {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	moveFileEx := kernel32.NewProc("MoveFileExW")

	// MoveFileEx(lpExistingFileName, NULL, MOVEFILE_DELAY_UNTIL_REBOOT)
	// MOVEFILE_DELAY_UNTIL_REBOOT = 0x00000004
	filePtr, _ := syscall.UTF16PtrFromString(filePath)
	ret, _, err := moveFileEx.Call(
		uintptr(unsafe.Pointer(filePtr)),
		uintptr(0), // lpNewFileName = NULL
		uintptr(4), // MOVEFILE_DELAY_UNTIL_REBOOT
	)

	if ret == 0 {
		return fmt.Errorf("failed to mark file for deletion on reboot: %w", err)
	}

	return nil
}
