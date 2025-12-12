# XBoard Agent 部署指南

XBoard Agent 是一个轻量级客户端，运行在 sing-box 服务器上，自动从面板获取配置并管理 sing-box 进程。

## 快速开始

### 1. 在面板创建主机

1. 登录管理后台
2. 进入「主机管理」
3. 点击「添加主机」，输入名称
4. 复制显示的一键安装命令

### 2. 一键部署（推荐）

在服务器上执行面板显示的命令：

```bash
curl -sL https://raw.githubusercontent.com/ZYHUO/dashGO/main/agent/install.sh | bash -s -- https://your-panel.com YOUR_TOKEN
```

脚本会自动：
- 安装 sing-box
- 下载并安装 Agent
- 创建 systemd 服务
- 启动服务

### 3. 手动部署

```bash
# 下载 Agent
wget https://github.com/ZYHUO/dashGO/releases/download/1.1/dashgo-agent
chmod +x dashgo-agent
mv dashgo-agent /usr/local/bin/

# 运行
xboard-agent -panel https://your-panel.com -token YOUR_TOKEN
```

### 3. 添加节点

1. 在面板选择刚创建的主机
2. 点击「添加节点」
3. 选择协议类型，配置参数
4. Agent 会自动获取配置并重启 sing-box

## 命令行参数

```
-panel    面板地址（必填）
-token    主机 Token（必填）
-config   sing-box 配置文件路径（默认: /etc/sing-box/config.json）
-singbox  sing-box 可执行文件路径（默认: sing-box）
```

## Systemd 服务

创建 `/etc/systemd/system/xboard-agent.service`:

```ini
[Unit]
Description=XBoard Agent
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/xboard-agent -panel https://your-panel.com -token YOUR_TOKEN
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

启用服务：

```bash
systemctl daemon-reload
systemctl enable xboard-agent
systemctl start xboard-agent
```

## 支持的协议

| 协议 | 说明 |
|------|------|
| shadowsocks | Shadowsocks 2022 (推荐 2022-blake3-aes-128-gcm) |
| vless | VLESS + Reality (自动生成密钥) |
| trojan | Trojan (需配置 TLS 证书) |
| hysteria2 | Hysteria2 (需配置 TLS 证书) |

## 默认配置

### Shadowsocks 2022

```json
{
  "type": "shadowsocks",
  "listen_port": 8388,
  "method": "2022-blake3-aes-128-gcm"
}
```

### VLESS Reality

```json
{
  "type": "vless",
  "listen_port": 443,
  "flow": "xtls-rprx-vision",
  "tls": {
    "enabled": true,
    "server_name": "addons.mozilla.org",
    "reality": {
      "enabled": true,
      "handshake": {
        "server": "addons.mozilla.org",
        "server_port": 443
      }
    }
  }
}
```

## 工作流程

1. Agent 启动后连接面板获取配置
2. 生成 sing-box 配置文件并启动
3. 每 30 秒发送心跳
4. 每 60 秒检查配置更新
5. 配置变化时自动重启 sing-box

## 故障排查

### Agent 无法连接面板

- 检查面板地址是否正确
- 检查 Token 是否有效
- 检查网络连通性

### sing-box 启动失败

- 检查 sing-box 是否已安装
- 检查配置文件路径权限
- 查看 sing-box 日志

### 用户无法连接

- 确认节点已添加且显示
- 确认用户在正确的用户组
- 检查端口是否开放
