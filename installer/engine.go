package installer

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"

	"GoFyneInstaller/script"
)

// ProgressUpdate はインストール進捗を表す
type ProgressUpdate struct {
	Percentage int
	Message    string
}

// Engine はインストール処理を実行
type Engine struct {
	config     *script.InstallConfig
	assets     embed.FS
	installDir string
}

// NewEngine は新しいインストールエンジンを作成
func NewEngine(config *script.InstallConfig, assets embed.FS, installDir string) *Engine {
	return &Engine{
		config:     config,
		assets:     assets,
		installDir: installDir,
	}
}

// Execute はインストール処理を実行
func (e *Engine) Execute(ctx context.Context, progressChan chan<- ProgressUpdate) error {
	defer close(progressChan)

	// ディレクトリ作成
	progressChan <- ProgressUpdate{Percentage: 5, Message: "Creating installation directory..."}
	if err := os.MkdirAll(e.installDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// ファイルコピー
	progressChan <- ProgressUpdate{Percentage: 10, Message: "Copying files..."}
	if err := e.copyFiles(ctx, progressChan); err != nil {
		return fmt.Errorf("failed to copy files: %w", err)
	}

	// アンインストーラーファイルをコピー（自分自身のコピー）
	progressChan <- ProgressUpdate{Percentage: 55, Message: "Setting up uninstaller..."}
	if err := e.setupUninstaller(ctx); err != nil {
		return fmt.Errorf("failed to setup uninstaller: %w", err)
	}

	// インストールコマンド実行
	progressChan <- ProgressUpdate{Percentage: 60, Message: "Executing installation commands..."}
	if err := e.executeCommands(ctx, e.config.Install.Commands, progressChan); err != nil {
		return fmt.Errorf("failed to execute commands: %w", err)
	}

	// Registry にアンインストール情報を登録
	progressChan <- ProgressUpdate{Percentage: 90, Message: "Registering uninstall information..."}
	if err := e.registerUninstallInfo(); err != nil {
		// Registry登録失敗はエラーにしない（管理者権限が必要な場合がある）
		fmt.Printf("Warning: Failed to register uninstall info: %v\n", err)
	}

	progressChan <- ProgressUpdate{Percentage: 100, Message: "Installation complete"}
	return nil
}

// copyFiles はembedファイルをinstallDirにコピー
func (e *Engine) copyFiles(ctx context.Context, progressChan chan<- ProgressUpdate) error {
	filesWalked := 0
	fileCopied := 0

	// ファイル数をカウント
	err := fs.WalkDir(e.assets, "assets", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			filesWalked++
		}
		return nil
	})
	if err != nil {
		return err
	}

	// ファイルコピー
	err = fs.WalkDir(e.assets, "assets", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// ファイルを読み込み
		data, err := fs.ReadFile(e.assets, path)
		if err != nil {
			return err
		}

		// install先を決定（assetプレフィックスを除外）
		// path は "assets/filename" の形式
		var relPath string
		if len(path) > 7 {
			relPath = path[7:] // "assets/" を除外
		} else {
			relPath = path
		}

		targetPath := filepath.Join(e.installDir, relPath)

		// ディレクトリ作成
		targetDir := filepath.Dir(targetPath)
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", targetPath, err)
		}

		// ファイル書き込み
		if err := os.WriteFile(targetPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", targetPath, err)
		}

		fileCopied++
		progress := 10 + int((float64(fileCopied)/float64(filesWalked))*40)
		progressChan <- ProgressUpdate{
			Percentage: progress,
			Message:    fmt.Sprintf("Copied %d/%d files...", fileCopied, filesWalked),
		}

		return nil
	})

	return err
}

// executeCommands はコマンドを実行
func (e *Engine) executeCommands(ctx context.Context, commands []script.Command, progressChan chan<- ProgressUpdate) error {
	if len(commands) == 0 {
		progressChan <- ProgressUpdate{Percentage: 85, Message: "No installation commands to run"}
		return nil
	}

	executor := script.NewCommandExecutor(e.installDir)

	for i, cmd := range commands {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		progressChan <- ProgressUpdate{
			Percentage: 60 + int((float64(i)/float64(len(commands)))*25),
			Message:    fmt.Sprintf("Executing command: %s...", cmd.Type),
		}

		if err := executor.ExecuteCommand(ctx, &cmd); err != nil {
			return fmt.Errorf("command execution failed (%s): %w", cmd.Type, err)
		}
	}

	progressChan <- ProgressUpdate{Percentage: 85, Message: "Installation scripts completed"}
	return nil
}

// Rollback はインストール失敗時のロールバック
func (e *Engine) Rollback() error {
	// TODO: インストールディレクトリ削除など
	fmt.Printf("Rolling back installation in %s\n", e.installDir)
	return nil
}

// setupUninstaller はアンインストーラーのセットアップ
func (e *Engine) setupUninstaller(ctx context.Context) error {
	// 実行中のexeファイルの場所を取得
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// exeをインストールディレクトリにコピー
	data, err := os.ReadFile(exePath)
	if err != nil {
		return fmt.Errorf("failed to read executable: %w", err)
	}

	uninstallerPath := filepath.Join(e.installDir, "uninstall_setup.exe")
	if err := os.WriteFile(uninstallerPath, data, 0755); err != nil {
		return fmt.Errorf("failed to create uninstaller: %w", err)
	}

	return nil
}

// registerUninstallInfo は Registry にアンインストール情報を登録
func (e *Engine) registerUninstallInfo() error {
	appName := e.config.Metadata.Name
	version := e.config.Metadata.Version
	publisher := e.config.Metadata.Publisher
	uninstallerPath := filepath.Join(e.installDir, "uninstall_setup.exe")

	// PowerShell でレジストリに登録
	regPath := fmt.Sprintf(`HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\%s`, appName)

	cmd := exec.Command("powershell.exe", "-NoProfile", "-Command", fmt.Sprintf(`
$regPath = '%s'
New-Item -Path $regPath -Force | Out-Null
New-ItemProperty -Path $regPath -Name 'DisplayName' -Value '%s' -Force | Out-Null
New-ItemProperty -Path $regPath -Name 'DisplayVersion' -Value '%s' -Force | Out-Null
New-ItemProperty -Path $regPath -Name 'Publisher' -Value '%s' -Force | Out-Null
New-ItemProperty -Path $regPath -Name 'UninstallString' -Value '"%s" --uninstall' -Force | Out-Null
New-ItemProperty -Path $regPath -Name 'InstallLocation' -Value '%s' -Force | Out-Null
New-ItemProperty -Path $regPath -Name 'NoModify' -Value 1 -PropertyType DWORD -Force | Out-Null
New-ItemProperty -Path $regPath -Name 'NoRepair' -Value 1 -PropertyType DWORD -Force | Out-Null
`, regPath, appName, version, publisher, uninstallerPath, e.installDir))

	return cmd.Run()
}
