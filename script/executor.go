package script

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CommandExecutor はスクリプトコマンドを実行
type CommandExecutor struct {
	installDir string
}

// NewCommandExecutor は新しいコマンド実行器を作成
func NewCommandExecutor(installDir string) *CommandExecutor {
	return &CommandExecutor{
		installDir: installDir,
	}
}

// ExecuteCommand は単一コマンドを実行
func (e *CommandExecutor) ExecuteCommand(ctx context.Context, cmd *Command) error {
	switch cmd.Type {
	case "copy_file":
		return e.executeCopyFile(ctx, cmd.Params)
	case "delete_folder":
		return e.executeDeleteFolder(ctx, cmd.Params)
	case "register_service":
		return e.executeRegisterService(ctx, cmd.Params)
	case "unregister_service":
		return e.executeUnregisterService(ctx, cmd.Params)
	case "start_service":
		return e.executeStartService(ctx, cmd.Params)
	case "stop_service":
		return e.executeStopService(ctx, cmd.Params)
	case "run_command":
		return e.executeRunCommand(ctx, cmd.Params)
	case "create_shortcut":
		return e.executeCreateShortcut(ctx, cmd.Params)
	default:
		return fmt.Errorf("unknown command type: %s", cmd.Type)
	}
}

// executeCopyFile はファイルコピーコマンドを実行
func (e *CommandExecutor) executeCopyFile(ctx context.Context, params map[string]interface{}) error {
	source, ok := params["source"].(string)
	if !ok {
		return fmt.Errorf("source parameter is required")
	}

	destination, ok := params["destination"].(string)
	if !ok {
		return fmt.Errorf("destination parameter is required")
	}

	// 変数置換
	destination = e.substituteVariables(destination)

	// ファイルコピー
	data, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	// 対象ディレクトリ作成
	dir := filepath.Dir(destination)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// ファイル書き込み
	if err := os.WriteFile(destination, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// executeDeleteFolder はフォルダ削除コマンドを実行
func (e *CommandExecutor) executeDeleteFolder(ctx context.Context, params map[string]interface{}) error {
	path, ok := params["path"].(string)
	if !ok {
		return fmt.Errorf("path parameter is required")
	}

	// 変数置換
	path = e.substituteVariables(path)

	// パスをトリム（余分な空白を削除）
	path = strings.TrimSpace(path)

	log.Printf("Attempting to delete folder: %s", path)

	// パスが存在しないかチェック
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// フォルダが既に存在しない場合は成功扱い
		log.Printf("Path does not exist: %s", path)
		return nil
	}

	// まず Go の os.RemoveAll で試行
	if err := os.RemoveAll(path); err == nil {
		// 成功
		log.Printf("Successfully deleted: %s", path)
		return nil
	}

	// Go での削除失敗した場合は PowerShell で強制削除を試みる
	log.Printf("Go deletion failed, trying PowerShell deletion: %s", path)

	// PowerShell スクリプト用にパスをエスケープ
	escapedPath := strings.ReplaceAll(path, "'", "''")

	powershellCmd := fmt.Sprintf(`
$path = '%s'
Write-Host "PowerShell: Attempting to delete: $path"
if (Test-Path $path) {
    Write-Host "PowerShell: Path exists, attempting removal"
    try {
        Remove-Item -Path $path -Recurse -Force -ErrorAction Stop
        Write-Host "PowerShell: Successfully deleted: $path"
    } catch {
        Write-Host "PowerShell: Failed to delete: $path - $_"
        exit 1
    }
} else {
    Write-Host "PowerShell: Path does not exist: $path"
    exit 0
}
`, escapedPath)

	cmd := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-Command", powershellCmd)

	// 標準出力とエラー出力をキャプチャ
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// PowerShell での削除も失敗した場合は警告として出力するが、エラーにしない
		// （ファイルがロック中などの理由で失敗することがあるため）
		log.Printf("Warning: Failed to delete folder %s", path)
		log.Printf("  Error: %v", err)
		if stdout.Len() > 0 {
			log.Printf("  Stdout: %s", stdout.String())
		}
		if stderr.Len() > 0 {
			log.Printf("  Stderr: %s", stderr.String())
		}
		return nil
	}

	if stdout.Len() > 0 {
		log.Printf("PowerShell output: %s", stdout.String())
	}

	return nil
}

// executeRegisterService はサービス登録コマンドを実行
func (e *CommandExecutor) executeRegisterService(ctx context.Context, params map[string]interface{}) error {
	name, ok := params["name"].(string)
	if !ok {
		return fmt.Errorf("name parameter is required")
	}

	binaryPath, ok := params["binary_path"].(string)
	if !ok {
		return fmt.Errorf("binary_path parameter is required")
	}

	displayName, ok := params["display_name"].(string)
	if !ok {
		displayName = name
	}

	// 変数置換
	binaryPath = e.substituteVariables(binaryPath)

	// sc.exe create コマンド実行
	cmd := exec.CommandContext(ctx, "sc.exe", "create", name, "binPath="+binaryPath, "DisplayName="+displayName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	return nil
}

// executeUnregisterService はサービス削除コマンドを実行
func (e *CommandExecutor) executeUnregisterService(ctx context.Context, params map[string]interface{}) error {
	name, ok := params["name"].(string)
	if !ok {
		return fmt.Errorf("name parameter is required")
	}

	// sc.exe delete コマンド実行
	cmd := exec.CommandContext(ctx, "sc.exe", "delete", name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to unregister service: %w", err)
	}

	return nil
}

// executeStartService はサービス開始コマンドを実行
func (e *CommandExecutor) executeStartService(ctx context.Context, params map[string]interface{}) error {
	name, ok := params["name"].(string)
	if !ok {
		return fmt.Errorf("name parameter is required")
	}

	// sc.exe start コマンド実行
	cmd := exec.CommandContext(ctx, "sc.exe", "start", name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	return nil
}

// executeStopService はサービス停止コマンドを実行
func (e *CommandExecutor) executeStopService(ctx context.Context, params map[string]interface{}) error {
	name, ok := params["name"].(string)
	if !ok {
		return fmt.Errorf("name parameter is required")
	}

	// sc.exe stop コマンド実行
	cmd := exec.CommandContext(ctx, "sc.exe", "stop", name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop service: %w", err)
	}

	return nil
}

// executeRunCommand は汎用コマンド実行
func (e *CommandExecutor) executeRunCommand(ctx context.Context, params map[string]interface{}) error {
	command, ok := params["command"].(string)
	if !ok {
		return fmt.Errorf("command parameter is required")
	}

	// argsは配列またはスペース区切り文字列
	var args []string
	if argsArray, ok := params["args"].([]interface{}); ok {
		for _, arg := range argsArray {
			args = append(args, fmt.Sprintf("%v", arg))
		}
	}

	// コマンド実行（標準出力・標準エラーをキャプチャ）
	cmd := exec.CommandContext(ctx, command, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := stderr.String()
		stdoutStr := stdout.String()

		// エラー詳細を構築
		errDetail := fmt.Sprintf("failed to execute command [%s %v]: %v", command, args, err)
		if stderrStr != "" {
			errDetail += "\nStderr: " + stderrStr
		}
		if stdoutStr != "" {
			errDetail += "\nStdout: " + stdoutStr
		}

		return fmt.Errorf(errDetail)
	}

	return nil
}

// executeCreateShortcut はショートカット作成コマンドを実行
func (e *CommandExecutor) executeCreateShortcut(ctx context.Context, params map[string]interface{}) error {
	targetPath, ok := params["target"].(string)
	if !ok {
		return fmt.Errorf("target parameter is required")
	}

	shortcutPath, ok := params["shortcut"].(string)
	if !ok {
		return fmt.Errorf("shortcut parameter is required")
	}

	// 変数置換
	targetPath = e.substituteVariables(targetPath)
	shortcutPath = e.substituteVariables(shortcutPath)

	// ショートカット作成用の PowerShell スクリプト
	cmd := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-Command", fmt.Sprintf(`
$WshShell = New-Object -ComObject WScript.Shell
$shortcut = $WshShell.CreateShortcut('%s')
$shortcut.TargetPath = '%s'
$shortcut.WorkingDirectory = Split-Path -Parent '%s'
$shortcut.IconLocation = '%s,0'
$shortcut.Description = 'Shortcut to %s'
$shortcut.Save()
Write-Host "Shortcut created: %s"
`, shortcutPath, targetPath, targetPath, targetPath, "Application", shortcutPath))

	return cmd.Run()
}

// substituteVariables は変数を置換
func (e *CommandExecutor) substituteVariables(input string) string {
	replacer := strings.NewReplacer(
		"{INSTALL_DIR}", e.installDir,
		"{PROGRAM_FILES}", "C:\\Program Files",
		"{WINDOWS}", os.Getenv("WINDIR"),
	)
	return replacer.Replace(input)
}
