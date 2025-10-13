@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

echo.
echo ======================================
echo notion-2api-go API 测试脚本 (Windows)
echo ======================================
echo.

REM 检查 curl 是否可用
where curl >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo [错误] 未找到 curl 命令
    echo [提示] 请安装 curl 或使用 Windows 10/11 自带的 curl
    pause
    exit /b 1
)

REM 从 .env 读取配置
set API_URL=http://localhost:8004
set API_KEY=1

if exist ".env" (
    for /f "usebackq tokens=1,2 delims==" %%a in (".env") do (
        if "%%a"=="NGINX_PORT" set API_URL=http://localhost:%%b
        if "%%a"=="API_MASTER_KEY" set API_KEY=%%b
    )
)

echo API URL: %API_URL%
echo.

REM 测试 1: 健康检查
echo [测试 1] 健康检查
echo GET %API_URL%/
curl -s %API_URL%/
if %ERRORLEVEL% EQU 0 (
    echo.
    echo [√] 服务运行正常
) else (
    echo [×] 健康检查失败
)
echo.
echo.

REM 测试 2: 获取模型列表
echo [测试 2] 获取模型列表
echo GET %API_URL%/v1/models
curl -s -H "Authorization: Bearer %API_KEY%" %API_URL%/v1/models
if %ERRORLEVEL% EQU 0 (
    echo.
    echo [√] 成功获取模型列表
) else (
    echo [×] 获取模型列表失败
)
echo.
echo.

REM 测试 3: 文档页面
echo [测试 3] 文档页面
echo GET %API_URL%/docs
curl -s %API_URL%/docs | findstr /C:"notion-2api-go" >nul
if %ERRORLEVEL% EQU 0 (
    echo [√] 文档页面访问正常
    echo 访问地址: %API_URL%/docs
) else (
    echo [×] 文档页面访问失败
)
echo.
echo.

echo [开始测试所有模型...]
echo.

set success_count=0
set fail_count=0

REM 定义模型数组
set models[0]=claude-sonnet-4.5:推荐
set models[1]=gpt-5:高级
set models[2]=claude-opus-4.1:复杂
set models[3]=gpt-4.1:快速
set models[4]=gemini-2.5-flash:快速
set models[5]=gemini-2.5-pro:高质量

for /L %%i in (0,1,5) do (
    set "model_info=!models[%%i]!"
    
    for /f "tokens=1,2 delims=:" %%a in ("!model_info!") do (
        set model=%%a
        set desc=%%b
        
        echo [测试模型] !model! ^(!desc!^)
        
        REM 创建临时 JSON 文件
        echo {"model":"!model!","messages":[{"role":"user","content":"你好，请用一句话介绍你自己"}],"stream":true} > temp_request.json
        
        REM 发送请求
        curl -s -N -X POST %API_URL%/v1/chat/completions ^
            -H "Authorization: Bearer %API_KEY%" ^
            -H "Content-Type: application/json" ^
            -d @temp_request.json > temp_response.txt
        
        REM 检查响应
        findstr /C:"error" temp_response.txt >nul
        if !ERRORLEVEL! EQU 0 (
            echo [×] !model! - 失败
            type temp_response.txt | findstr /C:"message"
            set /a fail_count+=1
        ) else (
            findstr /C:"content" temp_response.txt >nul
            if !ERRORLEVEL! EQU 0 (
                echo [√] !model! - 成功
                set /a success_count+=1
            ) else (
                echo [×] !model! - 无响应
                set /a fail_count+=1
            )
        )
        echo.
        
        REM 短暂延迟
        timeout /t 1 /nobreak >nul
    )
)

REM 清理临时文件
del temp_request.json 2>nul
del temp_response.txt 2>nul

echo ======================================
echo [测试完成]
echo ======================================
echo.
echo 测试统计：
echo   成功: %success_count%/6
echo   失败: %fail_count%/6
set /a success_rate=success_count*100/6
echo   成功率: %success_rate%%%
echo.
echo 提示：
echo - 如果测试失败，请检查 .env 配置是否正确
echo - 查看详细日志: type notion-2api.log
echo - 访问文档: %API_URL%/docs
echo - 确保 Notion 凭证有效且未过期
echo.

pause