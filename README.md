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
   - SQLite（推荐,轻量级,linux arm64不支持,请不要选择此项
   - MySQL（外部数据库）

2. **安装方式**
   - 预编译版本（推荐，快速）
   - 源码构建（支持自定义）

3. **自定义配置**
   - Web 访问端口（默认 80）
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

## 管理命令

```bash
# 查看服务状态
cd /opt/dashgo && docker compose ps

# 查看日志
docker compose logs -f

# 重启服务
docker compose restart

# 停止服务
docker compose down

# 更新服务
docker compose pull && docker compose up -d
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

详见 [SECURITY.md](SECURITY.md)

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
