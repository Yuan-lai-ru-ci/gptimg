@echo off
chcp 65001 >nul
echo ============================================================
echo   生图模型上下文能力测试
echo ============================================================
echo.
cd /d "%~dp0"
python test_image_context.py
pause
