@echo off
chcp 936 >nul
title AIStudio Stop Services
color 0C

echo ============================================
echo   Stop AIStudio Services
echo ============================================
echo.

echo [1/3] Stopping Python Engine...
taskkill /f /fi "WINDOWTITLE eq Python Engine*" 2>nul
taskkill /f /im python.exe 2>nul
echo   Done.

echo [2/3] Stopping Go Backend...
taskkill /f /fi "WINDOWTITLE eq Go Backend*" 2>nul
taskkill /f /im go.exe 2>nul
echo   Done.

echo [3/3] Stopping Frontend...
taskkill /f /fi "WINDOWTITLE eq Frontend*" 2>nul
taskkill /f /im node.exe 2>nul
echo   Done.

echo.
echo ============================================
echo   All services stopped
echo ============================================
echo.
timeout /t 3 /nobreak >nul