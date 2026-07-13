@echo off
title Frontend (AIStudio) - DO NOT CLOSE
color 0D
cd /d "%~dp0apps\desktop"
echo ============================================
echo   AIStudio Frontend
echo   Starting on port 5173...
echo ============================================
echo.
if not exist "node_modules\" (
    echo Installing npm dependencies...
    call npm install
    echo.
)
npm run dev
echo.
echo Frontend exited with code %errorlevel%
pause