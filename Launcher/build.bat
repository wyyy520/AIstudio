@echo off
REM =============================================================================
REM AIStudio Launcher - Build Script
REM =============================================================================
REM 编译 Launcher 并将 AIStudio.exe 复制到项目根目录
REM =============================================================================

setlocal

echo.
echo [Build] 开始编译 AIStudio Launcher...
echo.

REM 切换到 Launcher 目录
cd /d "%~dp0"

REM 编译
go build -o AIStudio.exe .

if %ERRORLEVEL% neq 0 (
    echo.
    echo [Build] 编译失败！
    exit /b 1
)

echo.
echo [Build] 编译成功: %CD%\AIStudio.exe
echo.

REM 复制到项目根目录（上一级）
copy /Y AIStudio.exe "..\AIStudio.exe" >nul 2>&1

if %ERRORLEVEL% equ 0 (
    echo [Build] 已复制到项目根目录: ..\AIStudio.exe
) else (
    echo [Build] 警告: 无法复制到项目根目录，请手动复制
)

echo.
echo [Build] 完成！双击 AIStudio.exe 即可启动整个系统。
echo.

endlocal
