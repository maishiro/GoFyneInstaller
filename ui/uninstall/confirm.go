package uninstall

import (
	"GoFyneInstaller/script"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ConfirmStep はアンインストール確認ステップを表す
type ConfirmStep struct {
	config *script.InstallConfig
}

// NewConfirmStep は新しいアンインストール確認ステップを作成
func NewConfirmStep(config *script.InstallConfig) *ConfirmStep {
	return &ConfirmStep{config: config}
}

// GetTitle はステップタイトルを返す
func (s *ConfirmStep) GetTitle() string {
	return "Uninstall Confirmation"
}

// GetContent はコンテンツウィジェットを返す
func (s *ConfirmStep) GetContent() fyne.CanvasObject {
	content := container.NewVBox(
		widget.NewLabel("Are you sure you want to uninstall "+s.config.Metadata.Name+"?"),
		widget.NewLabel(""),
		widget.NewLabel("This will remove all installed files and configurations."),
		widget.NewLabel(""),
		widget.NewLabel("Click Next to continue with the uninstallation."),
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
