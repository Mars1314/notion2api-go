@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

echo.
echo ╔════════════════════════════════════════════╗
echo ║  notion-2api 停止脚本 (Windows)        ║
echo ╚════════════════════════════════════════════╝
echo.

REM 检查是否有进程在运行
tasklist /FI "IMAGENAME eq notion-2api.exe" 2>NUL | find /I /N "notion-2api.exe">NUL
if "%ERRORLEVEL%"=="0" (
    echo [*] 正在停止 notion-2api 服务...
    
    REM 获取进程 PID
    for /f "tokens=2" %%a in ('tasklist /FI "IMAGENAME eq notion-2api.exe" /NH') do (
        echo [*] 找到进程 PID: %%a
        set PID=%%a
        goto :kill
    )
    
    :kill
    REM 强制终止进程
    taskkill /F /IM notion-2api.exe >nul 2>&1
    
    if %ERRORLEVEL% EQU 0 (
        echo [√] 服务已成功停止
        echo [*] 进程 PID %PID% 已终止
    ) else (
        echo [!] 停止服务时出现错误
    )
) else (
    echo [*] 没有检测到运行中的 notion-2api 服务
)

echo.
pause