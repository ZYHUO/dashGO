
# dashGO

这是一个用 Go 写的代理面板，反正够用就对了。

## 致谢

本项目的开发离不开以下开源项目和工具的支持：

- [Xboard](https://github.com/cedar2025/Xboard) - 感谢 cedar2025 提供的 Xboard 原版项目，本项目参考了其设计理念和数据库结构（注：本项目已重命名为 dashGO）
- [sing-box 脚本](https://github.com/fscarmen/sing-box) - 感谢 fscarmen 提供的 sing-box 一键安装脚本 参考了一下节点配置部分
- [AWS Kiro](https://kiro.dev) - 感谢 AWS Kiro 提供的 Claude AI 辅助开发

## 许可证

MIT License

## 已知问题

目前  **不支持多用户流控**  也就是说不统计流量
**无支付功能** 只提供余额和优惠券 后续也不可能写

---

## 有啥功能？

- 用户管理：注册、登录、改密码、看流量
- 套餐管理：周期、流量、速度都能限制
- 订单管理：下单、支付、取消
- 节点管理：支持 Shadowsocks、VMess、VLESS 等 (只测试了ss2022其他都没有 自己测测看)
- 订阅管理：Clash、sing-box、Base64 格式都支持
- 工单系统：用户提问题，管理员回复
- 邀请返利：邀请码、佣金统计
- 后台管理：该有的都有(应该够用)

---

## 怎么跑？

### 一键安装（推荐）

```bash
curl -sSL https://raw.githubusercontent.com/ZYHUO/dashGO/refs/heads/main/install.sh -o install.sh && bash install.sh
```

> 💡 脚本会自动下载预编译二进制文件，无需本地编译环境。

**功能菜单**：
- 1️⃣ 全新安装（本地开发）- SQLite/MySQL
- 2️⃣ 安装到现有 MySQL 数据库
- 3️⃣ 升级现有数据库（保留数据）
- 4️⃣ 修复迁移问题
- 5️⃣ 查看迁移状态
- 6️⃣ 生成配置文件

### 快速开始（需配置编译环境 提前git下）

```bash
bash setup.sh
# 选择 1 (全新安装)
# 选择 1 (SQLite - 推荐)
./dashgo-server
```

访问：http://localhost:8080

**默认管理员**：
- 邮箱：`admin@example.com`
- 密码：`admin123456`

> 💡 **新特性**：现在默认使用 SQLite 数据库，无需安装 MySQL 即可快速启动！适合开发、测试和小规模部署。

### 手动安装

如果不想用脚本，可以手动操作：

```bash
# 1. 配置文件（已有默认配置，可直接使用）
# configs/config.yaml 已配置为 SQLite

# 2. 编译
make build          # 编译 Server
make agent          # 编译 Agent
make frontend-build # 编译前端

# 3. 运行迁移
./migrate-linux-amd64 -config configs/config.yaml

# 4. 启动
./dashgo-server -config configs/config.yaml
```

### 数据库选择

**SQLite（默认）**：
- ✅ 零配置，开箱即用
- ✅ 适合 < 1000 用户
- ✅ 数据文件：`data/dashgo.db`

**MySQL**：
- 修改 `configs/config.yaml`：
```yaml
database:
  driver: "mysql"
  database: "dashgo"
  host: "127.0.0.1"
  port: 3306
  username: "root"
  password: "your_password"
```

### 详细文档

- 📖 [完整安装指南](README_SETUP.md)
- 🚀 [SQLite 快速开始](QUICK_START_SQLITE.md) ⭐ 推荐
- 📦 [预编译文件下载](docs/prebuilt-binaries.md)
- 🔧 [编译指南](BUILD.md)
- 📋 [更新日志](CHANGELOG.md)
- 🔄 [Agent 自动更新](docs/agent-auto-update.md)
- 📚 [更多文档](docs/)

---

## 编译

支持多平台编译：

```bash
# Linux/macOS
./build-all.sh all

# Windows
.\build-all.ps1 -Target all

# 或使用 Makefile
make release
```

详见 [编译指南](BUILD.md)

---

## 配置说明

主要配置项（`configs/config.yaml`）：

```yaml
app:
  listen: ":8080"

database:
  driver: "sqlite"              # sqlite 或 mysql
  database: "data/dashgo.db"    # SQLite 文件路径

redis:
  host: "127.0.0.1"
  port: 6379

jwt:
  secret: "your-random-secret"  # 改成随机字符串
  expire_hour: 24

node:
  token: "your-node-token"      # Agent 通信 Token
```

---

## 项目结构

```
dashGO/
├── cmd/
│   ├── server/          # Server 主程序
│   └── migrate/         # 数据库迁移工具
├── agent/               # Agent 程序
├── configs/             # 配置文件
├── internal/            # 后端核心
│   ├── handler/         # HTTP 处理器
│   ├── service/         # 业务逻辑
│   ├── model/           # 数据模型
│   ├── repository/      # 数据访问
│   └── protocol/        # 订阅协议
├── pkg/                 # 公共库
├── web/                 # Vue 前端
├── docs/                # 文档
└── migrations/          # 数据库迁移
```

---

## 常见问题

### 1. 如何切换数据库？

编辑 `configs/config.yaml`，修改 `database.driver` 为 `mysql` 或 `sqlite`。

### 2. 如何备份数据？

**SQLite**：
```bash
cp data/dashgo.db data/dashgo.db.backup
```

**MySQL**：
```bash
mysqldump -u root -p dashgo > backup.sql
```

### 3. 如何更新？

```bash
git pull
bash setup.sh  # 选择 3 (升级数据库)
```

### 4. Agent 如何配置？

参考 [Agent 自动更新文档](docs/agent-auto-update.md)

---

## API 文档

主要 API 端点：

**用户端**：
- `POST /api/v1/guest/register` - 注册
- `POST /api/v1/guest/login` - 登录
- `GET /api/v1/user/subscribe` - 获取订阅

**管理端**：
- `GET /api/v2/admin/stats/overview` - 数据概览
- `GET /api/v2/admin/users` - 用户管理
- `GET /api/v2/admin/servers` - 节点管理

完整 API 文档见 `docs/` 目录。
