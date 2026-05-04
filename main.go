package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"GoFyneInstaller/admin"
	"GoFyneInstaller/logger"
	"GoFyneInstaller/script"
	"GoFyneInstaller/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
)

// embeddedAssets is defined in assets_resources_*.go
// - assets_resources_embed.go (!debug): embeds installers.zip only
// - assets_resources_local.go (debug): references local assets folder

func main() {
	// トレースログ（昇起プロセスでも確認できるように）
	tmpDir := os.TempDir()
	logPath := filepath.Join(tmpDir, "GoFyneInstaller_trace.log")
	logFile, _ := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if logFile != nil {
		fmt.Fprintf(logFile, "[%s] Process started\n", time.Now().Format("15:04:05"))
		logFile.Close()
	}

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

			// 短い遅延後に昇格実行（ユーザーにメッセージが見えるようにするため）
			go func() {
				<-time.After(1 * time.Second)
				if err := admin.RelaunchAsAdmin(); err != nil {
					logFile, _ := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
					if logFile != nil {
						fmt.Fprintf(logFile, "[%s] RelaunchAsAdmin error: %v\n", time.Now().Format("15:04:05"), err)
						logFile.Close()
					}
				}
			}()

			mainWindow.ShowAndRun()
		} else {
			// サイレントモードの場合はそのまま昇起
			if err := admin.RelaunchAsAdmin(); err != nil {
				fmt.Fprintf(os.Stderr, "RelaunchAsAdmin error: %v\n", err)
			}
		}
		return
	}

	// Fyne アプリケーション初期化
	fyneApp := app.NewWithID("GoFyneInstaller")
	// テーマは Settings から設定可能

	// Asset Provider の初期化
	provider := NewAssetProvider(embeddedAssets)

	// インストール設定を読み込む
	config, err := script.LoadConfigFromAssets(provider)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// ログの初期化（アプリケーション名を使用）
	if err := logger.InitLogger(config.Metadata.Name); err != nil {
		log.Printf("Warning: Could not initialize log file: %v", err)
	}
	defer logger.Close()

	// ウィザード実行
	if *uninstallFlag {
		// アンインストーラー実行
		runUninstaller(fyneApp, config, provider, *silentFlag)
	} else {
		// インストーラー実行
		runInstaller(fyneApp, config, provider, *silentFlag)
	}
}

// インストーラー実行
func runInstaller(app fyne.App, config *script.InstallConfig, provider script.AssetProvider, silent bool) {
	mainWindow := app.NewWindow(fmt.Sprintf("%s Setup Wizard", config.Metadata.Name))
	mainWindow.Resize(fyne.NewSize(600, 500))

	// ウィザード作成（ウィンドウ閉じるコールバック付き）
	wizard := ui.NewWizardWithCallback(config, provider.GetFS(), false, mainWindow.Close, mainWindow)
	mainWindow.SetContent(wizard.GetContent())

	mainWindow.ShowAndRun()
}

// アンインストーラー実行
func runUninstaller(app fyne.App, config *script.InstallConfig, provider script.AssetProvider, silent bool) {
	mainWindow := app.NewWindow(fmt.Sprintf("Uninstall %s", config.Metadata.Name))
	mainWindow.Resize(fyne.NewSize(600, 500))

	// アンインストーラーウィザード作成（ウィンドウ閉じるコールバック付き）
	wizard := ui.NewWizardWithCallback(config, provider.GetFS(), true, mainWindow.Close, mainWindow)
	mainWindow.SetContent(wizard.GetContent())

	mainWindow.ShowAndRun()
}
