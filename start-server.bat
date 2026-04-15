@echo off
REM Sage-RoundTable 服务器快速启动脚本 (Windows)

echo ========================================
echo   Sage-RoundTable API Server
echo ========================================
echo.

REM 检查 .env 文件
if not exist .env (
    echo [WARN] .env file not found. Creating from .env.example...
    copy .env.example .env
    echo [INFO] Please edit .env file and set your configuration.
    pause
    exit /b 1
)

REM 创建数据目录
if not exist data (
    echo [INFO] Creating data directory...
    mkdir data
)

REM 检查依赖
echo [INFO] Checking dependencies...
go mod tidy >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Failed to install dependencies. Please run: go mod tidy
    pause
    exit /b 1
)

echo [INFO] Starting server on port 8080...
echo [INFO] Press Ctrl+C to stop the server.
echo.

REM 启动服务器
go run cmd/server/main.go

pause
