# kubectl-html Makefile

.PHONY: build install clean test help

# 默认目标
help:
	@echo "kubectl-html 构建工具"
	@echo ""
	@echo "可用命令:"
	@echo "  build    - 编译程序"
	@echo "  install  - 安装到系统"
	@echo "  clean    - 清理构建文件"
	@echo "  test     - 运行测试"
	@echo "  help     - 显示此帮助信息"

# 编译程序
build:
	@echo "📦 编译 kubectl-html..."
	go mod tidy
	go build -o kubectl-html main.go
	@echo "✅ 编译完成"

# 安装程序
install: build
	@echo "📋 安装 kubectl-html..."
ifeq ($(OS),Windows_NT)
	@echo "Windows 系统请手动运行 install.ps1"
else
	chmod +x install.sh
	./install.sh
endif

# 清理构建文件
clean:
	@echo "🧹 清理构建文件..."
	rm -f kubectl-html kubectl-html.exe
	@echo "✅ 清理完成"

# 运行测试
test:
	@echo "🧪 运行测试..."
	go test -v ./...
	@echo "✅ 测试完成"

# 快速开始
quick: build
	@echo "🚀 快速启动 kubectl-html..."
	@echo "💡 示例: ./kubectl-html get pods"
	@echo "🌐 Web界面: http://localhost:8000"

# 检查依赖
check:
	@echo "🔍 检查依赖..."
	@which go > /dev/null || (echo "❌ 需要安装 Go" && exit 1)
	@which kubectl > /dev/null || (echo "❌ 需要安装 kubectl" && exit 1)
	@echo "✅ 依赖检查通过"