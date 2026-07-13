@echo off
chcp 65001 >nul
title AIStudio Git Push

:: AIStudio - 一键推送到 GitHub 主干 (main)
set "GIT=D:\Program Files\Git\bin\git.exe"
set "REMOTE_URL=https://github.com/wyyy520/AIstudio.git"

echo.
echo ========================================
echo   AIStudio Git Push
echo ========================================
echo.

:: 检查 Git
if not exist "%GIT%" (
    echo [!] Git 未找到: "%GIT%"
    pause
    exit /b 1
)

:: 初始化仓库（如果未初始化）
if not exist ".git" (
    "%GIT%" init
    "%GIT%" remote add origin %REMOTE_URL%
    echo [*] Git 仓库已初始化
)

:: 确保 remote 存在
"%GIT%" remote get-url origin >nul 2>&1
if errorlevel 1 (
    "%GIT%" remote add origin %REMOTE_URL%
)

:: 确保本地分支名为 main（主干）
"%GIT%" branch --show-current > .branch.tmp
set /p CUR_BRANCH=<.branch.tmp
del .branch.tmp

if not "%CUR_BRANCH%"=="main" (
    "%GIT%" branch -M main
)

:: 暂存所有变更
"%GIT%" add -A

:: 检查是否有变更需要提交
"%GIT%" diff --cached --quiet >nul 2>&1
if errorlevel 1 (
    "%GIT%" commit -m "update: %DATE% %TIME%"
) else (
    echo [*] 没有新的变更
)

:: 拉取最新代码防止冲突
echo [*] 同步远程最新代码...
"%GIT%" pull origin main --rebase --autostash >nul 2>&1

:: 推送到主干
echo [*] 推送到 GitHub main 主干...
"%GIT%" push origin main
if errorlevel 1 (
    echo [!] 推送失败
    pause
    exit /b 1
)

echo.
echo ========================================
echo   ✓ 推送成功！
echo   主干: main
echo   仓库: https://github.com/wyyy520/AIstudio
echo ========================================
echo.

pause
