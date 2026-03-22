package steps

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// CompleteStep はインストール完了ステップを表す
type CompleteStep struct {
	onClose func()
}

// NewCompleteStep は新しい完了ステップを作成
func NewCompleteStep(onClose func()) *CompleteStep {
	return &CompleteStep{
		onClose: onClose,
	}
}

// GetTitle はステップタイトルを返す
func (s *CompleteStep) GetTitle() string {
	return "Installation Complete"
}

// GetContent はコンテンツウィジェットを返す
func (s *CompleteStep) GetContent() fyne.CanvasObject {
	finishBtn := widget.NewButton("Finish", func() {
		if s.onClose != nil {
			s.onClose()
		}
	})

	content := container.NewVBox(
		widget.NewLabel("Installation completed successfully!"),
		widget.NewLabel(""),
		widget.NewLabel("The application has been installed to your computer."),
		widget.NewLabel(""),
		widget.NewLabel("Click Finish to close this wizard."),
		widget.NewLabel(""),
		finishBtn,
	)
	return content
}

// Validate はバリデーション処理を実行
func (s *CompleteStep) Validate() error {
	return nil
}

// OnNext は次へボタン押下時の処理
func (s *CompleteStep) OnNext() error {
	return nil
}

// OnPrevious は戻るボタン押下時の処理
func (s *CompleteStep) OnPrevious() error {
	return nil
}
