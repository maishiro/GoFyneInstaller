package uninstall

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// CompleteStep はアンインストール完了ステップを表す
type CompleteStep struct {
	onClose func()
}

// NewCompleteStep は新しいアンインストール完了ステップを作成
func NewCompleteStep(onClose func()) *CompleteStep {
	return &CompleteStep{
		onClose: onClose,
	}
}

// GetTitle はステップタイトルを返す
func (s *CompleteStep) GetTitle() string {
	return "Uninstallation Complete"
}

// GetContent はコンテンツウィジェットを返す
func (s *CompleteStep) GetContent() fyne.CanvasObject {
	finishBtn := widget.NewButton("Finish", func() {
		if s.onClose != nil {
			s.onClose()
		}
	})

	content := container.NewVBox(
		widget.NewLabel("Uninstallation completed successfully!"),
		widget.NewLabel(""),
		widget.NewLabel("The application has been removed from your computer."),
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
	if s.onClose != nil {
		s.onClose()
	}
	return nil
}

// OnPrevious は戻るボタン押下時の処理
func (s *CompleteStep) OnPrevious() error {
	return nil
}
