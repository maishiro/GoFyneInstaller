# ==========================================
# GoFyneInstaller Build Script (PowerShell)
# ==========================================
#
# 使用方法:
#   .\build.ps1                  # 本番ビルド（最適化、zip埋め込み） ← デフォルト
#   .\build.ps1 -Debug           # 開発ビルド（高速、ローカル参照）
#   .\build.ps1 -Clean           # キャッシュ削除後に本番ビルド
#   .\build.ps1 -Debug -Clean    # キャッシュ削除後に開発ビルド
#
# オプション:
#   -Debug     : 開発版としてビルド（ローカルアセット参照、高速）
#   -Clean     : go clean を実行してからビルド
#

param(
    [switch]$Debug,
    [switch]$Clean
)

$ErrorActionPreference = "Stop"

Write-Host ""
Write-Host "=========================================="
if ($Debug) {
    Write-Host "GoFyneInstaller Build (DEBUG/DEVELOPMENT)"
} else {
    Write-Host "GoFyneInstaller Build (RELEASE)"
}
Write-Host "=========================================="
Write-Host ""

# 環境変数設定
$env:GOOS = "windows"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = 1

Write-Host "[Info] Build Environment:"
Write-Host "  GOOS=$($env:GOOS)"
Write-Host "  GOARCH=$($env:GOARCH)"
Write-Host "  CGO_ENABLED=$($env:CGO_ENABLED)"
if ($Debug) {
    Write-Host "  Mode=DEBUG (local assets, fast build)"
} else {
    Write-Host "  Mode=RELEASE (zip embedded, optimized)"
}
Write-Host ""

# クリーンオプション
if ($Clean) {
    Write-Host "[Info] Cleaning previous builds..."
    try {
        go clean -cache
        Write-Host "[Info] Cache cleaned"
    } catch {
        Write-Host "[Error] Failed to clean cache: $_"
        exit 1
    }
}

# リリースモードでは zip ファイルが必要
if (-not $Debug) {
    Write-Host "[Info] Creating/updating compressed assets..."
    try {
        & powershell -ExecutionPolicy Bypass -File script/create-assets-zip.ps1 2>&1
        if ($LASTEXITCODE -ne 0) {
            Write-Host "[Error] Failed to create assets zip"
            exit 1
        }
    } catch {
        Write-Host "[Error] Failed to run create-assets-zip.ps1: $_"
        exit 1
    }
    Write-Host "[Info] Compressed assets ready: installers.zip"
} else {
    Write-Host "[Info] Development mode - using local assets"
}

Write-Host ""

# ビルドフラグ設定
$ldflags = "-H windowsgui -s -w"  # GUI subsystem + Strip symbols + Strip debug

Write-Host "[Info] Checking for rsrc tool (for manifest embedding)..."
$rsrcPath = (where.exe rsrc 2>$null)
if (-not $rsrcPath) {
    Write-Host "[Info] Installing rsrc tool..."
    try {
        & go install github.com/akavel/rsrc@latest 2>&1
        if ($LASTEXITCODE -ne 0) {
            Write-Host "[Warning] rsrc installation may have issues, but continuing..."
        }
        $rsrcPath = (where.exe rsrc 2>$null)
    } catch {
        Write-Host "[Warning] Could not install rsrc: $_"
    }
}

# manifest.xml がある場合、rsrc で .syso ファイルを生成
if (Test-Path "manifest.xml") {
    Write-Host "[Info] Generating resource file from manifest.xml..."
    try {
        if ($rsrcPath) {
            & rsrc -manifest manifest.xml -o rsrc_windows_amd64.syso 2>&1
            if ($LASTEXITCODE -eq 0) {
                Write-Host "[Info] Resource file generated: rsrc_windows_amd64.syso"
            } else {
                Write-Host "[Warning] rsrc failed, manifest may not be embedded"
            }
        } else {
            Write-Host "[Warning] rsrc not found, manifest will not be embedded"
        }
    } catch {
        Write-Host "[Warning] Failed to generate resource: $_"
    }
}

Write-Host "[Info] Building setup.exe with optimized flags..."
Write-Host "  Flags: $ldflags"
Write-Host ""

# 既存の setup.exe を削除
if (Test-Path "setup.exe") {
    Write-Host "[Info] Removing old setup.exe..."
    Remove-Item -Path setup.exe -Force
}

# ビルドコマンド構築
if ($Debug) {
    # デバッグモード：debug タグ付き
    Write-Host "[Info] Debug build (local assets)..."
    & go build -tags debug -ldflags="$ldflags" -v -o setup.exe 2>&1
} else {
    # リリースモード：debug タグなし
    Write-Host "[Info] Release build (zip embedded)..."
    & go build -ldflags="$ldflags" -v -o setup.exe 2>&1
}

if ($LASTEXITCODE -ne 0) {
    Write-Host ""
    Write-Host "[Error] Build failed with exit code $LASTEXITCODE"
    exit 1
}

# ビルド成功
Write-Host ""
Write-Host "[Success] Build completed!"
Write-Host ""

# ファイル情報表示
$fileInfo = Get-Item -Path setup.exe
$sizeMB = [math]::Round($fileInfo.Length / 1MB, 2)

Write-Host "Output: setup.exe"
Write-Host "Size: $($fileInfo.Length) bytes (~$sizeMB MB)"
Write-Host "Modified: $($fileInfo.LastWriteTime)"
Write-Host ""

if (-not $Debug) {
    if (Test-Path "installers.zip") {
        $zipInfo = Get-Item -Path "installers.zip"
        $zipMB = [math]::Round($zipInfo.Length / 1MB, 2)
        Write-Host "Embedded assets: installers.zip (~$zipMB MB)"
    }
}

Write-Host ""
Write-Host "Ready for distribution!"
Write-Host ""
