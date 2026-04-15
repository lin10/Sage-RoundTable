#!/bin/bash
# Sage-RoundTable 服务器快速启动脚本 (Linux/Mac)

echo "========================================"
echo "  Sage-RoundTable API Server"
echo "========================================"
echo ""

# 检查 .env 文件
if [ ! -f .env ]; then
    echo "[WARN] .env file not found. Creating from .env.example..."
    cp .env.example .env
    echo "[INFO] Please edit .env file and set your configuration."
    exit 1
fi

# 创建数据目录
if [ ! -d data ]; then
    echo "[INFO] Creating data directory..."
    mkdir -p data
fi

# 检查依赖
echo "[INFO] Checking dependencies..."
go mod tidy > /dev/null 2>&1
if [ $? -ne 0 ]; then
    echo "[ERROR] Failed to install dependencies. Please run: go mod tidy"
    exit 1
fi

echo "[INFO] Starting server on port 8080..."
echo "[INFO] Press Ctrl+C to stop the server."
echo ""

# 启动服务器
go run cmd/server/main.go
