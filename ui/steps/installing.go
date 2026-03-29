package steps

import (
	"context"
	"embed"
	"fmt"
	"log"
	"strings"

	"GoFyneInstaller/installer"
	"GoFyneInstaller/script"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

// InstallingStep はインストール進捗ステップを表す
type InstallingStep struct {
	config      *script.InstallConfig
	assets      embed.FS
	installDir  string
	progressBar *widget.ProgressBar
	statusLabel *widget.Label
	logText     *widget.RichText

	// Bindings for thread-safe UI updates
	progressBinding   binding.Float
	statusBinding     binding.String
	logBinding        binding.String
	completionBinding binding.Bool

	fullLog string // キャッシュされたログ全文

	engine     *installer.Engine
	ctx        context.Context
	cancel     context.CancelFunc
	onComplete func() // インストール完了時のコールバック
}

// NewInstallingStep は新しいインストール中ステップを作成
func NewInstallingStep(config *script.InstallConfig, assets embed.FS, installDir string) *InstallingStep {
	ctx, cancel := context.WithCancel(context.Background())

	step := &InstallingStep{
		config:            config,
		assets:            assets,
		installDir:        installDir,
		ctx:               ctx,
		cancel:            cancel,
		progressBinding:   binding.NewFloat(),
		statusBinding:     binding.NewString(),
		logBinding:        binding.NewString(),
		completionBinding: binding.NewBool(),
	}

	return step
}

// SetCompleteCallback は完了時のコールバックを設定
func (s *InstallingStep) SetCompleteCallback(callback func()) {
	s.onComplete = callback
}

// GetTitle はステップタイトルを返す
func (s *InstallingStep) GetTitle() string {
	return "Installing"
}

// GetContent はコンテンツウィジェットを返す
func (s *InstallingStep) GetContent() fyne.CanvasObject {
	// プログレスバー
	s.progressBar = widget.NewProgressBarWithData(s.progressBinding)
	s.progressBar.Min = 0
	s.progressBar.Max = 100

	// ステータスラベル
	s.statusLabel = widget.NewLabelWithData(s.statusBinding)
	s.statusBinding.Set("Preparing installation...")

	// ログテキスト
	s.logText = widget.NewRichText()
	s.fullLog = "Initializing...\n"
	s.logBinding.Set(s.fullLog)

	// ログ更新用のリスナー（メインスレッドで実行されることが保証されている）
	s.logBinding.AddListener(binding.NewDataListener(func() {
		val, _ := s.logBinding.Get()
		s.logText.ParseMarkdown(val)
		s.logText.Refresh()
	}))

	// 完了通知用のリスナー（メインスレッドで実行されることが保証されている）
	s.completionBinding.AddListener(binding.NewDataListener(func() {
		if done, _ := s.completionBinding.Get(); done && s.onComplete != nil {
			s.onComplete()
		}
	}))

	// スクロール可能なログコンテナ
	logScroll := container.NewScroll(s.logText)
	logScroll.SetMinSize(fyne.NewSize(400, 150))

	content := container.NewVBox(
		s.statusLabel,
		s.progressBar,
		widget.NewLabel("Installation Log:"),
		logScroll,
	)

	// インストール処理を非同期で実行
	go s.executeInstallation()

	return content
}

// executeInstallation はインストール処理を実行
func (s *InstallingStep) executeInstallation() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Installation panic: %v", r)
			s.appendLog(fmt.Sprintf("Error: %v", r))
		}
	}()

	// インストールエンジン作成
	log.Printf("Initializing installer engine with target directory: %s", s.installDir)
	s.engine = installer.NewEngine(s.config, s.assets, s.installDir)

	// 進捗チャネルの用意
	progressChan := make(chan installer.ProgressUpdate, 10)

	// インストール実行エラーをキャッチするためのチャネル
	errChan := make(chan error, 1)

	log.Print("Starting installation process in a new goroutine.")
	// インストール実行（別のゴルーチンで実行）
	go func() {
		if err := s.engine.Execute(s.ctx, progressChan); err != nil {
			errChan <- err
			return
		}
		errChan <- nil
	}()

	// 進捗更新を監視
	for {
		select {
		case <-s.ctx.Done():
			s.appendLog("Installation cancelled.")
			log.Print("Installation process cancelled.")
			return

		case update := <-progressChan:
			s.progressBinding.Set(float64(update.Percentage))
			log.Printf("Progress Update: %s (%d%%)", update.Message, update.Percentage) // Log progress to file
			s.statusBinding.Set(update.Message)
			s.appendLog(update.Message)

		case err := <-errChan:
			if err != nil {
				s.appendLog(fmt.Sprintf("Installation failed: %v", err))
			} else {
				log.Print("Installation process completed successfully.")
				s.appendLog("Installation completed successfully!")

				// インストール完了時にコールバックを実行（自動遷移）
				// 完了フラグを立てることで、メインスレッドのリスナーが onComplete を呼び出します
				s.completionBinding.Set(true)
			}
			return
		}
	}
}

// appendLog はログを追加（スレッドセーフな実装）
func (s *InstallingStep) appendLog(message string) {
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

	// ファイルログにも出力
	log.Print(message)

	// ログを蓄積（RichText.String() は存在しないため変数で保持）
	s.fullLog += strings.TrimSpace(message) + "\n"

	// ログ行数を制限（メモリ使用量を抑える）
	lines := strings.Split(s.fullLog, "\n")
	if len(lines) > 1000 {
		lines = lines[len(lines)-1000:]
		s.fullLog = strings.Join(lines, "\n")
	}

	// Binding を更新（これによりリスナーがメインスレッドで起動する）
	s.logBinding.Set(s.fullLog)
}

// Validate はバリデーション処理を実行
func (s *InstallingStep) Validate() error {
	return nil
}

// OnNext は次へボタン押下時の処理（インストール中は無効）
func (s *InstallingStep) OnNext() error {
	return nil
}

// OnPrevious は戻るボタン押下時の処理（インストール中は無効）
func (s *InstallingStep) OnPrevious() error {
	return nil
}

// Cancel はインストール処理をキャンセル
func (s *InstallingStep) Cancel() {
	if s.cancel != nil {
		s.cancel()
	}
}
