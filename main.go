package main

import (
	"embed"
	"flag"
	"fmt"
	"log"

	"GoFyneInstaller/admin"
	"GoFyneInstaller/script"
	"GoFyneInstaller/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
)

//go:embed assets/*
var embeddedAssets embed.FS

func main() {
	// コマンドラインフラグの定義
	uninstallFlag := flag.Bool("uninstall", false, "Run uninstaller")
	silentFlag := flag.Bool("silent", false, "Silent install/uninstall")
	flag.Parse()

	// 管理者権限チェック
	if !admin.IsAdminAlt() {
		// GUI で権限昇格が必要であることを通知
		if !*silentFlag {
			fyneApp := app.NewWithID("GoFyneInstallerAdminCheck")
			mainWindow := fyneApp.NewWindow("Administrator Rights Required")
			mainWindow.Resize(fyne.NewSize(400, 150))

			dlg := dialog.NewInformation(
				"Administrator Rights Required",
				"This installer requires administrator privileges.\n\nThe application will be relaunched with administrator rights.",
				mainWindow,
			)
			dlg.Show()

			go func() {
				admin.RelaunchAsAdmin()
			}()

			mainWindow.ShowAndRun()
		} else {
			// サイレントモードの場合はそのまま昇格
			admin.RelaunchAsAdmin()
		}
		return
	}

	// Fyne アプリケーション初期化
	fyneApp := app.NewWithID("GoFyneInstaller")
	// テーマは Settings から設定可能

	// インストール設定を読み込む
	config, err := script.LoadConfigFromAssets(embeddedAssets)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// ウィザード実行
	if *uninstallFlag {
		// アンインストーラー実行
		runUninstaller(fyneApp, config, *silentFlag)
	} else {
		// インストーラー実行
		runInstaller(fyneApp, config, *silentFlag)
	}
}

// インストーラー実行
func runInstaller(app fyne.App, config *script.InstallConfig, silent bool) {
	mainWindow := app.NewWindow(fmt.Sprintf("%s Setup Wizard", config.Metadata.Name))
	mainWindow.Resize(fyne.NewSize(600, 500))

	// ウィザード作成（ウィンドウ閉じるコールバック付き）
	wizard := ui.NewWizardWithCallback(config, embeddedAssets, false, mainWindow.Close, mainWindow)
	mainWindow.SetContent(wizard.GetContent())

	mainWindow.ShowAndRun()
}

// アンインストーラー実行
func runUninstaller(app fyne.App, config *script.InstallConfig, silent bool) {
	mainWindow := app.NewWindow(fmt.Sprintf("Uninstall %s", config.Metadata.Name))
	mainWindow.Resize(fyne.NewSize(600, 500))

	// アンインストーラーウィザード作成（ウィンドウ閉じるコールバック付き）
	wizard := ui.NewWizardWithCallback(config, embeddedAssets, true, mainWindow.Close, mainWindow)
	mainWindow.SetContent(wizard.GetContent())

	mainWindow.ShowAndRun()
}
