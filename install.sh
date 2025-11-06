#!/bin/bash
# kubectl-html 一键安装脚本 (Linux/macOS)

set -e

echo "🚀 开始安装 kubectl-html..."

# 检查依赖
echo "🔍 检查依赖..."

if ! command -v go &> /dev/null; then
    echo "❌ 错误: 需要安装 Go 1.21+"
    echo "📋 请访问 https://golang.org/dl/ 下载安装"
    exit 1
fi

if ! command -v kubectl &> /dev/null; then
    echo "❌ 错误: 需要安装 kubectl"
    echo "📋 请访问 https://kubernetes.io/docs/tasks/tools/ 查看安装说明"
    exit 1
fi

echo "✅ 依赖检查通过"

# 检查 Go 版本
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
REQUIRED_VERSION="1.21"

if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
    echo "⚠️  警告: Go 版本 $GO_VERSION 可能不兼容，建议使用 Go 1.21+"
fi

# 编译程序
echo "📦 编译 kubectl-html..."
go mod tidy
go build -o kubectl-html main.go

if [ ! -f "kubectl-html" ]; then
    echo "❌ 编译失败"
    exit 1
fi

echo "✅ 编译成功"

# 安装程序
echo "📋 安装程序到系统..."

# 检查是否有 sudo 权限
if [ "$EUID" -eq 0 ]; then
    # 以 root 身份运行
    cp kubectl-html /usr/local/bin/
    chmod +x /usr/local/bin/kubectl-html
else
    # 需要 sudo
    if command -v sudo &> /dev/null; then
        sudo cp kubectl-html /usr/local/bin/
        sudo chmod +x /usr/local/bin/kubectl-html
    else
        echo "❌ 需要 sudo 权限来安装到 /usr/local/bin/"
        echo "📋 请手动复制 kubectl-html 到 PATH 目录"
        echo "   例如: sudo cp kubectl-html /usr/local/bin/"
        exit 1
    fi
fi

# 验证安装
echo "🔍 验证安装..."

if command -v kubectl-html &> /dev/null; then
    echo "✅ kubectl-html 安装成功!"
    echo ""
    echo "🎯 使用方法:"
    echo "   kubectl html get pods"
    echo "   kubectl html get deployments -A"
    echo "   kubectl html get po,svc,deploy -n kube-system"
    echo ""
    echo "🌐 Web 界面将在 http://localhost:8000 启动"
    echo ""
    
    # 显示版本信息
    echo "📋 安装信息:"
    echo "   程序位置: $(which kubectl-html)"
    echo "   Go 版本: $GO_VERSION"
    echo "   kubectl 版本: $(kubectl version --client --short 2>/dev/null || echo "未知")"
    
else
    echo "❌ 安装验证失败"
    echo "📋 请检查 /usr/local/bin 是否在 PATH 中"
    exit 1
fi

echo ""
echo "🎉 安装完成! 现在可以使用 'kubectl html' 命令了"