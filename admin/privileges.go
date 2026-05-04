package admin

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
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

	// ShellExecute wrapper を使用
	shell32 := syscall.NewLazyDLL("shell32.dll")
	shellExecute := shell32.NewProc("ShellExecuteW")

	// ShellExecute(hwnd, lpVerb, lpFile, lpParameters, lpDirectory, nShowCmd)
	exePtr, _ := syscall.UTF16PtrFromString(exe)
	verbPtr, _ := syscall.UTF16PtrFromString("runas")

	// パラメータ文字列を構築
	paramsStr := ""
	for i, arg := range args {
		if i > 0 {
			paramsStr += " "
		}
		paramsStr += `"` + arg + `"`
	}
	paramsPtr, _ := syscall.UTF16PtrFromString(paramsStr)
	dirPtr, _ := syscall.UTF16PtrFromString("")

	// ShellExecute 実行（SW_SHOWNORMAL = 1）
	ret, _, _ := shellExecute.Call(
		uintptr(0), // hwnd = NULL
		uintptr(unsafe.Pointer(verbPtr)),
		uintptr(unsafe.Pointer(exePtr)),
		uintptr(unsafe.Pointer(paramsPtr)),
		uintptr(unsafe.Pointer(dirPtr)),
		uintptr(1), // SW_SHOWNORMAL
	)

	// 戻り値チェック（32以上が成功）
	if ret <= 32 {
		errCode := int(ret)
		errMsg := fmt.Sprintf("[Admin] ShellExecuteW failed with error code %d\n", errCode)
		fmt.Fprint(os.Stderr, errMsg)

		// ログファイルにも記録
		tmpDir := os.TempDir()
		logFile, _ := os.OpenFile(filepath.Join(tmpDir, "GoFyneInstaller_trace.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if logFile != nil {
			fmt.Fprintf(logFile, "[%s] %s", time.Now().Format("15:04:05"), errMsg)
			logFile.Close()
		}

		return fmt.Errorf("ShellExecuteW failed with error code %d", errCode)
	}

	fmt.Fprintf(os.Stderr, "[Admin] Successfully launched elevated process\n")
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
