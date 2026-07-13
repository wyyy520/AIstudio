@echo off
chcp 936 >nul
title AIStudio E2E Test
color 0B

set ROOT=%~dp0
set ROOT=%ROOT:~0,-1%

echo ============================================
echo   AIStudio End-to-End Test Launcher
echo ============================================
echo.
echo   Project: %ROOT%
echo.
echo Prerequisites:
echo   - Go 1.22+
echo   - Python 3.8+
echo   - Node.js 18+
echo.
echo ============================================
pause

:menu
cls
echo ============================================
echo   Select service to start
echo ============================================
echo.
echo   [1] Start Python Engine only
echo   [2] Start Go Backend only
echo   [3] Start Engine + Backend
echo   [4] Start All (Engine + Backend + Frontend)
echo   [5] Exit
echo.
set /p choice=Enter choice (1-5):

if "%choice%"=="1" goto start_engine
if "%choice%"=="2" goto start_backend
if "%choice%"=="3" goto start_engine
if "%choice%"=="4" goto start_engine
if "%choice%"=="5" goto end
goto menu

:start_engine
cls
echo ============================================
echo   Step 1/3: Python AI Engine
echo ============================================
echo.
echo   Port: 8082
echo.
start "Python Engine" cmd /c "call "%ROOT%\launch_engine.bat""
echo.
echo   Engine window should now be open.
echo   Waiting for it to be ready...
echo.
timeout /t 5 /nobreak >nul
echo [OK] Proceeding...
echo.
goto check_choice_after_engine

:check_choice_after_engine
if "%choice%"=="1" goto done_msg
if "%choice%"=="3" goto start_backend
if "%choice%"=="4" goto start_backend

:start_backend
cls
echo ============================================
echo   Step 2/3: Go Backend
echo ============================================
echo.
echo   Port: 8081
echo.
start "Go Backend" cmd /c "call "%ROOT%\launch_backend.bat""
echo.
echo   Backend window should now be open.
echo   Waiting for it to compile and start...
echo.
timeout /t 10 /nobreak >nul
echo [OK] Proceeding...
echo.
goto check_choice_after_backend

:check_choice_after_backend
if "%choice%"=="2" goto done_msg
if "%choice%"=="4" goto start_frontend
goto check_services

:start_frontend
cls
echo ============================================
echo   Step 3/3: Frontend
echo ============================================
echo.
echo   Port: 5173
echo.
start "Frontend" cmd /c "call "%ROOT%\launch_frontend.bat""
echo.
echo   Frontend window should now be open.
echo.
timeout /t 8 /nobreak >nul
echo [OK] Proceeding...
echo.
goto check_services

:check_services
cls
echo ============================================
echo   All services launched!
echo ============================================
echo.
echo   Service windows are now running.
echo   You can minimize them.
echo.
echo   URLs:
echo     Python Engine:  http://127.0.0.1:8082/health
echo     Go Backend:     http://127.0.0.1:8081/api/health
echo     Frontend:       http://localhost:5173
echo.
echo   To stop everything:
echo     Method 1: Run stop_services.bat
echo     Method 2: Close the 3 service windows
echo.
echo   Press any key to return to menu
echo ============================================
pause >nul
goto menu

:done_msg
echo.
echo ============================================
echo   Service started in a new window.
echo ============================================
pause >nul
goto menu

:end
cls
echo Exiting...
timeout /t 2 /nobreak >nul
exit /b 0