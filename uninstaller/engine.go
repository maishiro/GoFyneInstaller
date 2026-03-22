package uninstaller

import (
	"context"
	"embed"
	"fmt"
	"os"
	"os/exec"

	"GoFyneInstaller/admin"
	"GoFyneInstaller/script"
)

// ProgressUpdate はアンインストール進捗を表す
type ProgressUpdate struct {
	Percentage int
	Message    string
}

// Engine はアンインストール処理を実行
type Engine struct {
	config *script.InstallConfig
	assets embed.FS
}

// NewEngine は新しいアンインストールエンジンを作成
func NewEngine(config *script.InstallConfig, assets embed.FS) *Engine {
	return &Engine{
		config: config,
		assets: assets,
	}
}

// Execute はアンインストール処理を実行
func (e *Engine) Execute(ctx context.Context, installDir string, progressChan chan<- ProgressUpdate) error {
	defer close(progressChan)

	// アンインストールコマンド実行
	progressChan <- ProgressUpdate{Percentage: 10, Message: "Executing uninstallation commands..."}
	if err := e.executeCommands(ctx, e.config.Uninstall.Commands, installDir, progressChan); err != nil {
		return fmt.Errorf("failed to execute commands: %w", err)
	}

	// Registry からのアンインストール情報削除
	progressChan <- ProgressUpdate{Percentage: 80, Message: "Removing installation information..."}
	if err := e.removeRegistryInfo(e.config.Metadata.Name); err != nil {
		// Registry削除失敗はエラーにしない（管理者権限が必要な場合がある）
		fmt.Printf("Warning: Failed to remove registry info: %v\n", err)
	}

	// アンインストール実行ファイル自体を再起動後に削除予定とマーク
	progressChan <- ProgressUpdate{Percentage: 90, Message: "Marking uninstaller for deletion on reboot..."}
	exe, err := os.Executable()
	if err == nil {
		if err := admin.MarkFileForDeletionOnReboot(exe); err != nil {
			// 削除失敗もエラーにしない（管理者権限が十分でない可能性）
			fmt.Printf("Warning: Failed to mark uninstaller for deletion: %v\n", err)
		}
	}

	progressChan <- ProgressUpdate{Percentage: 100, Message: "Uninstallation complete"}
	return nil
}

// executeCommands はコマンドを実行
func (e *Engine) executeCommands(ctx context.Context, commands []script.Command, installDir string, progressChan chan<- ProgressUpdate) error {
	if len(commands) == 0 {
		progressChan <- ProgressUpdate{Percentage: 50, Message: "No uninstallation commands to run"}
		return nil
	}

	executor := script.NewCommandExecutor(installDir)

	for i, cmd := range commands {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		progressChan <- ProgressUpdate{
			Percentage: 10 + int((float64(i)/float64(len(commands)))*60),
			Message:    fmt.Sprintf("Executing command: %s...", cmd.Type),
		}

		if err := executor.ExecuteCommand(ctx, &cmd); err != nil {
			return fmt.Errorf("command execution failed (%s): %w", cmd.Type, err)
		}
	}

	progressChan <- ProgressUpdate{Percentage: 70, Message: "Uninstallation scripts completed"}
	return nil
}

// removeRegistryInfo は Registry からのアンインストール情報を削除
func (e *Engine) removeRegistryInfo(appName string) error {
	// Windows Registry から以下の場所を削除:
	// HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\{AppName}
	// HKCU\Software\Microsoft\Windows\CurrentVersion\Uninstall\{AppName}

	// PowerShell で Registry 削除コマンドを実行
	cmd := exec.Command("powershell.exe", "-Command",
		fmt.Sprintf(`
if (Test-Path 'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\%s') {
    Remove-Item -Path 'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\%s' -Force -ErrorAction SilentlyContinue
}
if (Test-Path 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\%s') {
    Remove-Item -Path 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\%s' -Force -ErrorAction SilentlyContinue
}
`, appName, appName, appName, appName))

	return cmd.Run()
}
