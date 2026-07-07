@echo off
REM ============================================================
REM  AIStudio 自动推送到 GitHub 仓库脚本
REM  远程地址: https://github.com/wyyy520/AIstudio.git
REM  使用方式: 双击运行即可自动提交并推送
REM  注意: 本文件必须以GBK编码保存，否则中文会乱码
REM ============================================================

setlocal enabledelayedexpansion

REM --- 切换到脚本所在目录（即项目根目录）---
cd /d "%~dp0"

echo ============================================================
echo   AIStudio - 自动推送到 GitHub
echo ============================================================
echo.

REM --- 配置远程仓库地址 ---
set "REMOTE_URL=https://github.com/wyyy520/AIstudio.git"

REM --- 检查是否已初始化 Git 仓库 ---
if not exist ".git" (
    echo [初始化] 未检测到 Git 仓库，正在初始化...
    git init
    if errorlevel 1 (
        echo [错误] Git 初始化失败，请检查是否已安装 Git
        pause
        exit /b 1
    )
    echo [成功] Git 仓库已初始化
    echo.
)

REM --- 配置远程仓库（自动检测：不存在则添加，存在则更新）---
git remote get-url origin >nul 2>&1
if errorlevel 1 (
    echo [配置] 添加远程仓库: %REMOTE_URL%
    git remote add origin "%REMOTE_URL%"
) else (
    echo [检查] 远程仓库已存在，更新地址为: %REMOTE_URL%
    git remote set-url origin "%REMOTE_URL%"
)
echo.

REM --- 确保本地分支为 master ---
git checkout -b master >nul 2>&1

REM --- 添加所有更改（包括新增、修改、删除）---
echo [提交] 添加所有更改文件...
git add -A
echo.

REM --- 检查是否有待提交的更改 ---
git diff --cached --quiet >nul 2>&1
if not errorlevel 1 (
    echo [提示] 没有检测到更改，无需提交
    echo.
    pause
    exit /b 0
)

REM --- 用PowerShell生成时间戳（兼容各种日期格式）---
for /f "delims=" %%i in ('powershell -NoProfile -Command "Get-Date -Format yyyy-MM-dd_HHmmss"') do set "TS=%%i"
set "COMMIT_MSG=Auto push %TS%"

echo [提交] 提交信息: %COMMIT_MSG%
git commit -m "%COMMIT_MSG%"
if errorlevel 1 (
    echo [错误] 提交失败
    pause
    exit /b 1
)
echo.

REM --- 推送到远程仓库 ---
echo [推送] 正在推送到 GitHub...
echo.
git push -u origin master
if errorlevel 1 (
    echo.
    echo ============================================================
    echo   [推送失败] 可能原因：
    echo   1. 认证失败 - 需要输入用户名和 Personal Access Token
    echo   2. 首次推送需先 pull: git pull origin master --allow-unrelated-histories
    echo   3. 网络问题 - 检查代理是否正常
    echo.
    echo   凭据配置: git config --global credential.helper store
    echo   然后在终端手动推送一次，输入用户名和Token后会自动记住
    echo ============================================================
    pause
    exit /b 1
)

echo.
echo ============================================================
echo   [成功] 推送完成！
echo   仓库地址: https://github.com/wyyy520/AIstudio.git
echo   提交时间: %TS%
echo ============================================================
echo.
pause
