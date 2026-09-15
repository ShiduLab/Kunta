@echo off
cd /d "%~dp0"
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0PUBBLICA_KUNTA.ps1"
echo.
pause
