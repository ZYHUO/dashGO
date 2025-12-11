# XBoard 快速开始 (SQLite)

XBoard 现在默认使用 SQLite 数据库，无需安装 MySQL 即可快速启动！

## 一键启动（推荐）

```bash
# 1. 使用安装脚本
bash setup.sh

# 选择 "1) 全新安装 (本地开发)"
# 选择 "1) SQLite (推荐用于开发和小规模部署)"
```

## 手动启动

### 1. 准备配置文件

```bash
# 配置文件已经默认使用 SQLite
cp configs/config.example.yaml configs/config.yaml

# 或者直接使用已有的 config.yaml（已配置为 SQLite）
```

### 2. 编译项目

```bash
# 编译 Server
make build

# 编译 Agent
make agent
```

### 3. 运行数据库迁移

```bash
# 使用 migrate 工具
./migrate-linux-amd64 -config configs/config.yaml

# 或使用 make
make migrate
```

### 4. 启动服务

```bash
# 启动 Server
./xboard -config configs/config.yaml

# 或使用 make
make run
```

### 5. 访问面板

打开浏览器访问：`http://localhost:8080`

默认管理员账号：
- 邮箱：`admin@example.com`
- 密码：`admin123456`

**⚠️ 首次登录后请立即修改密码！**

## SQLite vs MySQL

### SQLite（默认）

**优点**：
- ✅ 无需安装数据库服务器
- ✅ 零配置，开箱即用
- ✅ 适合开发、测试和小规模部署
- ✅ 数据库文件便于备份和迁移
- ✅ 性能足够应对中小规模使用

**适用场景**：
- 本地开发和测试
- 个人使用或小团队
- 用户数 < 1000
- 节点数 < 50

### MySQL

**优点**：
- ✅ 更好的并发性能
- ✅ 适合大规模生产环境
- ✅ 支持主从复制和集群
- ✅ 更丰富的管理工具

**适用场景**：
- 生产环境
- 大规模部署
- 用户数 > 1000
- 需要高可用性

## 切换到 MySQL

如果需要切换到 MySQL，编辑 `configs/config.yaml`：

```yaml
database:
  driver: "mysql"  # 改为 mysql
  database: "xboard"  # 数据库名
  host: "127.0.0.1"
  port: 3306
  username: "root"
  password: "your_password"
```

然后重新运行迁移：

```bash
./migrate-linux-amd64 -config configs/config.yaml
```

## 数据库文件位置

SQLite 数据库文件默认位置：`data/xboard.db`

### 备份数据库

```bash
# 简单备份
cp data/xboard.db data/xboard.db.backup

# 带时间戳的备份
cp data/xboard.db "data/xboard.db.$(date +%Y%m%d_%H%M%S).backup"
```

### 恢复数据库

```bash
# 停止服务
pkill xboard

# 恢复备份
cp data/xboard.db.backup data/xboard.db

# 重启服务
./xboard -config configs/config.yaml
```

## 常见问题

### 1. 数据库文件权限错误

```bash
# 确保 data 目录存在且有写权限
mkdir -p data
chmod 755 data
```

### 2. 数据库被锁定

SQLite 不支持高并发写入。如果遇到 "database is locked" 错误：

- 检查是否有多个进程在访问数据库
- 考虑切换到 MySQL

### 3. 性能优化

SQLite 性能优化建议：

```yaml
# 在 config.yaml 中可以添加 SQLite 特定配置
database:
  driver: "sqlite"
  database: "data/xboard.db?cache=shared&mode=rwc&_journal_mode=WAL"
```

WAL 模式可以提高并发性能。

### 4. 数据迁移

从 SQLite 迁移到 MySQL：

```bash
# 1. 导出 SQLite 数据
sqlite3 data/xboard.db .dump > xboard_dump.sql

# 2. 转换 SQL 语法（SQLite -> MySQL）
# 需要手动调整一些语法差异

# 3. 导入到 MySQL
mysql -u root -p xboard < xboard_converted.sql

# 4. 修改 config.yaml 为 MySQL
# 5. 重启服务
```

## 开发建议

### 开发环境

使用 SQLite，快速启动：

```bash
# 开发模式（带热重载）
make dev-watch
```

### 生产环境

使用 MySQL，更好的性能和可靠性：

```bash
# 生产模式
./xboard -config configs/config.yaml
```

## 相关文档

- [完整安装指南](README_SETUP.md)
- [数据库迁移](docs/database-migration.md)
- [编译指南](BUILD.md)
- [Agent 自动更新](docs/agent-auto-update.md)

## 技术支持

如有问题，请查看：
- [GitHub Issues](https://github.com/your-org/xboard-go/issues)
- [文档目录](docs/)
