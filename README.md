# kubectl-html - 高级 Kubernetes 资源查看器

一个强大的 kubectl 插件，支持以美观的 HTML 格式查看各种 Kubernetes 资源，包括 CRD（自定义资源定义）和复杂的多子资源结构。

## ✨ 主要特性

### 🎯 全面的资源支持
- **标准资源**: Pod, Deployment, Service, ConfigMap, Secret 等
- **CRD 支持**: 完全支持自定义资源定义
- **复杂结构**: 自动解析包含多个子资源的 YAML
- **多文档**: 支持 `---` 分隔的多文档 YAML

### 🎨 现代化界面
- **响应式设计**: 适配桌面和移动设备
- **单页概览**: 清爽的资源概览界面
- **模态框详情**: 点击资源卡片弹出 YAML 详情
- **实时状态**: 智能识别资源运行状态
- **统计信息**: 资源类型统计和命名空间计数

### 🔧 智能解析
- **状态检测**: 自动识别 Pod、Deployment 等资源状态
- **年龄计算**: 显示资源创建时间
- **命名空间**: 自动提取和统计命名空间信息
- **API 版本**: 显示资源的 API 版本信息

## 📦 安装

### 方法一：直接编译安装（推荐）

1. **前置要求**
   ```bash
   # 确保已安装 Go 1.21+ 和 kubectl
   go version
   kubectl version --client
   ```

2. **下载和编译**
   ```bash
   # 克隆项目
   git clone <repository-url>
   cd kubectl-html
   
   # 安装依赖
   go mod tidy
   
   # 编译程序
   go build -o kubectl-html main.go

   # 安装
   go install .
   ```

3. **安装为 kubectl 插件**

   **Linux/macOS:**
   ```bash
   # 方法1: 复制到 PATH 目录
   sudo cp kubectl-html /usr/local/bin/
   
   # 方法2: 创建符号链接
   sudo ln -s $(pwd)/kubectl-html /usr/local/bin/kubectl-html
   
   # 验证安装
   kubectl html --help
   ```

   **Windows:**
   ```cmd
   # 方法1: 复制到 PATH 目录 (需要管理员权限)
   copy kubectl-html.exe C:\Windows\System32\
   
   # 方法2: 添加当前目录到 PATH 环境变量
   # 在系统环境变量中添加当前目录路径
   
   # 验证安装
   kubectl html --help
   ```

### 方法二：作为 kubectl 插件安装

1. **重命名可执行文件**
   ```bash
   # Linux/macOS
   mv kubectl-html kubectl-html
   
   # Windows
   ren kubectl-html.exe kubectl-html.exe
   ```

2. **放置到 kubectl 插件目录**
   ```bash
   # 创建插件目录 (如果不存在)
   mkdir -p ~/.kube/plugins
   
   # 复制插件
   cp kubectl-html ~/.kube/plugins/
   chmod +x ~/.kube/plugins/kubectl-html
   ```

3. **使用插件**
   ```bash
   # 现在可以使用 kubectl html 命令
   kubectl html get pods
   kubectl html get deployments -A
   ```

### 方法三：一键安装脚本

**Linux/macOS 一键安装:**
```bash
#!/bin/bash
# install.sh

set -e

echo "🚀 安装 kubectl-html..."

# 检查依赖
if ! command -v go &> /dev/null; then
    echo "❌ 需要安装 Go 1.21+"
    exit 1
fi

if ! command -v kubectl &> /dev/null; then
    echo "❌ 需要安装 kubectl"
    exit 1
fi

# 编译
echo "📦 编译程序..."
go mod tidy
go build -o kubectl-html main.go

# 安装
echo "📋 安装到系统..."
sudo cp kubectl-html /usr/local/bin/
sudo chmod +x /usr/local/bin/kubectl-html

# 验证
echo "✅ 安装完成!"
echo "🎯 使用方法: kubectl html get pods"
kubectl html --help
```

**Windows PowerShell 一键安装:**
```powershell
# install.ps1

Write-Host "🚀 安装 kubectl-html..." -ForegroundColor Green

# 检查依赖
if (!(Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "❌ 需要安装 Go 1.21+" -ForegroundColor Red
    exit 1
}

if (!(Get-Command kubectl -ErrorAction SilentlyContinue)) {
    Write-Host "❌ 需要安装 kubectl" -ForegroundColor Red
    exit 1
}

# 编译
Write-Host "📦 编译程序..." -ForegroundColor Yellow
go mod tidy
go build -o kubectl-html.exe main.go

# 提示手动安装
Write-Host "✅ 编译完成!" -ForegroundColor Green
Write-Host "📋 请手动将 kubectl-html.exe 复制到 PATH 目录" -ForegroundColor Yellow
Write-Host "🎯 使用方法: kubectl html get pods" -ForegroundColor Cyan
```

## 🚀 使用方法

### 基本用法
```bash
# 查看 Pod
kubectl html get pods
kubectl html get po

# 查看指定命名空间的资源
kubectl html get pods -n kube-system

# 查看所有命名空间的 Deployment
kubectl html get deployments --all-namespaces
kubectl html get deploy -A

# 查看多种资源类型
kubectl html get po,svc,deploy

# 查看 CRD
kubectl html get crd

# 查看自定义资源
kubectl html get certificates.cert-manager.io
```

### 高级用法
```bash
# 使用标签选择器
kubectl html get pods -l app=nginx

# 查看特定资源
kubectl html get pod nginx-pod

# 查看集群级别资源
kubectl html get nodes

# 查看存储相关资源
kubectl html get pv,pvc,storageclass
```

## 🌐 Web 界面功能

### 📊 资源概览
- 资源类型统计卡片
- 资源列表网格视图
- 状态徽章显示
- 命名空间和年龄信息

### 📦 结构化详情模态框
- 点击任意资源卡片查看详情
- **结构化视图**: 美观的分组显示，包括：
  - 🔖 API 版本
  - 📦 资源类型  
  - 📋 元数据
  - ⚙️ 规格配置
  - 📊 状态信息
  - 💾 数据字段
- **YAML 源码**: 完整的原始 YAML 配置
- **全屏模式**: 点击 🔍 按钮或按 F11 放大到全窗口
- 支持键盘 ESC 关闭
- 点击外部区域关闭
- 可折叠的区域展示

## 🎨 支持的资源状态

### Pod 状态
- `Running` - 运行中 (绿色)
- `Pending` - 等待中 (黄色)
- `Failed` - 失败 (红色)
- `Unknown` - 未知 (灰色)

### Deployment/StatefulSet/DaemonSet 状态
- 基于 `conditions` 字段的 `Available` 条件
- 智能检测就绪状态

### CRD 和自定义资源状态
- 自动检测 `Ready`、`Available` 等条件
- 支持各种自定义状态字段

## 🔧 技术特性

- **零依赖部署**: 单个二进制文件
- **内存高效**: 流式处理大型 YAML
- **安全**: HTML 转义防止 XSS
- **快速**: 本地 HTTP 服务器
- **可扩展**: 易于添加新的资源类型支持

## 📱 响应式设计

- 桌面: 多列网格布局
- 平板: 自适应列数
- 手机: 单列堆叠布局
- 触摸友好的交互元素

## 🔄 实时功能

- 手动刷新按钮
- 可选的自动刷新 (注释掉的代码)
- API 端点支持 JSON 输出

## 🎯 使用场景

1. **开发调试**: 快速查看资源状态和配置
2. **运维监控**: 美观的资源概览界面
3. **文档生成**: 导出资源配置用于文档
4. **培训演示**: 直观展示 Kubernetes 资源结构
5. **故障排查**: 详细的资源信息和状态

## 📋 常用命令速查

| 资源类型 | 完整命令 | 简写命令 |
|---------|---------|---------|
| Pod | `kubectl html get pods` | `kubectl html get po` |
| Service | `kubectl html get services` | `kubectl html get svc` |
| Deployment | `kubectl html get deployments` | `kubectl html get deploy` |
| ConfigMap | `kubectl html get configmaps` | `kubectl html get cm` |
| Secret | `kubectl html get secrets` | - |
| Node | `kubectl html get nodes` | `kubectl html get no` |
| Namespace | `kubectl html get namespaces` | `kubectl html get ns` |
| Ingress | `kubectl html get ingresses` | `kubectl html get ing` |

## 🚨 注意事项

1. **性能考虑**: 避免在大集群中查询过多资源
2. **权限要求**: 需要对应的 kubectl 权限
3. **网络访问**: 确保能访问 Kubernetes API Server
4. **端口占用**: 默认使用 8000 端口，确保端口未被占用
5. **Windows 权限**: 避免将程序放在系统目录，推荐使用用户目录

## 🔧 故障排查

### ❌ "Access is denied" 错误

**问题**: 运行 kubectl html 时提示 "Access is denied"

**解决方案**:

1. **检查程序位置** (推荐)
   ```cmd
   # 不要放在系统目录，使用用户目录
   mkdir %USERPROFILE%\bin
   copy kubectl-html.exe %USERPROFILE%\bin\
   
   # 添加到 PATH (一次性设置)
   setx PATH "%PATH%;%USERPROFILE%\bin"
   ```

2. **使用便携模式**
   ```cmd
   # 直接在项目目录运行
   .\kubectl-html.exe get pods
   
   # 或创建批处理文件
   echo @echo off > kubectl-html.bat
   echo %~dp0kubectl-html.exe %* >> kubectl-html.bat
   ```

3. **检查 kubectl 配置权限**
   ```cmd
   # 检查 kubeconfig 文件权限
   dir %USERPROFILE%\.kube\config
   
   # 如果文件不存在或权限问题，重新配置
   kubectl config view
   ```

4. **PowerShell 执行策略**
   ```powershell
   # 检查执行策略
   Get-ExecutionPolicy
   
   # 如果受限，设置为当前用户
   Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
   ```

### 🔧 其他常见问题

如果遇到其他问题，请检查：

1. kubectl 是否正确安装和配置
2. 是否有访问集群的权限
3. 资源名称是否正确
4. 命名空间是否存在
5. 网络连接是否正常
6. 端口 8000 是否被占用
7. Windows 防火墙是否阻止了程序

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License