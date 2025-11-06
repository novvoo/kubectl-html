# 🌐 kubectl-html 网络访问配置指南

## 📋 监听地址选项

### 1️⃣ 本机访问 (默认)
```bash
# 仅本机可访问
kubectl-html get pods
kubectl-html -host localhost get pods
kubectl-html -host 127.0.0.1 get pods
```
- ✅ 最安全的选项
- ✅ 适合个人开发和调试
- ❌ 其他设备无法访问

### 2️⃣ 局域网访问
```bash
# 允许局域网内所有设备访问
kubectl-html -host 0.0.0.0 get pods
kubectl-html -host 0.0.0.0 -port 9000 get deployments -A
```
- ✅ 团队可以共享查看
- ✅ 适合演示和协作
- ⚠️ 需要注意网络安全

### 3️⃣ 指定网卡
```bash
# 绑定到特定网络接口
kubectl-html -host 192.168.1.100 get pods
kubectl-html -host 10.0.0.50 -port 8080 get services
```
- ✅ 精确控制访问接口
- ✅ 适合多网卡环境
- ⚠️ 需要知道具体IP地址

## 🎯 使用场景

### 个人开发
```bash
# 本机调试，最安全
kubectl-html get pods -n development
```

### 团队协作
```bash
# 团队成员可以通过局域网访问
kubectl-html -host 0.0.0.0 -port 8080 get pods --all-namespaces

# 团队成员访问: http://你的IP:8080
```

### 演示展示
```bash
# 会议室演示，投屏展示
kubectl-html -host 0.0.0.0 get pods,services,deployments -n production
```

### 远程访问
```bash
# 通过 SSH 隧道安全访问
ssh -L 8000:localhost:8000 user@remote-server
# 在远程服务器上运行: kubectl-html get pods
# 本地访问: http://localhost:8000
```

## 🔧 网络配置

### 查看本机IP地址

**Windows:**
```cmd
ipconfig
# 或
ipconfig | findstr IPv4
```

**Linux/macOS:**
```bash
ip addr show
# 或
ifconfig
# 或
hostname -I
```

### 防火墙配置

**Windows 防火墙:**
```cmd
# 允许特定端口
netsh advfirewall firewall add rule name="kubectl-html" dir=in action=allow protocol=TCP localport=8000

# 删除规则
netsh advfirewall firewall delete rule name="kubectl-html"
```

**Linux iptables:**
```bash
# 允许特定端口
sudo iptables -A INPUT -p tcp --dport 8000 -j ACCEPT

# 限制来源IP
sudo iptables -A INPUT -p tcp -s 192.168.1.0/24 --dport 8000 -j ACCEPT
```

**macOS:**
```bash
# 系统偏好设置 -> 安全性与隐私 -> 防火墙
# 或使用 pfctl 配置
```

## 🛡️ 安全最佳实践

### 1. 网络隔离
- 在可信的内网环境使用
- 避免在公网或不安全网络使用
- 使用 VPN 进行远程访问

### 2. 访问控制
```bash
# 使用 SSH 隧道
ssh -L 8000:localhost:8000 user@k8s-server
kubectl-html get pods  # 在远程服务器运行
# 本地访问 http://localhost:8000
```

### 3. 临时使用
```bash
# 使用完毕立即停止
kubectl-html get pods  # Ctrl+C 停止
```

### 4. 端口选择
```bash
# 使用非标准端口
kubectl-html -port 9876 get pods
```

## 📱 移动设备访问

### 手机/平板访问
1. 确保设备在同一局域网
2. 启动服务: `kubectl-html -host 0.0.0.0 get pods`
3. 手机浏览器访问: `http://电脑IP:8000`
4. 响应式界面自动适配移动端

### 二维码分享
```bash
# 生成访问链接的二维码 (需要安装 qrencode)
echo "http://$(hostname -I | awk '{print $1}'):8000" | qrencode -t UTF8
```

## 🔍 故障排查

### 无法访问问题
1. **检查监听地址**
   ```bash
   netstat -an | grep 8000  # Windows/Linux
   lsof -i :8000           # macOS/Linux
   ```

2. **检查防火墙**
   ```bash
   # 临时关闭防火墙测试
   # Windows: 控制面板 -> Windows Defender 防火墙
   # Linux: sudo ufw disable
   ```

3. **检查网络连通性**
   ```bash
   # 从其他设备测试
   telnet 192.168.1.100 8000
   # 或
   curl http://192.168.1.100:8000
   ```

### 性能优化
```bash
# 限制资源查询范围
kubectl-html -host 0.0.0.0 get pods -n specific-namespace

# 使用标签过滤
kubectl-html -host 0.0.0.0 get pods -l app=nginx
```

## 💡 高级技巧

### 反向代理
使用 nginx 或其他反向代理:
```nginx
server {
    listen 80;
    server_name k8s-dashboard.local;
    
    location / {
        proxy_pass http://localhost:8000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### Docker 部署
```dockerfile
FROM golang:alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o kubectl-html main.go

FROM alpine:latest
RUN apk add --no-cache kubectl
COPY --from=builder /app/kubectl-html /usr/local/bin/
EXPOSE 8000
ENTRYPOINT ["kubectl-html", "-host", "0.0.0.0"]
```

### 自动发现
```bash
# 使用 mDNS 广播服务 (需要 avahi-daemon)
kubectl-html -host 0.0.0.0 get pods &
avahi-publish -s "Kubernetes Dashboard" _http._tcp 8000
```