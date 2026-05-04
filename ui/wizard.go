package ui

import (
	"context"
	"fmt"
	"image/color"
	"io/fs"

	"GoFyneInstaller/script"
	"GoFyneInstaller/ui/steps"
	"GoFyneInstaller/ui/uninstall"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// WizardStep はウィザードの各ステップを表すインターフェース
type WizardStep interface {
	GetTitle() string
	GetContent() fyne.CanvasObject
	Validate() error
	OnNext() error     // 次へボタン押下時の処理
	OnPrevious() error // 戻るボタン押下時の処理
}

// Wizard はインストールまたはアンインストールウィザードを管理
type Wizard struct {
	config        *script.InstallConfig
	assets        fs.FS
	isUninstall   bool
	steps         []WizardStep
	currentIndex  int
	installDir    string
	mainContainer *fyne.Container
	contentArea   *fyne.Container // コンテンツ領域（中央）
	buttonBar     *fyne.Container
	nextBtn       *widget.Button
	prevBtn       *widget.Button
	cancelBtn     *widget.Button
	finishBtn     *widget.Button // 完了画面用ボタン
	onCloseFunc   func()
	window        fyne.Window
}

// NewWizard は新しいウィザードを作成
func NewWizard(config *script.InstallConfig, assets fs.FS, isUninstall bool, window fyne.Window) *Wizard {
	return NewWizardWithCallback(config, assets, isUninstall, nil, window)
}

// NewWizardWithCallback は新しいウィザードを作成（クローズコールバック付き）
func NewWizardWithCallback(config *script.InstallConfig, assets fs.FS, isUninstall bool, onClose func(), window fyne.Window) *Wizard {
	w := &Wizard{
		config:      config,
		assets:      assets,
		isUninstall: isUninstall,
		steps:       []WizardStep{},
		onCloseFunc: onClose,
		window:      window,
	}

	// ウィザードステップを初期化
	w.initializeSteps()

	return w
}

// initializeSteps は各モードのステップを初期化
func (w *Wizard) initializeSteps() {
	if w.isUninstall {
		// アンインストーラーステップ
		confirmStep := uninstall.NewConfirmStep(w.config)
		uninstallingStep := uninstall.NewUninstallingStep(w.config, w.assets)
		completeStep := uninstall.NewCompleteStep(w.onCloseFunc)

		// アンインストール完了時に自動的に Complete ステップに遷移
		uninstallingStep.SetCompleteCallback(func() {
			w.currentIndex = 2 // Complete ステップのインデックス
			w.updateStep()
		})

		w.steps = []WizardStep{
			confirmStep,
			uninstallingStep,
			completeStep,
		}
	} else {
		// インストーラーステップ
		welcomeStep := steps.NewWelcomeStep(w.config)
		folderStep := steps.NewFolderStep(w.config, func(dir string) { w.installDir = dir }, w.window)
		confirmStep := steps.NewConfirmStep(w.config, func() string { return w.installDir })
		installingStep := steps.NewInstallingStep(w.config, w.assets, w.installDir)
		completeStep := steps.NewCompleteStep(w.onCloseFunc)

		// インストール完了時に自動的に Complete ステップに遷移
		installingStep.SetCompleteCallback(func() {
			w.currentIndex = 4 // Complete ステップのインデックス
			w.updateStep()
		})

		w.steps = []WizardStep{
			welcomeStep,
			folderStep,
			confirmStep,
			installingStep,
			completeStep,
		}
	}

	w.currentIndex = 0
}

// GetContent はウィザードコンテンツを取得
func (w *Wizard) GetContent() fyne.CanvasObject {
	// メインコンテナとボタンバーを作成
	w.createLayout()
	w.updateStep()

	return w.mainContainer
}

// createLayout はUIレイアウトを作成
func (w *Wizard) createLayout() {
	// 次へボタン
	w.nextBtn = widget.NewButton("Next", func() {
		if err := w.steps[w.currentIndex].OnNext(); err != nil {
			// エラーダイアログを表示（TODO）
			fmt.Printf("Validation error: %v\n", err)
			return
		}

		if w.currentIndex < len(w.steps)-1 {
			w.currentIndex++
			w.updateStep()
		}
	})

	// 戻るボタン
	w.prevBtn = widget.NewButton("Previous", func() {
		if err := w.steps[w.currentIndex].OnPrevious(); err != nil {
			fmt.Printf("Previous error: %v\n", err)
			return
		}

		if w.currentIndex > 0 {
			w.currentIndex--
			w.updateStep()
		}
	})

	// キャンセルボタン
	w.cancelBtn = widget.NewButton("Cancel", func() {
		// TODO: 確認ダイアログを表示
		if w.onCloseFunc != nil {
			w.onCloseFunc()
		}
	})

	// 完了ボタン（完了画面用）
	w.finishBtn = widget.NewButton("Finish", func() {
		if w.onCloseFunc != nil {
			w.onCloseFunc()
		}
	})

	// ボタンサイズ設定（最小幅を統一）
	buttonMinWidth := float32(90)
	w.prevBtn.Resize(fyne.NewSize(buttonMinWidth, w.prevBtn.MinSize().Height))
	w.nextBtn.Resize(fyne.NewSize(buttonMinWidth, w.nextBtn.MinSize().Height))
	w.cancelBtn.Resize(fyne.NewSize(buttonMinWidth, w.cancelBtn.MinSize().Height))
	w.finishBtn.Resize(fyne.NewSize(buttonMinWidth, w.finishBtn.MinSize().Height))

	// ボタンバー: Previous | Next | [スペース] | Cancel
	// 区切り線を追加
	divider := canvas.NewLine(color.NRGBA{R: 100, G: 100, B: 100, A: 255})
	divider.StrokeWidth = 1

	// ボタンコンテナ
	buttonContainer := container.NewHBox(
		w.prevBtn,
		w.nextBtn,
		layout.NewSpacer(), // 可変スペース
		w.cancelBtn,
	)

	// パディング付きのボタンバー
	w.buttonBar = container.NewVBox(
		divider,
		container.NewPadded(buttonContainer),
	)

	// コンテンツ領域（中央）
	w.contentArea = container.NewVBox()

	// メインコンテナ（BorderLayout で下端にボタンバーを配置）
	w.mainContainer = container.NewBorder(
		nil,           // top
		w.buttonBar,   // bottom
		nil,           // left
		nil,           // right
		w.contentArea, // center
	)
}

// updateStep は現在のステップに基づいてUIを更新
func (w *Wizard) updateStep() {
	if w.currentIndex >= len(w.steps) {
		return
	}

	step := w.steps[w.currentIndex]

	// タイトルを更新
	titleWidget := widget.NewRichTextFromMarkdown(fmt.Sprintf("# %s", step.GetTitle()))

	// メインコンテンツを更新
	contentBox := container.NewVBox(step.GetContent())

	// コンテンツ領域のオブジェクトを更新
	w.contentArea.Objects = []fyne.CanvasObject{
		titleWidget,
		contentBox,
	}

	// ボタンバーを更新
	w.updateButtonBar()
	w.mainContainer.Refresh()
}

// updateButtonBar はボタンバーの状態を更新
func (w *Wizard) updateButtonBar() {
	// インストール/アンインストール中のステップではボタンを無効にする
	isInstallingStep := !w.isUninstall && w.currentIndex >= 3
	isUninstallingStep := w.isUninstall && w.currentIndex >= 1

	if isInstallingStep || isUninstallingStep {
		w.prevBtn.Disable()
		w.nextBtn.Disable()
		w.cancelBtn.Disable()
		return
	}

	// 最後のステップでは Finish ボタンをフルワイド配置
	if w.currentIndex == len(w.steps)-1 {
		// 区切り線を追加
		divider := canvas.NewLine(color.NRGBA{R: 100, G: 100, B: 100, A: 255})
		divider.StrokeWidth = 1

		// Finish ボタンをフルワイドで配置（Border で全幅に広げる）
		finishContainer := container.NewBorder(
			nil, nil, nil, nil,
			w.finishBtn,
		)

		// Finish ボタン用ボタンバー
		finishButtonBar := container.NewVBox(
			divider,
			container.NewPadded(finishContainer),
		)

		// メインコンテナを Finish ボタン用に再構築
		w.mainContainer = container.NewBorder(
			nil,             // top
			finishButtonBar, // bottom
			nil,             // left
			nil,             // right
			w.contentArea,   // center
		)
		return
	}

	// 通常のステップ：Previous, Next, Cancel を表示
	w.prevBtn.Show()
	w.nextBtn.Show()
	w.cancelBtn.Show()

	// 最初のステップでは Previous を無効化
	if w.currentIndex == 0 {
		w.prevBtn.Disable()
	} else {
		w.prevBtn.Enable()
	}

	// Next と Cancel を有効化
	w.nextBtn.Enable()
	w.cancelBtn.Enable()

	// Next ボタンのテキストと動作を設定
	w.nextBtn.SetText("Next")
	w.nextBtn.OnTapped = func() {
		if err := w.steps[w.currentIndex].OnNext(); err != nil {
			fmt.Printf("Validation error: %v\n", err)
			return
		}

		if w.currentIndex < len(w.steps)-1 {
			w.currentIndex++
			w.updateStep()
		}
	}

	// 通常のボタンバー
	divider := canvas.NewLine(color.NRGBA{R: 100, G: 100, B: 100, A: 255})
	divider.StrokeWidth = 1

	buttonContainer := container.NewHBox(
		w.prevBtn,
		w.nextBtn,
		layout.NewSpacer(), // 可変スペース
		w.cancelBtn,
	)

	normalButtonBar := container.NewVBox(
		divider,
		container.NewPadded(buttonContainer),
	)

	// メインコンテナを通常のボタンバー用に再構築
	w.mainContainer = container.NewBorder(
		nil,             // top
		normalButtonBar, // bottom
		nil,             // left
		nil,             // right
		w.contentArea,   // center
	)
}

// SetOnClose はウィザード終了時のコールバックを設定
func (w *Wizard) SetOnClose(fn func()) {
	w.onCloseFunc = fn
}

// GetInstallDir はインストール先ディレクトリを取得
func (w *Wizard) GetInstallDir() string {
	return w.installDir
}

// GetContext はキャンセル可能なコンテキストを取得
func (w *Wizard) GetContext() context.Context {
	return context.Background()
}
