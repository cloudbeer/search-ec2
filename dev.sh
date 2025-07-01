#!/bin/bash

# Search EC2 开发模式启动脚本

set -e

echo "🚀 启动 Search EC2 开发环境"
echo "=========================="

# 检查 Go 版本
echo "1. 检查 Go 环境..."
if ! command -v go &> /dev/null; then
    echo "❌ Go 未安装，请先安装 Go 1.21+"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo "✅ Go 版本: $GO_VERSION"

# 检查 Docker
echo "2. 检查 Docker 环境..."
if ! command -v docker &> /dev/null; then
    echo "❌ Docker 未安装，请先安装 Docker"
    exit 1
fi

if ! docker info &> /dev/null; then
    echo "❌ Docker 未运行，请启动 Docker"
    exit 1
fi

echo "✅ Docker 运行正常"

# 检查 Qdrant 容器
echo "3. 检查 Qdrant 数据库..."
if ! docker ps | grep -q qdrant; then
    echo "⚠️  Qdrant 容器未运行，正在启动..."
    docker run -d --name=qdrant --restart=always \
      -p 6333:6333 \
      -p 6334:6334 \
      -e QDRANT__SERVICE__API_KEY=Axyz.One234 \
      -v qdrant_storage:/qdrant/storage \
      qdrant/qdrant
    
    echo "⏳ 等待 Qdrant 启动..."
    sleep 5
else
    echo "✅ Qdrant 运行正常"
fi

# 检查环境变量文件
echo "4. 检查配置文件..."
if [ ! -f .env ]; then
    if [ -f .env.example ]; then
        echo "⚠️  .env 文件不存在，从 .env.example 复制..."
        cp .env.example .env
        echo "📝 请编辑 .env 文件配置您的 API 密钥"
    else
        echo "❌ .env.example 文件不存在"
        exit 1
    fi
else
    echo "✅ 配置文件存在"
fi

# 安装依赖
echo "5. 安装 Go 依赖..."
go mod download
echo "✅ 依赖安装完成"

# 检查开发工具
echo "6. 检查开发工具..."
if ! command -v air &> /dev/null; then
    echo "⚠️  Air 热重载工具未安装，正在安装..."
    go install github.com/cosmtrek/air@latest
fi

if command -v air &> /dev/null; then
    echo "✅ Air 热重载工具可用"
    USE_AIR=true
else
    echo "⚠️  Air 不可用，将使用普通模式"
    USE_AIR=false
fi

# 设置开发环境变量
export GIN_MODE=debug
export LOG_LEVEL=debug

echo ""
echo "🎉 开发环境准备完成！"
echo ""
echo "📋 服务信息:"
echo "   - API 服务: http://localhost:8080"
echo "   - Qdrant UI: http://localhost:6333/dashboard"
echo "   - 健康检查: http://localhost:8080/api/health"
echo ""
echo "🛠️  开发工具:"
echo "   - 热重载: $(if [ "$USE_AIR" = true ]; then echo "启用"; else echo "禁用"; fi)"
echo "   - 调试模式: 启用"
echo "   - 日志级别: debug"
echo ""

# 启动服务
if [ "$USE_AIR" = true ]; then
    echo "🔥 使用 Air 启动热重载开发服务器..."
    air
else
    echo "🚀 启动开发服务器..."
    go run cmd/server/main.go
fi
