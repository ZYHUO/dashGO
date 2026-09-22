# dashGO

一个现代化的代理面板管理系统，基于 Go + Vue 3 构建。

> **演示站点**：[misaka.cfd](https://misaka.cfd/) — 出于安全考虑，演示凭据不再公开。
> 想试用的话，自建最快：下面一条命令两分钟就起来了。

## 特性

- 🚀 **高性能**：Go 后端，Vue 3 前端
- 💾 **灵活数据库**：支持 SQLite（默认）和 MySQL
- 🔒 **安全可靠**：JWT 认证、权限控制、SQL 注入防护
- 🐳 **容器化部署**：Docker Compose 一键部署
- 📊 **流量统计**：实时流量监控和统计
- 🎨 **现代 UI**：响应式设计，支持移动端

## 这是什么，不是什么

**dashGO 是一个自托管的代理面板**：Go 后端 + Vue 3 前端，底层 sing-box。
你在自己的一台 Linux VPS 上跑起来，用它管节点、管用户、看流量。

### 适合你，如果

- 你有一台 Linux VPS，且愿意自己维护
- 你想**自托管**，不想把用户数据交给第三方面板
- 你能接受用 Docker 部署
- 你缺的是一个面板，不是一整套生态

### 不适合你，如果

- **你没有 Linux 服务器** —— 目前只支持 Linux（Ubuntu/Debian/CentOS），没有 Windows/macOS 服务端
- **你跑 arm64 又不想装 MySQL** —— SQLite 模式不支持 linux arm64
- **你不想碰 Docker** —— 部署依赖 Docker + Docker Compose，没有免 Docker 的二进制包

### 先说清楚的局限

| 项 | 现状 |
|---|---|
| 部署方式 | `install.sh` 一键脚本 / Docker Compose；**没有**免 Docker 安装包 |
| 数据库 | SQLite（默认）或 MySQL；**SQLite 不支持 linux arm64** |
| 服务端平台 | Linux only |
| 最低配置 | 512MB 内存 / 1GB 磁盘，推荐 1GB+ |
| 界面截图 | 暂无。装完打开就能看到，或 `bash install.sh` 两分钟自己验证 |
| 许可证 | MIT —— 可自改、可分发、可商用 |

### 怎么判断要不要用它

别看宣传，问自己四个问题：

1. **愿不愿意用 Docker？** 不愿意 → 跳过，没有别的安装方式
2. **要不要 SQLite 的轻量？** 要 → 确认自己不是 arm64（arm64 必须 MySQL）
3. **要不要自托管？** 要 → 方向对；不想维护服务器 → 这类面板都不适合你
4. **要不要能改？** MIT，随便改随便分发

**都没有问题 → 直接往下走，两分钟装完。**

## 快速开始

### 安装

```bash
curl -sSL https://raw.githubusercontent.com/ZYHUO/dashGO/refs/heads/main/install.sh -o install.sh && bash install.sh
```

### 安装选项

安装时会提示选择：

1. **数据库类型**
   - SQLite（推荐，轻量级，linux arm64不支持，请不要选择此项）
   - MySQL（外部数据库）

2. **安装方式**
   - 预编译版本（推荐，快速）
   - 源码构建（支持自定义）

3. **HTTPS 配置**
   - 启用 HTTPS（443 端口）
   - 证书类型：
     - Cloudflare Origin Certificate（推荐）
     - 自签名证书（测试用）
     - 自有证书

4. **自定义配置**
   - Web 访问端口（HTTP: 80 / HTTPS: 443）
   - 管理员邮箱和密码


## 系统要求

- **操作系统**：Linux（Ubuntu/Debian/CentOS）
- **内存**：最低 512MB，推荐 1GB+
- **磁盘**：最低 1GB 可用空间
- **软件**：Docker 和 Docker Compose

## 文档

- [构建指南](BUILD.md) - 如何从源码构建
- [故障排除](#faq) - 常见问题解决（部署打不开、端口、架构不匹配等）
- [安全指南](SECURITY.md) - 安全配置和最佳实践
- [更新日志](CHANGELOG.md) - 版本更新记录

## CDN 和域名配置

### 使用 Cloudflare CDN

**1. 启用 HTTPS（推荐）**

安装时选择启用 HTTPS，使用 Cloudflare Origin Certificate：

```bash
bash install.sh panel
# 选择 "是否启用 HTTPS" → y
# 选择证书类型 → 1 (Cloudflare Origin Certificate)
```

获取 Cloudflare Origin Certificate：
1. 登录 Cloudflare 控制台
2. 选择你的域名
3. 进入 **SSL/TLS** → **Origin Server**
4. 点击 **Create Certificate**
5. 复制证书和私钥，粘贴到安装脚本提示中

Cloudflare SSL/TLS 设置：
- 加密模式：**Full (strict)**
- 最低 TLS 版本：TLS 1.2

**2. 使用 HTTP（不推荐）**

如果不想配置 HTTPS，可以：
- Cloudflare SSL/TLS 模式设为 **Flexible**
- 或关闭 Cloudflare 代理（DNS 记录的橙色云图标改为灰色）

### 不使用 CDN

直接通过 IP 或域名访问：
- HTTP: `http://your-ip:80`
- HTTPS: `https://your-domain.com:443`（需配置证书）

## 管理命令

```bash
# 查看服务状态
cd /opt/dashgo && docker compose ps

# 查看日志
docker compose logs -f

# 查看特定服务日志
docker compose logs -f dashgo
docker compose logs -f nginx

# 重启服务
docker compose restart

# 重启特定服务
docker compose restart dashgo
docker compose restart nginx

# 停止服务
docker compose down

# 更新服务
docker compose pull && docker compose up -d --build
```

## 数据库

### SQLite（默认）

- **优点**：轻量级，无需额外配置
- **数据文件**：`/opt/dashgo/data/dashgo.db`
- **适用场景**：中小规模部署（< 10,000 用户）

### MySQL（可选）

- **优点**：高并发性能更好
- **配置**：安装时输入外部 MySQL 连接信息
- **适用场景**：大规模生产环境

### 切换数据库

```bash
cd /opt/dashgo
./use-sqlite.sh  # 切换到 SQLite
```

## 开发

### 本地开发

```bash
# 后端
go run cmd/server/main.go -config configs/config.yaml

# 前端
cd web
npm install
npm run dev
```

### 构建

```bash
# 构建所有组件
./build-all.sh all

# 仅构建前端
./build-all.sh frontend

# 仅构建后端
./build-all.sh server

# 使用 Docker 构建（支持交叉编译）
./build-all.sh server-docker
```

## 架构

```
dashGO/
├── cmd/              # 主程序入口
│   ├── server/       # 面板服务
│   └── migrate/      # 数据库迁移工具
├── internal/         # 内部包
│   ├── config/       # 配置管理
│   ├── handler/      # HTTP 处理器
│   ├── middleware/   # 中间件
│   ├── model/        # 数据模型
│   ├── repository/   # 数据访问层
│   └── service/      # 业务逻辑层
├── pkg/              # 公共包
│   ├── cache/        # 缓存
│   ├── database/     # 数据库
│   └── utils/        # 工具函数
├── web/              # 前端（Vue 3）
│   ├── src/
│   │   ├── api/      # API 调用
│   │   ├── views/    # 页面组件
│   │   ├── router/   # 路由配置
│   │   └── stores/   # 状态管理
│   └── dist/         # 构建产物
├── agent/            # 节点代理
├── configs/          # 配置文件
├── migrations/       # 数据库迁移
└── docs/             # 文档
```

## 技术栈

### 后端
- **语言**：Go 1.22+
- **框架**：Gin
- **ORM**：GORM
- **缓存**：Redis
- **认证**：JWT

### 前端
- **框架**：Vue 3
- **构建**：Vite
- **UI**：Tailwind CSS
- **路由**：Vue Router
- **状态**：Pinia

### 部署
- **容器**：Docker
- **编排**：Docker Compose
- **反向代理**：Nginx
- **数据库**：SQLite / MySQL

## 安全特性

- ✅ SQL 注入防护（GORM 参数化查询）
- ✅ XSS 防护（输入清理 + CSP）
- ✅ JWT Token 认证
- ✅ HTTPS/TLS 加密（可选）


## FAQ

### 1. 部署后无法访问面板

按顺序排查：

1. **服务是否起来了**：`cd /opt/dashgo && docker compose ps`，确认 `dashgo` 和 `nginx` 都是 Up
2. **看日志**：`docker compose logs -f dashgo`，启动阶段的报错一般就在这里
3. **端口**：确认 80/443 已放行（见下方防火墙配置）
4. **架构不匹配**：SQLite **不支持 linux arm64**，arm64 机器请选 MySQL
5. **反向代理**：套了 Nginx/Caddy 的话，确认 `/api` 和 WebSocket 已正确转发

以上都试过还不行，带上下面的信息开个 Issue：部署方式（预编译/源码）、数据库类型、CPU 架构、`docker compose logs --tail=30 dashgo` 的输出。

### 2. 未登录用户访问根目录报错

已修复：未登录用户访问 `/` 会自动重定向到 `/login`

### 3. 防火墙/安全组配置

确保开放以下端口：
- **80**：HTTP 访问
- **443**：HTTPS 访问（如果启用）
- **6379**：Redis（仅内部，不对外开放）

```bash
# Ubuntu/Debian
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# CentOS/RHEL
sudo firewall-cmd --permanent --add-service=http
sudo firewall-cmd --permanent --add-service=https
sudo firewall-cmd --reload
```

## 相关项目

同一套自建生态里的其他东西，按需取用：

| 项目 | 说明 |
|---|---|
| [nyat-bot](https://github.com/ZYHUO/nyat-bot) | 🐱 Telegram 群聊 agent，不是指令机器人——会读气氛、只在有话说的时候开口 |
| [CLIProxy-Quota-Tray](https://github.com/ZYHUO/CLIProxy-Quota-Tray) | 🖥️ Windows / Linux 托盘，实时盯 CLIProxyAPI 各 OAuth 账号的配额窗口 |
| [worker-falling-grass-c407](https://github.com/ZYHUO/worker-falling-grass-c407) | ⚡ Cloudflare Worker 单文件部署：VLESS 订阅 → Clash 配置 |
| [tg-newsbot](https://github.com/ZYHUO/tg-newsbot) | 📰 Telegram 频道新闻推送：RSS 轮询 → 去重 → LLM 中文摘要 |
| [nyatdb](https://github.com/ZYHUO/nyatdb) | 🗄️ Rust 写的嵌入式页面引擎（nyat-bot 在用） |

---

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License

## 支持

- **问题反馈**：[GitHub Issues](https://github.com/ZYHUO/dashGO/issues)
---

**⭐ 如果这个项目对你有帮助，请给个 Star！**

**And give me a cup of coffee🎇** 0x728426bb2d4121da5316f795017cbf068e0db0d0 polygon
0x728426bb2d4121da5316f795017cbf068e0db0d0 xlayer
