package steps

import (
	"GoFyneInstaller/script"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ConfirmStep はインストール確認ステップを表す
type ConfirmStep struct {
	config        *script.InstallConfig
	getInstallDir func() string
}

// NewConfirmStep は新しいインストール確認ステップを作成
func NewConfirmStep(config *script.InstallConfig, getInstallDir func() string) *ConfirmStep {
	return &ConfirmStep{
		config:        config,
		getInstallDir: getInstallDir,
	}
}

// GetTitle はステップタイトルを返す
func (s *ConfirmStep) GetTitle() string {
	return "Confirm Installation"
}

// GetContent はコンテンツウィジェットを返す
func (s *ConfirmStep) GetContent() fyne.CanvasObject {
	installDir := ""
	if s.getInstallDir != nil {
		installDir = s.getInstallDir()
	}

	// 読み取り専用テキスト表示
	content := container.NewVBox(
		widget.NewLabel("Ready to install "+s.config.Metadata.Name+"."),
		widget.NewLabel(""),
		widget.NewLabel("Installation Details:"),
		widget.NewLabel(""),
		widget.NewLabel("  Application: "+s.config.Metadata.Name+" v"+s.config.Metadata.Version),
		widget.NewLabel("  Publisher: "+s.config.Metadata.Publisher),
		widget.NewLabel("  Installation Folder:"),
		widget.NewLabel("    "+installDir),
		widget.NewLabel(""),
		widget.NewLabel("Click Next to begin the installation."),
	)

	return content
}

// Validate はバリデーション処理を実行
func (s *ConfirmStep) Validate() error {
	return nil
}

// OnNext は次へボタン押下時の処理
func (s *ConfirmStep) OnNext() error {
	return s.Validate()
}

// OnPrevious は戻るボタン押下時の処理
func (s *ConfirmStep) OnPrevious() error {
	return nil
}
