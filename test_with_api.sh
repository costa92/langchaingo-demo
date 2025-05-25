#!/bin/bash

# 测试脚本：使用真实 API 密钥测试翻译功能
# 使用方法：./test_with_api.sh YOUR_API_KEY

if [ $# -eq 0 ]; then
    echo "Usage: $0 <API_KEY>"
    echo "Example: $0 sk-your-api-key-here"
    exit 1
fi

API_KEY=$1

echo "🚀 Testing LangChain Go Agent with real API..."
echo "Using API Key: ${API_KEY:0:10}..."

# 设置环境变量并运行测试
export SILICONFLOW_API_KEY="$API_KEY"
export SILICONFLOW_API_URL="https://api.siliconflow.cn/v1"

echo ""
echo "📝 Running translation test..."
go run main.go

echo ""
echo "✅ Test completed!" 