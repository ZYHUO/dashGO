# dashGO

一个现代化的代理面板管理系统，基于 Go + Vue 3 构建。

[測試站点](https://misaka.cfd/) 
帐密
admin@example.com
admin123456
## 特性

- 🚀 **高性能**：Go 后端，Vue 3 前端
- 💾 **灵活数据库**：支持 SQLite（默认）和 MySQL
- 🔒 **安全可靠**：JWT 认证、权限控制、SQL 注入防护
- 🐳 **容器化部署**：Docker Compose 一键部署
- 📊 **流量统计**：实时流量监控和统计
- 🎨 **现代 UI**：响应式设计，支持移动端

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

### 默认账号

```
邮箱：admin@example.com
密码：admin123456
```

**⚠️ 首次登录后请立即修改密码！**

## 系统要求

- **操作系统**：Linux（Ubuntu/Debian/CentOS）
- **内存**：最低 512MB，推荐 1GB+
- **磁盘**：最低 1GB 可用空间
- **软件**：Docker 和 Docker Compose

## 文档

- [构建指南](BUILD.md) - 如何从源码构建
- [故障排除](TROUBLESHOOTING.md) - 常见问题解决
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
- ✅ 密码 bcrypt 加密
- ✅ 管理员权限控制
- ✅ 安全响应头
- ✅ CORS 配置
- ✅ IP 白名单（可选）
- ✅ HTTPS/TLS 加密（可选）
- ✅ CDN 支持（Cloudflare 等）

详见 [SECURITY.md](SECURITY.md)

## 常见问题

### 1. 使用域名访问超时（ERR_TIMED_OUT）

**原因**：Cloudflare CDN 使用 HTTPS，但后端只有 HTTP

**解决方案**：
- **方案 A**：启用 HTTPS（推荐）
  ```bash
  bash install.sh panel
  # 选择启用 HTTPS，使用 Cloudflare Origin Certificate
  ```
- **方案 B**：Cloudflare 设置 SSL/TLS 为 **Flexible** 模式

### 2. 未登录用户访问根目录报错

已修复：未登录用户访问 `/` 会自动重定向到 `/login`

### 3. 日志中出现 "record not found"

已优化：GORM 日志级别调整为 Error，正常的查询不存在记录不会再记录日志

### 4. 防火墙/安全组配置

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

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License

## 支持

- **问题反馈**：[GitHub Issues](https://github.com/ZYHUO/dashGO/issues)
---

**⭐ 如果这个项目对你有帮助，请给个 Star！**
**Give me a cup of coffee** 0x728426bb2d4121da5316f795017cbf068e0db0d0 polygon
0x728426bb2d4121da5316f795017cbf068e0db0d0 xlayer
