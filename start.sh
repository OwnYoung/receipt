#!/bin/bash

# 启动脚本
echo "🚀 启动收据生成服务..."

# 检查 Go 是否安装
if ! command -v go &> /dev/null; then
    echo "❌ Go 未安装，请先安装 Go 1.21 或更高版本"
    exit 1
fi

# 检查是否在项目根目录
if [ ! -f "go.mod" ]; then
    echo "❌ 请在项目根目录运行此脚本"
    exit 1
fi

# 检查 PDF 模板是否存在
if [ ! -f "templates/receipt_template.pdf" ]; then
    echo "⚠️  警告: 未找到 PDF 模板文件 templates/receipt_template.pdf"
    echo "请将您的 acroForm PDF 模板文件放置在 templates/receipt_template.pdf"
    echo "参考 templates/README.md 了解模板要求"
fi

# 创建输出目录
mkdir -p output

# 安装依赖
echo "📦 安装依赖..."
go mod tidy

# 启动服务
echo "🌟 启动服务在端口 8080..."
echo "访问 http://localhost:8080/health 检查服务状态"
echo "使用 Ctrl+C 停止服务"
echo ""

go run cmd/main.go