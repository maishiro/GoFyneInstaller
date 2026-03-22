@echo off
echo ==========================================
echo GoFyneInstaller Build
echo ==========================================
echo.

setlocal
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=1

echo [Info] Cleaning cache...
go clean -cache

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
go build -ldflags="-H windowsgui -s -w" -v -o setup.exe

if errorlevel 1 (
    echo [Error] Build failed!
    exit /b 1
)

echo [Success] Build completed!
echo Output: setup.exe
echo.
echo Ready for distribution!

endlocal
