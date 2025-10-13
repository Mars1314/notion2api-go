@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

echo.
echo ╔════════════════════════════════════════════╗
echo ║  notion-2api 一键启动脚本 (Windows)    ║
echo ╚════════════════════════════════════════════╝
echo.

REM 检查 .env 文件
if not exist ".env" (
    echo [错误] 未找到 .env 配置文件
    echo [提示] 请从 .env.example 复制并配置 .env 文件
    pause
    exit /b 1
)
echo [√] .env 配置文件检查通过

REM 检测是否已有进程在运行
tasklist /FI "IMAGENAME eq notion-2api.exe" 2>NUL | find /I /N "notion-2api.exe">NUL
if "%ERRORLEVEL%"=="0" (
    echo [!] 检测到已有进程在运行，正在重启...
    taskkill /F /IM notion-2api.exe >nul 2>&1
    timeout /t 2 /nobreak >nul
)

REM 检查 Go 环境
where go >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo [√] 检测到 Go 环境，使用本地方式启动
    
    REM 编译项目
    echo [*] 正在编译项目...
    go build -o notion-2api.exe .
    if %ERRORLEVEL% NEQ 0 (
        echo [错误] 编译失败
        pause
        exit /b 1
    )
    
    REM 启动服务
    echo [*] 启动服务...
    start /B notion-2api.exe > notion-2api.log 2>&1
    
    REM 等待服务启动
    timeout /t 3 /nobreak >nul
    
) else (
    echo [!] 未检测到 Go 环境
    
    REM 检查是否存在已编译的可执行文件
    if exist "notion-2api.exe" (
        echo [√] 使用已编译的可执行文件启动
        start /B notion-2api.exe > notion-2api.log 2>&1
        timeout /t 3 /nobreak >nul
    ) else (
        echo [错误] 未找到可执行文件，且无法编译
        echo [提示] 请安装 Go 环境或使用已编译的可执行文件
        pause
        exit /b 1
    )
)

REM 从 .env 读取端口配置
set PORT=8004
for /f "tokens=1,2 delims==" %%a in (.env) do (
    if "%%a"=="NGINX_PORT" set PORT=%%b
)

REM 测试服务
echo [*] 测试服务连接...
curl -s http://localhost:%PORT%/ >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo [√] 🎉 启动成功！
    echo.
    echo   服务地址: http://localhost:%PORT%
    echo   文档地址: http://localhost:%PORT%/docs
    echo   日志文件: type notion-2api.log
    echo   停止服务: stop.bat
    echo.
    
    REM 获取进程信息
    for /f "tokens=2" %%a in ('tasklist /FI "IMAGENAME eq notion-2api.exe" /NH') do (
        echo   进程 PID: %%a
        goto :found
    )
    :found
    
    echo.
    echo 📋 支持的模型：
    echo   [√] claude-sonnet-4.5  ^(推荐^)
    echo   [√] gpt-5
    echo   [√] claude-opus-4.1
    echo   [√] gpt-4.1
    echo   [!] gemini-2.5-flash   ^(可能不稳定^)
    echo   [!] gemini-2.5-pro     ^(可能不稳定^)
    echo.
    echo [*] 提示: 使用 stop.bat 可以停止服务
) else (
    echo [错误] 服务启动失败，请检查日志文件
    echo [提示] 使用命令: type notion-2api.log
)

echo.
pause