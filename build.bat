@echo off
echo ==========================================
echo GoFyneInstaller Build
echo ==========================================
echo.

setlocal
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=1

REM Parse arguments
set DEBUG_BUILD=0
if "%1"=="debug" set DEBUG_BUILD=1

echo [Info] Build mode:
if %DEBUG_BUILD% equ 1 (
    echo   Mode=DEBUG (local assets, fast build)
) else (
    echo   Mode=RELEASE (zip embedded, optimized)
)
echo.

echo [Info] Cleaning cache...
go clean -cache

REM Check for release mode requirements and create zip if needed
if %DEBUG_BUILD% equ 0 (
    echo [Info] Creating/updating compressed assets...
    powershell -ExecutionPolicy Bypass -File script/create-assets-zip.ps1
    if errorlevel 1 (
        echo [Error] Failed to create assets zip
        exit /b 1
    )
    echo [Info] Compressed assets ready: installers.zip
) else (
    echo [Info] Development mode - using local assets
)
echo.

echo [Info] Checking for rsrc tool...
where rsrc >nul 2>&1
if errorlevel 1 (
    echo [Info] Installing rsrc tool...
    go install github.com/akavel/rsrc@latest
)

if exist manifest.xml (
    echo [Info] Generating resource file from manifest.xml...
    rsrc -manifest manifest.xml -o rsrc_windows_amd64.syso
    if errorlevel 1 (
        echo [Warning] rsrc failed, continuing without manifest embedding...
    ) else (
        echo [Info] Resource file generated
    )
)

echo [Info] Building setup.exe...
if exist setup.exe (
    echo [Info] Removing old setup.exe...
    del /f /q setup.exe
)
if %DEBUG_BUILD% equ 1 (
    go build -tags debug -ldflags="-H windowsgui -s -w" -v -o setup.exe
) else (
    go build -ldflags="-H windowsgui -s -w" -v -o setup.exe
)

if errorlevel 1 (
    echo [Error] Build failed!
    exit /b 1
)

REM Restore backed up files if release build
if %DEBUG_BUILD% equ 0 (
    if exist "installers.zip" (
        echo [Info] Successfully created setup.exe with embedded assets
    )
)

echo [Success] Build completed!
echo Output: setup.exe
echo.
echo Ready for distribution!

endlocal
