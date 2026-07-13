@echo off
title Go Backend (AIStudio) - DO NOT CLOSE
color 0B
cd /d "%~dp0apps\backend"
set AISTUDIO_ENV=development
set ENGINE_URL=http://127.0.0.1:8082
echo ============================================
echo   AIStudio Go Backend
echo   Starting on port 8081...
echo   ENGINE_URL=%ENGINE_URL%
echo ============================================
echo.
go run ./cmd/main.go
echo.
echo Backend exited with code %errorlevel%
pause