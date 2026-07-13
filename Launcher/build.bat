@echo off
chcp 65001 >nul 2>&1
setlocal

cd /d "%~dp0"

echo.
echo  ╔══════════════════════════════════════════╗
echo  ║     AIStudio Launcher - Build            ║
echo  ╚══════════════════════════════════════════╝
echo.

where go >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo   [ERROR] Go not found. Install from https://go.dev/dl/
    exit /b 1
)

echo   [1/3] Tidying modules...
go mod tidy
if %ERRORLEVEL% neq 0 (
    echo   [ERROR] go mod tidy failed
    exit /b 1
)

echo   [2/3] Compiling AIStudio.exe...
go build -ldflags="-s -w" -o AIStudio.exe .
if %ERRORLEVEL% neq 0 (
    echo   [ERROR] Build failed
    exit /b 1
)

echo   [3/3] Deploying to project root...
copy /Y AIStudio.exe "..\AIStudio.exe" >nul 2>&1

for %%A in ("AIStudio.exe") do set "SIZE=%%~zA"
set /a SIZE_MB=%SIZE% / 1048576

echo.
echo  ──────────────────────────────────────────
echo   Build successful!
echo   Output:  %CD%\AIStudio.exe (%SIZE_MB% MB)
echo   Root:    %CD%\..\AIStudio.exe
echo  ──────────────────────────────────────────
echo.
echo   Usage:
echo     AIStudio.exe              Start all services
echo     AIStudio.exe --dev        Start with Vite dev server
echo     AIStudio.exe --debug      Enable debug logging
echo     AIStudio.exe --version    Show version
echo.

endlocal