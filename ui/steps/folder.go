package steps

import (
	"fmt"
	"path/filepath"

	"GoFyneInstaller/script"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// FolderStep はフォルダ選択ステップを表す
type FolderStep struct {
	config       *script.InstallConfig
	selectedPath string
	onPathChange func(string)
	pathEntry    *widget.Entry
	browseButton *widget.Button
	errorLabel   *widget.Label
	window       fyne.Window
}

// NewFolderStep は新しいフォルダ選択ステップを作成
func NewFolderStep(config *script.InstallConfig, onPathChange func(string), window fyne.Window) *FolderStep {
	step := &FolderStep{
		config:       config,
		onPathChange: onPathChange,
		window:       window,
	}

	// デフォルトインストール先を設定
	step.selectedPath = step.getDefaultInstallPath()
	if step.onPathChange != nil {
		step.onPathChange(step.selectedPath)
	}

	return step
}

// GetDefaultInstallPath はデフォルトのインストールパスを取得
func (s *FolderStep) getDefaultInstallPath() string {
	// Program Files に配置
	programFiles := "C:\\Program Files"
	return filepath.Join(programFiles, s.config.Metadata.Name)
}

// GetTitle はステップタイトルを返す
func (s *FolderStep) GetTitle() string {
	return "Select Installation Folder"
}

// GetContent はコンテンツウィジェットを返す
func (s *FolderStep) GetContent() fyne.CanvasObject {
	// パス入力フィールド
	s.pathEntry = widget.NewEntry()
	s.pathEntry.SetText(s.selectedPath)
	s.pathEntry.OnChanged = func(content string) {
		s.selectedPath = content
		if s.onPathChange != nil {
			s.onPathChange(content)
		}
		s.errorLabel.SetText("")
	}

	// フォルダ選択ボタン
	s.browseButton = widget.NewButton("Browse", s.onBrowseClick)

	// エラーラベル
	s.errorLabel = widget.NewLabel("")
	s.errorLabel.Alignment = fyne.TextAlignLeading

	// パス選択コンテナ - "Folder:" ラベルを左に、Browse ボタンを右に固定し、パス表示を全幅で表示
	pathContainer := container.NewBorder(
		nil, nil,
		widget.NewLabel("Folder:"),
		s.browseButton,
		s.pathEntry,
	)

	content := container.NewVBox(
		widget.NewLabel("Select the folder where you want to install "+s.config.Metadata.Name),
		widget.NewLabel(""),
		pathContainer,
		widget.NewLabel(""),
		s.errorLabel,
	)

	return content
}

// onBrowseClick はフォルダ選択ダイアログを表示
func (s *FolderStep) onBrowseClick() {
	if s.window == nil {
		s.errorLabel.SetText("Error: Window reference not available")
		return
	}

	// フォルダ選択ダイアログを表示
	dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil {
			s.errorLabel.SetText(fmt.Sprintf("Error opening folder dialog: %v", err))
			return
		}

		if uri != nil {
			// ユーザーがフォルダを選択した場合
			selectedPath := uri.Path()
			s.selectedPath = selectedPath
			s.pathEntry.SetText(selectedPath)
			s.errorLabel.SetText("")

			// コールバック実行
			if s.onPathChange != nil {
				s.onPathChange(selectedPath)
			}
		}
		// キャンセルされた場合は何もしない
	}, s.window)
}

// Validate はバリデーション処理を実行
func (s *FolderStep) Validate() error {
	if s.selectedPath == "" {
		return fmt.Errorf("installation folder must be selected")
	}
	return nil
}

// OnNext は次へボタン押下時の処理
func (s *FolderStep) OnNext() error {
	return s.Validate()
}

// OnPrevious は戻るボタン押下時の処理
func (s *FolderStep) OnPrevious() error {
	return nil
}

// GetSelectedPath は選択されたパスを取得
func (s *FolderStep) GetSelectedPath() string {
	return s.selectedPath
}

// SetSelectedPath は選択パスを設定
func (s *FolderStep) SetSelectedPath(path string) {
	s.selectedPath = path
	if s.pathEntry != nil {
		s.pathEntry.SetText(path)
	}
	if s.onPathChange != nil {
		s.onPathChange(path)
	}
}
