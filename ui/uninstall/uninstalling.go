package uninstall

import (
	"context"
	"embed"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"GoFyneInstaller/script"
	"GoFyneInstaller/uninstaller"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// UninstallingStep はアンインストール進捗ステップを表す
type UninstallingStep struct {
	config      *script.InstallConfig
	assets      embed.FS
	progressBar *widget.ProgressBar
	statusLabel *widget.Label
	logText     *widget.RichText
	engine      *uninstaller.Engine
	ctx         context.Context
	cancel      context.CancelFunc
	onComplete  func() // アンインストール完了時のコールバック
}

// NewUninstallingStep は新しいアンインストール中ステップを作成
func NewUninstallingStep(config *script.InstallConfig, assets embed.FS) *UninstallingStep {
	ctx, cancel := context.WithCancel(context.Background())

	step := &UninstallingStep{
		config: config,
		assets: assets,
		ctx:    ctx,
		cancel: cancel,
	}

	return step
}

// SetCompleteCallback は完了時のコールバックを設定
func (s *UninstallingStep) SetCompleteCallback(callback func()) {
	s.onComplete = callback
}

// GetTitle はステップタイトルを返す
func (s *UninstallingStep) GetTitle() string {
	return "Uninstalling"
}

// GetContent はコンテンツウィジェットを返す
func (s *UninstallingStep) GetContent() fyne.CanvasObject {
	// プログレスバー
	s.progressBar = widget.NewProgressBar()
	s.progressBar.Min = 0
	s.progressBar.Max = 100

	// ステータスラベル
	s.statusLabel = widget.NewLabel("Preparing uninstallation...")

	// ログテキスト
	s.logText = widget.NewRichTextFromMarkdown("Initializing...\n")

	// スクロール可能なログコンテナ
	logScroll := container.NewScroll(s.logText)
	logScroll.SetMinSize(fyne.NewSize(400, 150))

	content := container.NewVBox(
		s.statusLabel,
		s.progressBar,
		widget.NewLabel("Uninstallation Log:"),
		logScroll,
	)

	// アンインストール処理を非同期で実行
	go s.executeUninstallation()

	return content
}

// executeUninstallation はアンインストール処理を実行
func (s *UninstallingStep) executeUninstallation() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Uninstallation panic: %v", r)
			s.appendLog(fmt.Sprintf("Error: %v", r))
		}
	}()

	// Registry からインストール先を取得
	appName := s.config.Metadata.Name
	installDir, err := getInstallDirFromRegistry(appName)
	if err != nil {
		s.appendLog(fmt.Sprintf("Warning: Could not get installation directory from registry: %v", err))
		// Registry から取得できない場合は、デフォルトパスを試す
		installDir = getDefaultInstallPath(appName)
		s.appendLog(fmt.Sprintf("Using default path: %s", installDir))
	}

	// アンインストールエンジン作成
	s.engine = uninstaller.NewEngine(s.config, s.assets)

	// 進捗チャネルの用意
	progressChan := make(chan uninstaller.ProgressUpdate, 10)

	// アンインストール実行エラーをキャッチするためのチャネル
	errChan := make(chan error, 1)

	// アンインストール実行（別のゴルーチンで実行）
	go func() {
		if err := s.engine.Execute(s.ctx, installDir, progressChan); err != nil {
			errChan <- err
			return
		}
		errChan <- nil
	}()

	// 進捗更新を監視
	for {
		select {
		case <-s.ctx.Done():
			s.appendLog("Uninstallation cancelled.")
			return

		case update := <-progressChan:
			s.progressBar.SetValue(float64(update.Percentage))
			s.statusLabel.SetText(update.Message)
			s.appendLog(update.Message)

		case err := <-errChan:
			if err != nil {
				s.appendLog(fmt.Sprintf("Uninstallation failed: %v", err))
			} else {
				s.appendLog("Uninstallation completed successfully!")

				// アンインストール完了時にコールバックを実行（自動遷移）
				if s.onComplete != nil {
					s.onComplete()
				}
			}
			return
		}
	}
}

// getInstallDirFromRegistry は Registry からインストール先を取得
func getInstallDirFromRegistry(appName string) (string, error) {
	cmd := exec.Command("powershell.exe", "-NoProfile", "-Command", fmt.Sprintf(`
$regPath = 'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\%s'
if (Test-Path $regPath) {
    $installDir = (Get-ItemProperty -Path $regPath -Name 'InstallLocation' -ErrorAction SilentlyContinue).InstallLocation
    if ($installDir) {
        Write-Output $installDir.Trim()
    }
}
`, appName))

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to query registry: %w", err)
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		return "", fmt.Errorf("InstallLocation not found in registry")
	}

	return result, nil
}

// getDefaultInstallPath はデフォルトのインストール先パスを返す
func getDefaultInstallPath(appName string) string {
	programFiles := os.Getenv("ProgramFiles")
	if programFiles == "" {
		programFiles = "C:\\Program Files"
	}
	return fmt.Sprintf("%s\\%s", programFiles, appName)
}

// appendLog はログを追加（スレッドセーフな実装）
func (s *UninstallingStep) appendLog(message string) {
	if s.logText == nil {
		return
	}

	// Fyne の UI 更新をメインゴルーチン（UI スレッド）で実行
	// NOTE: 別のゴルーチンから呼ばれることがあるため、UI スレッドセーフにする必要がある
	// ここでは直接更新せず、canvas に Refresh() を依頼する安全な方法に変更
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Failed to append log: %v", r)
		}
	}()

	// ログテキストを安全に更新
	current := s.logText.String()
	safeLogs := current + strings.TrimSpace(message) + "\n"

	// ログ行数を制限（メモリ使用量を抑える）
	lines := strings.Split(safeLogs, "\n")
	if len(lines) > 1000 {
		lines = lines[len(lines)-1000:]
		safeLogs = strings.Join(lines, "\n")
	}

	// テキスト更新とリフレッシュを同時に行う
	s.logText.ParseMarkdown(safeLogs)

	// ウィジェットをリフレッシュしてUI を再描画
	s.logText.Refresh()
}

// Validate はバリデーション処理を実行
func (s *UninstallingStep) Validate() error {
	return nil
}

// OnNext は次へボタン押下時の処理（アンインストール中は無効）
func (s *UninstallingStep) OnNext() error {
	return nil
}

// OnPrevious は戻るボタン押下時の処理（アンインストール中は無効）
func (s *UninstallingStep) OnPrevious() error {
	return nil
}

// Cancel はアンインストール処理をキャンセル
func (s *UninstallingStep) Cancel() {
	if s.cancel != nil {
		s.cancel()
	}
}
