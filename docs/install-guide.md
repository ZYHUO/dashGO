# XBoard 一键安装指南

## 快速开始

### 一键安装面板

```bash
# 下载并运行安装脚本
curl -sL https://raw.githubusercontent.com/ZYHUO/dashGO/main/install.sh | bash -s -- panel
```

### 一键安装节点 (Agent)

```bash
# 在节点服务器上运行
curl -sL https://raw.githubusercontent.com/ZYHUO/dashGO/main/install.sh | bash -s -- agent <面板地址> <Token>

# 示例
curl -sL https://raw.githubusercontent.com/ZYHUO/dashGO/main/install.sh | bash -s -- agent https://panel.example.com abc123def456
```

### 完整安装 (面板 + 节点)

```bash
curl -sL https://raw.githubusercontent.com/ZYHUO/dashGO/main/install.sh | bash -s -- all
```

### 交互式安装

```bash
# 下载脚本
wget -O install.sh https://raw.githubusercontent.com/ZYHUO/dashGO/main/install.sh
chmod +x install.sh

# 运行交互式菜单
./install.sh
```

## 安装选项

| 命令 | 说明 |
|------|------|
| `panel` | 安装面板 |
| `agent [url] [token]` | 安装节点 |
| `all` | 完整安装 |
| `uninstall-panel` | 卸载面板 |
| `uninstall-agent` | 卸载节点 |
| `update-panel` | 更新面板 |
| `update-agent` | 更新节点 |

## 系统要求

### 面板服务器
- 操作系统: Ubuntu 20.04+, Debian 10+, CentOS 7+
- CPU: 1 核心以上
- 内存: 1GB 以上
- 硬盘: 10GB 以上
- 需要安装 Docker

### 节点服务器
- 操作系统: Ubuntu 20.04+, Debian 10+, CentOS 7+, Alpine
- CPU: 1 核心以上
- 内存: 512MB 以上
- 硬盘: 5GB 以上

## 安装后配置

### 面板

安装完成后，访问:
- 前台: `http://YOUR_IP:80`
- 后台: `http://YOUR_IP:80/admin`

默认管理员账户:
- 邮箱: `admin@xboard.local`
- 密码: `admin123`

**请立即修改默认密码！**

### 配置文件

面板配置文件位于: `/opt/xboard/config.yaml`

主要配置项:
```yaml
app:
  name: "XBoard"
  url: "https://your-domain.com"  # 修改为你的域名
  
mail:
  host: "smtp.gmail.com"
  username: "your-email@gmail.com"
  password: "your-app-password"
  
telegram:
  bot_token: "your-bot-token"
```

### 配置 HTTPS

1. 准备 SSL 证书
2. 将证书放到 `/opt/xboard/ssl/` 目录
3. 编辑 `/opt/xboard/nginx.conf`，取消 HTTPS 配置的注释
4. 重启服务: `cd /opt/xboard && docker compose restart nginx`

## 常用命令

### 面板管理

```bash
# 进入安装目录
cd /opt/xboard

# 查看服务状态
docker compose ps

# 查看日志
docker compose logs -f

# 重启服务
docker compose restart

# 停止服务
docker compose down

# 启动服务
docker compose up -d
```

### 节点管理

```bash
# 查看 Agent 状态
systemctl status xboard-agent

# 查看 Agent 日志
journalctl -u xboard-agent -f

# 重启 Agent
systemctl restart xboard-agent

# 查看 sing-box 状态
systemctl status sing-box

# 查看 sing-box 日志
journalctl -u sing-box -f
```

## 添加节点

1. 登录面板后台
2. 进入「主机管理」
3. 点击「添加主机」
4. 填写主机信息，获取 Token
5. 在节点服务器上运行安装命令

## 故障排除

### 面板无法访问

```bash
# 检查服务状态
cd /opt/xboard && docker compose ps

# 检查端口
netstat -tlnp | grep -E '80|8080'

# 检查防火墙
ufw status
firewall-cmd --list-all
```

### 节点无法连接

```bash
# 检查 Agent 状态
systemctl status xboard-agent

# 检查网络连接
curl -v https://your-panel.com/api/agent/config

# 检查 Token 是否正确
cat /etc/systemd/system/xboard-agent.service
```

### 数据库连接失败

```bash
# 检查 MySQL 容器
docker logs xboard-mysql

# 进入 MySQL
docker exec -it xboard-mysql mysql -u xboard -p
```

## 更新

### 更新面板

```bash
bash /opt/dashgo/install.sh update-panel
# 或
curl -sL https://raw.githubusercontent.com/ZYHUO/dashGO/main/install.sh | bash -s -- update-panel
```

### 更新节点

```bash
bash install.sh update-agent
# 或
curl -sL https://raw.githubusercontent.com/ZYHUO/dashGO/main/install.sh | bash -s -- update-agent
```

## 卸载

### 卸载面板

```bash
bash install.sh uninstall-panel
```

### 卸载节点

```bash
bash install.sh uninstall-agent
```

## 支持

如有问题，请提交 Issue: https://github.com/ZYHUO/dashGO/issues
