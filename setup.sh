#!/bin/bash
# Sage-RoundTable 快速启动脚本 (Linux/Mac)

echo "========================================"
echo "  Sage-RoundTable 快速配置向导"
echo "========================================"
echo ""

# 检查配置文件是否存在
if [ ! -f "config/models.yaml" ]; then
    echo "[1/3] 正在创建模型配置文件..."
    cp config/models.example.yaml config/models.yaml
    echo "✓ 已创建 config/models.yaml"
    echo ""
    echo "⚠️  请编辑 config/models.yaml 并启用你需要的模型"
    echo "   推荐启用以下模型之一："
    echo "   - qwen-max (阿里云通义千问)"
    echo "   - glm-4 (智谱AI)"
    echo "   - ollama-llama3 (本地运行，免费)"
    echo ""
    read -p "按回车键继续..."
else
    echo "[1/3] 模型配置文件已存在 ✓"
    echo ""
fi

# 检查 .env 文件是否存在
if [ ! -f ".env" ]; then
    echo "[2/3] 正在创建环境变量文件..."
    cp .env.example .env
    echo "✓ 已创建 .env"
    echo ""
    echo "⚠️  请编辑 .env 并填入你的 API Keys"
    echo "   获取 API Key 的地址："
    echo "   - 阿里云通义千问: https://dashscope.console.aliyun.com/apiKey"
    echo "   - 智谱AI: https://open.bigmodel.cn/usercenter/apikeys"
    echo "   - 百度文心一言: https://console.bce.baidu.com/qianfan/"
    echo ""
    read -p "按回车键继续..."
else
    echo "[2/3] 环境变量文件已存在 ✓"
    echo ""
fi

echo "[3/3] 准备启动程序..."
echo ""

# 启动程序
echo "正在启动 Sage-RoundTable..."
echo ""
go run cmd/cli/main.go
