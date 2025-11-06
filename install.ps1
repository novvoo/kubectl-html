# kubectl-html 一键安装脚本 (Windows PowerShell)

Write-Host "🚀 开始安装 kubectl-html..." -ForegroundColor Green

# 检查依赖
Write-Host "🔍 检查依赖..." -ForegroundColor Yellow

if (!(Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "❌ 错误: 需要安装 Go 1.21+" -ForegroundColor Red
    Write-Host "📋 请访问 https://golang.org/dl/ 下载安装" -ForegroundColor Cyan
    exit 1
}

if (!(Get-Command kubectl -ErrorAction SilentlyContinue)) {
    Write-Host "❌ 错误: 需要安装 kubectl" -ForegroundColor Red
    Write-Host "📋 请访问 https://kubernetes.io/docs/tasks/tools/ 查看安装说明" -ForegroundColor Cyan
    exit 1
}

Write-Host "✅ 依赖检查通过" -ForegroundColor Green

# 检查 Go 版本
$goVersion = (go version).Split()[2].Replace("go", "")
Write-Host "📋 Go 版本: $goVersion" -ForegroundColor Cyan

# 编译程序
Write-Host "📦 编译 kubectl-html..." -ForegroundColor Yellow
go mod tidy
go build -o kubectl-html.exe main.go

if (!(Test-Path "kubectl-html.exe")) {
    Write-Host "❌ 编译失败" -ForegroundColor Red
    exit 1
}

Write-Host "✅ 编译成功" -ForegroundColor Green

# 安装提示
Write-Host "📋 安装说明:" -ForegroundColor Yellow
Write-Host ""
Write-Host "方法1: 复制到系统目录 (需要管理员权限)" -ForegroundColor Cyan
Write-Host "   copy kubectl-html.exe C:\Windows\System32\" -ForegroundColor White
Write-Host ""
Write-Host "方法2: 添加到 PATH 环境变量" -ForegroundColor Cyan
Write-Host "   1. 记住当前目录: $PWD" -ForegroundColor White
Write-Host "   2. 打开系统属性 -> 高级 -> 环境变量" -ForegroundColor White
Write-Host "   3. 在 PATH 中添加当前目录路径" -ForegroundColor White
Write-Host ""
Write-Host "方法3: 使用 PowerShell 配置文件 (推荐)" -ForegroundColor Cyan
Write-Host "   Set-Alias kubectl-html '$PWD\kubectl-html.exe'" -ForegroundColor White
Write-Host ""

# 尝试自动添加别名到当前会话
Set-Alias kubectl-html "$PWD\kubectl-html.exe" -Scope Global

Write-Host "✅ 已为当前 PowerShell 会话创建别名" -ForegroundColor Green
Write-Host ""
Write-Host "🎯 使用方法:" -ForegroundColor Yellow
Write-Host "   kubectl html get pods" -ForegroundColor White
Write-Host "   kubectl html get deployments -A" -ForegroundColor White
Write-Host "   kubectl html get po,svc,deploy -n kube-system" -ForegroundColor White
Write-Host ""
Write-Host "🌐 Web 界面将在 http://localhost:8000 启动" -ForegroundColor Cyan
Write-Host ""

# 显示安装信息
Write-Host "📋 安装信息:" -ForegroundColor Yellow
Write-Host "   程序位置: $PWD\kubectl-html.exe" -ForegroundColor White
Write-Host "   Go 版本: $goVersion" -ForegroundColor White

try {
    $kubectlVersion = (kubectl version --client --short 2>$null)
    Write-Host "   kubectl 版本: $kubectlVersion" -ForegroundColor White
} catch {
    Write-Host "   kubectl 版本: 未知" -ForegroundColor White
}

Write-Host ""
Write-Host "🎉 安装完成! 现在可以使用 'kubectl html' 命令了" -ForegroundColor Green
Write-Host ""
Write-Host "💡 提示: 要在新的 PowerShell 会话中使用，请按照上述方法2或3进行配置" -ForegroundColor Yellow