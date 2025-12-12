# 预编译二进制文件说明

## 下载地址

所有预编译二进制文件托管在：`https://download.sharon.wiki/`

### Server (Dashboard) 文件

**目录**：`https://download.sharon.wiki/server/`

| 文件名 | 平台 | 架构 | 说明 |
|--------|------|------|------|
| `xboard-server-linux-amd64` | Linux | x86_64 | Server 主程序 |
| `xboard-server-linux-arm64` | Linux | ARM64 | Server 主程序 |
| `xboard-server-windows-amd64.exe` | Windows | x86_64 | Server 主程序 |
| `xboard-server-darwin-amd64` | macOS | x86_64 | Server 主程序 |
| `xboard-server-darwin-arm64` | macOS | ARM64 (Apple Silicon) | Server 主程序 |
| `migrate-linux-amd64` | Linux | x86_64 | 数据库迁移工具 |
| `migrate-linux-arm64` | Linux | ARM64 | 数据库迁移工具 |

### Agent 文件

**目录**：`https://download.sharon.wiki/agent/`

| 文件名 | 平台 | 架构 | 说明 |
|--------|------|------|------|
| `xboard-agent-linux-amd64` | Linux | x86_64 | Agent 程序 |
| `xboard-agent-linux-arm64` | Linux | ARM64 | Agent 程序 |
| `xboard-agent-linux-386` | Linux | x86 (32位) | Agent 程序 |
| `xboard-agent-windows-amd64.exe` | Windows | x86_64 | Agent 程序 |
| `xboard-agent-windows-386.exe` | Windows | x86 (32位) | Agent 程序 |
| `xboard-agent-darwin-amd64` | macOS | x86_64 | Agent 程序 |
| `xboard-agent-darwin-arm64` | macOS | ARM64 (Apple Silicon) | Agent 程序 |
| `xboard-agent-freebsd-amd64` | FreeBSD | x86_64 | Agent 程序 |

## 使用方法

### 自动下载（推荐）

使用安装脚本会自动下载对应架构的二进制文件：

```bash
# Server 安装
bash setup.sh

# Agent 安装
curl -sL https://raw.githubusercontent.com/ZYHUO/dashGO/main/agent/install.sh | bash -s -- <面板地址> <Token>
```

### 手动下载

#### Linux Server

```bash
# 下载 Server (amd64)
wget https://download.sharon.wiki/server/xboard-server-linux-amd64
chmod +x xboard-server-linux-amd64

# 下载迁移工具
wget https://download.sharon.wiki/server/migrate-linux-amd64
chmod +x migrate-linux-amd64

# 运行
./xboard-server-linux-amd64 -config configs/config.yaml
```

#### Linux Agent

```bash
# 下载 Agent (amd64)
wget https://download.sharon.wiki/agent/xboard-agent-linux-amd64
chmod +x xboard-agent-linux-amd64

# 运行
./xboard-agent-linux-amd64 -config config.yaml
```

#### Windows Server

```powershell
# 使用 PowerShell 下载
Invoke-WebRequest -Uri "https://download.sharon.wiki/server/xboard-server-windows-amd64.exe" -OutFile "xboard-server.exe"

# 运行
.\xboard-server.exe -config configs\config.yaml
```

#### macOS Server (Apple Silicon)

```bash
# 下载
curl -L https://download.sharon.wiki/server/xboard-server-darwin-arm64 -o xboard-server
chmod +x xboard-server

# 运行
./xboard-server -config configs/config.yaml
```

## 文件校验

每个构建都会生成 SHA256 校验和文件：

```bash
# Server 校验和
wget https://download.sharon.wiki/server/SHA256SUMS
sha256sum -c SHA256SUMS

# Agent 校验和
wget https://download.sharon.wiki/agent/SHA256SUMS
sha256sum -c SHA256SUMS
```

## 版本信息

查看二进制文件的版本信息：

```bash
# Server
./xboard-server-linux-amd64 -version

# Agent
./xboard-agent-linux-amd64 -version
```

## 构建信息

这些二进制文件由以下脚本构建：

- **Linux/macOS**：`build-all.sh`
- **Windows**：`build-all.ps1`

构建命令：

```bash
# Linux/macOS
./build-all.sh all

# Windows
.\build-all.ps1 -Target all
```

详见 [编译指南](../BUILD.md)

## 更新频率

- 每次发布新版本时更新
- 重大 bug 修复时更新
- 安全更新时立即更新

## 支持的平台

### Server (Dashboard)

| 操作系统 | 架构 | 支持状态 |
|---------|------|---------|
| Linux | amd64 | ✅ 完全支持 |
| Linux | arm64 | ✅ 完全支持 |
| Windows | amd64 | ✅ 完全支持 |
| macOS | amd64 | ✅ 完全支持 |
| macOS | arm64 | ✅ 完全支持 |

### Agent

| 操作系统 | 架构 | 支持状态 |
|---------|------|---------|
| Linux | amd64 | ✅ 完全支持 |
| Linux | arm64 | ✅ 完全支持 |
| Linux | 386 | ✅ 完全支持 |
| Windows | amd64 | ✅ 完全支持 |
| Windows | 386 | ✅ 完全支持 |
| macOS | amd64 | ✅ 完全支持 |
| macOS | arm64 | ✅ 完全支持 |
| FreeBSD | amd64 | ✅ 完全支持 |

## 常见问题

### 1. 下载速度慢怎么办？

可以使用代理或镜像站点：

```bash
# 使用代理
export https_proxy=http://your-proxy:port
wget https://download.sharon.wiki/server/xboard-server-linux-amd64
```

### 2. 如何验证文件完整性？

使用 SHA256 校验：

```bash
# 下载校验和文件
wget https://download.sharon.wiki/server/SHA256SUMS

# 验证
sha256sum -c SHA256SUMS
```

### 3. 文件权限问题

Linux/macOS 下载后需要添加执行权限：

```bash
chmod +x xboard-server-linux-amd64
chmod +x xboard-agent-linux-amd64
```

### 4. Windows 安全警告

Windows 可能会提示"未知发布者"，这是正常的。可以：

1. 右键点击文件 → 属性
2. 勾选"解除锁定"
3. 点击"应用"

### 5. macOS 安全限制

macOS 可能会阻止运行未签名的应用：

```bash
# 移除隔离属性
xattr -d com.apple.quarantine xboard-server-darwin-arm64

# 或在系统设置中允许
# 系统偏好设置 → 安全性与隐私 → 通用 → 仍要打开
```

## 自行构建

如果不想使用预编译文件，可以自行构建：

```bash
# 克隆项目
git clone https://github.com/ZYHUO/dashGO.git
cd dashGO

# 构建所有平台
./build-all.sh all

# 或只构建当前平台
make build
make agent
```

详见 [编译指南](../BUILD.md)

## 相关文档

- [编译指南](../BUILD.md)
- [安装指南](../README_SETUP.md)
- [快速开始](../QUICK_START_SQLITE.md)
- [Agent 自动更新](agent-auto-update.md)

## 技术支持

如有问题，请提交 Issue：
https://github.com/ZYHUO/dashGO/issues
