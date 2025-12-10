# XBoard 快速安装

## 🚀 一键安装

### 方式 1: 使用安装脚本（推荐）

```bash
# 克隆项目
git clone https://github.com/ZYHUO/xboard-go.git
cd xboard-go

# 运行安装脚本
chmod +x local-install.sh
bash local-install.sh
```

### 方式 2: 使用 Makefile

```bash
# 开发环境
make install-dev

# 生产环境
make install-prod

# 查看所有命令
make help
```

## 📦 安装选项

### 开发环境（本地测试）

```bash
# 使用脚本
bash local-install.sh dev

# 或使用 Makefile
make install-dev

# 启动
make run
```

**特点：**
- ✅ SQLite 数据库（无需 MySQL）
- ✅ 快速启动
- ✅ 支持热重载
- ✅ Debug 模式

### 生产环境（Docker）

```bash
# 使用脚本
bash local-install.sh prod

# 或使用 Makefile
make install-prod

# 管理
docker compose ps      # 查看状态
docker compose logs -f # 查看日志
```

**特点：**
- ✅ MySQL + Redis
- ✅ Nginx 反向代理
- ✅ 容器化部署
- ✅ 自动生成密码

### 编译二进制

```bash
# 编译所有平台
make build-all

# 只编译当前平台
make build

# 编译 Agent
make agent-all
```

## 🔧 依赖要求

### 开发环境
- Go >= 1.21
- Node.js >= 16
- npm >= 8

### 生产环境
- Docker >= 20.10
- Docker Compose >= 2.0

## 📝 配置文件

配置文件位置：`configs/config.yaml`

```yaml
app:
  name: "XBoard"
  url: "http://localhost:8080"
  
database:
  type: "sqlite"  # 或 "mysql"
  database: "xboard.db"
  
# ... 更多配置
```

## 🎯 快速命令

```bash
# 开发
make dev              # 启动开发服务器
make dev-watch        # 启动并监听文件变化

# 构建
make build            # 构建服务器
make agent            # 构建 Agent
make release          # 构建所有平台

# 前端
make frontend-dev     # 启动前端开发服务器
make frontend-build   # 构建前端

# 数据库
make migrate          # 运行迁移

# Docker
make docker-run       # 启动容器
make docker-stop      # 停止容器

# 测试
make test             # 运行测试

# 帮助
make help             # 查看所有命令
```

## 🌐 访问地址

### 开发环境
- 后端: http://localhost:8080
- 前端: http://localhost:3000
- 后台: http://localhost:8080/admin

### 生产环境
- 面板: http://YOUR_IP:80
- 后台: http://YOUR_IP:80/admin

## 🔑 默认账户

```
邮箱: admin@xboard.local
密码: admin123
```

**⚠️ 请及时修改默认密码！**

## 📚 详细文档

- [完整安装指南](docs/local-installation.md)
- [安装指南](docs/install-guide.md)
- [API 文档](docs/README.md)
- [用户组设计](docs/user-group-design.md)

## ❓ 常见问题

### 端口被占用？

```bash
# 修改配置文件中的端口
vim configs/config.yaml

# 或停止占用端口的程序
lsof -i :8080
kill -9 <PID>
```

### 数据库连接失败？

检查 `configs/config.yaml` 中的数据库配置。

### Docker 启动失败？

```bash
# 检查 Docker 服务
sudo systemctl status docker

# 启动 Docker
sudo systemctl start docker
```

## 🆘 获取帮助

- Issues: https://github.com/ZYHUO/xboard-go/issues
- Discussions: https://github.com/ZYHUO/xboard-go/discussions

## 📄 许可证

MIT License


## 🔄 升级现有数据库

如果你已经有旧版本的数据库，想要在**不清除数据**的情况下升级：

```bash
# 使用升级脚本（推荐）
bash upgrade.sh

# 或手动升级
bash migrate.sh up
```

**升级脚本会：**
- ✅ 自动备份数据库
- ✅ 检查数据完整性
- ✅ 只执行新的迁移
- ✅ 保留所有现有数据
- ✅ 验证升级结果

**详细说明：** [docs/upgrade-guide.md](docs/upgrade-guide.md)
