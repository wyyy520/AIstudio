@echo off
title Python Engine (AIStudio) - DO NOT CLOSE
color 0A
cd /d "%~dp0Engine"
echo ============================================
echo   AIStudio Python Engine 
echo   Starting on port 8082...
echo ============================================
echo.
python server.py --port 8082
echo.
echo Engine exited with code %errorlevel%
pause