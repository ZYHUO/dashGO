# XBoard 编译指南

本文档说明如何编译 XBoard 的 Dashboard 和 Agent 组件。

## 预编译二进制文件

如果不想自己编译，可以直接下载预编译的二进制文件：

**下载地址**：`https://download.sharon.wiki/`

- **Server (Dashboard)**：`https://download.sharon.wiki/server/`
  - `xboard-server-linux-amd64`
  - `xboard-server-linux-arm64`
  - `xboard-server-windows-amd64.exe`
  - `xboard-server-darwin-amd64`
  - `xboard-server-darwin-arm64`
  - `migrate-linux-amd64`
  - `migrate-linux-arm64`

- **Agent**：`https://download.sharon.wiki/agent/`
  - `xboard-agent-linux-amd64`
  - `xboard-agent-linux-arm64`
  - `xboard-agent-linux-386`
  - `xboard-agent-windows-amd64.exe`
  - `xboard-agent-windows-386.exe`
  - `xboard-agent-darwin-amd64`
  - `xboard-agent-darwin-arm64`
  - `xboard-agent-freebsd-amd64`

**使用方法**：
```bash
# 下载 Server
wget https://download.sharon.wiki/server/xboard-server-linux-amd64
chmod +x xboard-server-linux-amd64

# 下载 Agent
wget https://download.sharon.wiki/agent/xboard-agent-linux-amd64
chmod +x xboard-agent-linux-amd64
```

---

## 前置要求

### 必需工具
- Go 1.19 或更高版本
- Node.js 16+ 和 npm (用于前端构建)
- Git

### Linux 环境
- Bash shell
- 标准 Unix 工具 (sha256sum, etc.)

### Windows 环境
- PowerShell 5.1 或更高版本

## 快速开始

### Linux/macOS 环境

```bash
# 赋予执行权限
chmod +x build-all.sh

# 构建所有组件
./build-all.sh all

# 或使用 make
make release
```

### Windows 环境

```powershell
# 构建所有组件
.\build-all.ps1 -Target all

# 指定版本号
.\build-all.ps1 -Target all -Version "1.0.0"
```

## 编译选项

### 使用 build-all.sh (Linux/macOS)

```bash
# 仅清理构建文件
./build-all.sh clean

# 仅构建前端
./build-all.sh frontend

# 仅构建 Server (Dashboard)
./build-all.sh server

# 仅构建 Agent (全架构)
./build-all.sh agent

# 构建所有组件
./build-all.sh all
```

### 使用 build-all.ps1 (Windows)

```powershell
# 仅清理构建文件
.\build-all.ps1 -Target clean

# 仅构建前端
.\build-all.ps1 -Target frontend

# 仅构建 Server (Dashboard)
.\build-all.ps1 -Target server

# 仅构建 Agent (全架构)
.\build-all.ps1 -Target agent

# 构建所有组件
.\build-all.ps1 -Target all
```

### 使用 Makefile

```bash
# 构建 Server (当前平台)
make build

# 构建 Server (所有平台)
make build-all

# 构建 Agent (当前平台)
make agent

# 构建 Agent (所有平台)
make agent-all

# 构建所有发布版本
make release

# 查看所有可用命令
make help
```

## 构建产物

编译完成后，所有二进制文件将位于 `dist/` 目录：

```
dist/
├── agent/
│   ├── xboard-agent-linux-amd64
│   ├── xboard-agent-linux-arm64
│   ├── xboard-agent-linux-386
│   ├── xboard-agent-windows-amd64.exe
│   ├── xboard-agent-windows-386.exe
│   ├── xboard-agent-darwin-amd64
│   ├── xboard-agent-darwin-arm64
│   ├── xboard-agent-freebsd-amd64
│   └── SHA256SUMS
├── server/
│   ├── xboard-server-linux-amd64
│   ├── xboard-server-linux-arm64
│   ├── xboard-server-windows-amd64.exe
│   ├── xboard-server-darwin-amd64
│   ├── xboard-server-darwin-arm64
│   ├── migrate-linux-amd64
│   ├── migrate-linux-arm64
│   └── SHA256SUMS
├── web/
│   └── dist/
└── VERSION.txt
```

## 支持的平台

### Server (Dashboard)
- Linux (amd64, arm64)
- Windows (amd64)
- macOS (amd64, arm64/Apple Silicon)

### Agent
- Linux (amd64, arm64, 386)
- Windows (amd64, 386)
- macOS (amd64, arm64/Apple Silicon)
- FreeBSD (amd64)

## 版本信息

编译时可以通过环境变量指定版本信息：

```bash
# Linux/macOS
VERSION=1.0.0 ./build-all.sh all

# Windows
.\build-all.ps1 -Target all -Version "1.0.0"
```

版本信息会被嵌入到二进制文件中，可以通过以下命令查看：

```bash
# Server
./xboard-server-linux-amd64 -version

# Agent
./xboard-agent-linux-amd64 -version
```

## 校验和验证

每个构建都会生成 SHA256 校验和文件 (`SHA256SUMS`)，用于验证文件完整性：

```bash
# Linux/macOS
cd dist/agent
sha256sum -c SHA256SUMS

# Windows (PowerShell)
cd dist\agent
Get-Content SHA256SUMS | ForEach-Object {
    $hash, $file = $_ -split '\s+', 2
    $computed = (Get-FileHash $file -Algorithm SHA256).Hash.ToLower()
    if ($hash -eq $computed) {
        Write-Host "OK: $file" -ForegroundColor Green
    } else {
        Write-Host "FAILED: $file" -ForegroundColor Red
    }
}
```

## 手动编译

如果需要手动编译特定平台：

### Server

```bash
# Linux amd64
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o xboard-server-linux-amd64 ./cmd/server

# Windows amd64
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o xboard-server-windows-amd64.exe ./cmd/server

# macOS arm64 (Apple Silicon)
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o xboard-server-darwin-arm64 ./cmd/server
```

### Agent

```bash
cd agent

# Linux amd64
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o xboard-agent-linux-amd64 .

# Windows amd64
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o xboard-agent-windows-amd64.exe .

# macOS arm64 (Apple Silicon)
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o xboard-agent-darwin-arm64 .
```

### 前端

```bash
cd web
npm install
npm run build
# 构建产物在 web/dist/ 目录
```

## 常见问题

### 1. 前端构建失败

确保已安装 Node.js 和 npm：

```bash
node --version
npm --version
```

如果 npm install 失败，尝试清理缓存：

```bash
cd web
rm -rf node_modules package-lock.json
npm cache clean --force
npm install
```

### 2. Go 模块下载失败

设置 Go 代理（中国大陆用户）：

```bash
go env -w GOPROXY=https://goproxy.cn,direct
```

### 3. 交叉编译失败

确保 CGO_ENABLED=0，因为交叉编译不支持 CGO：

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ...
```

### 4. Windows 上执行脚本被阻止

如果 PowerShell 执行策略阻止脚本运行：

```powershell
# 临时允许（当前会话）
Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass

# 或者直接运行
PowerShell -ExecutionPolicy Bypass -File .\build-all.ps1 -Target all
```

## 持续集成

可以在 CI/CD 环境中使用这些脚本：

### GitHub Actions 示例

```yaml
name: Build

on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Set up Node.js
        uses: actions/setup-node@v3
        with:
          node-version: '18'
      
      - name: Build all
        run: |
          chmod +x build-all.sh
          VERSION=${{ github.ref_name }} ./build-all.sh all
      
      - name: Upload artifacts
        uses: actions/upload-artifact@v3
        with:
          name: xboard-binaries
          path: dist/
```

## 开发构建

对于开发环境，可以使用更快的构建方式：

```bash
# 仅构建当前平台的 Server
make build

# 仅构建当前平台的 Agent
make agent

# 或直接运行
go run ./cmd/server -config configs/config.yaml
cd agent && go run . -config ../configs/config.yaml
```

## 相关文档

- [安装指南](README_SETUP.md)
- [Agent 自动更新](docs/agent-auto-update.md)
- [数据库迁移](docs/database-migration.md)
- [本地安装](docs/local-installation.md)

## 技术支持

如有问题，请查看：
- [GitHub Issues](https://github.com/ZYHUO/dashGO/issues)
- [文档目录](docs/)
