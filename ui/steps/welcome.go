package steps

import (
	"GoFyneInstaller/script"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// WelcomeStep はウェルカムスクリーンを表す
type WelcomeStep struct {
	config *script.InstallConfig
}

// NewWelcomeStep は新しいウェルカムステップを作成
func NewWelcomeStep(config *script.InstallConfig) *WelcomeStep {
	return &WelcomeStep{config: config}
}

// GetTitle はステップタイトルを返す
func (s *WelcomeStep) GetTitle() string {
	return "Welcome"
}

// GetContent はコンテンツウィジェットを返す
func (s *WelcomeStep) GetContent() fyne.CanvasObject {
	content := container.NewVBox(
		widget.NewLabel(s.config.Metadata.Name),
		widget.NewLabel(""),
		widget.NewLabel("Version: "+s.config.Metadata.Version),
		widget.NewLabel("Publisher: "+s.config.Metadata.Publisher),
		widget.NewLabel(""),
		widget.NewLabel(s.config.Metadata.Description),
		widget.NewLabel(""),
		widget.NewLabel("Click Next to continue with the installation."),
	)
	return content
}

// Validate はバリデーション処理を実行
func (s *WelcomeStep) Validate() error {
	return nil
}

// OnNext は次へボタン押下時の処理
func (s *WelcomeStep) OnNext() error {
	return s.Validate()
}

// OnPrevious は戻るボタン押下時の処理
func (s *WelcomeStep) OnPrevious() error {
	return nil
}
