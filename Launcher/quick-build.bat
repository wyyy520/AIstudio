@echo off
chcp 65001 >nul 2>&1
setlocal

cd /d "%~dp0"

echo.
echo  [Build] Compiling AIStudio Launcher...
echo.

go build -ldflags="-s -w" -o AIStudio.exe .

if %ERRORLEVEL% neq 0 (
    echo   [ERROR] Build failed!
    pause
    exit /b 1
)

copy /Y AIStudio.exe "..\AIStudio.exe" >nul 2>&1

echo   [OK] AIStudio.exe built successfully
echo   [OK] Copied to project root
echo.

endlocal