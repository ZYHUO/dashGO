# XBoard 本地安装指南

本指南适用于从 Git 克隆后的本地安装。

## 快速开始

### 1. 克隆项目

```bash
git clone https://github.com/ZYHUO/xboard-go.git
cd xboard-go
```

### 2. 运行安装脚本

```bash
chmod +x local-install.sh
bash local-install.sh
```

## 安装选项

### 选项 1: 开发环境安装

适合本地开发和测试，使用 SQLite 数据库。

```bash
bash local-install.sh dev
```

**特点：**
- 使用 SQLite 数据库（无需安装 MySQL）
- 启用 Debug 模式
- 支持热重载
- 快速启动

**启动方式：**

```bash
# 启动后端
go run ./cmd/server -config configs/config.yaml

# 或使用 Makefile
make run

# 启动前端（另一个终端）
cd web && npm run dev
```

**访问地址：**
- 后端 API: http://localhost:8080
- 前端界面: http://localhost:3000

### 选项 2: 生产环境安装

使用 Docker Compose 部署完整的生产环境。

```bash
bash local-install.sh prod
```

**特点：**
- 使用 MySQL 数据库
- 使用 Redis 缓存
- 使用 Nginx 反向代理
- 自动生成安全密码
- 容器化部署

**管理命令：**

```bash
# 查看状态
docker compose ps

# 查看日志
docker compose logs -f

# 重启服务
docker compose restart

# 停止服务
docker compose down
```

**访问地址：**
- 面板地址: http://YOUR_IP:80
- 后台地址: http://YOUR_IP:80/admin

### 选项 3: 编译二进制文件

编译可执行文件，用于手动部署。

```bash
bash local-install.sh build
```

**输出文件：**
- `build/xboard` - 主程序
- `build/xboard-migrate` - 数据库迁移工具
- `build/xboard-agent` - 节点 Agent

**运行方式：**

```bash
./build/xboard -config build/configs/config.yaml
```

### 选项 4: 运行数据库迁移

手动运行数据库迁移脚本。

```bash
bash local-install.sh migrate
```

**说明：**
- 自动读取 `configs/config.yaml` 中的数据库配置
- 执行 `migrations/` 目录下的所有 SQL 文件
- 跳过 `*_rollback.sql` 文件

### 选项 5: 构建前端

单独构建前端资源。

```bash
bash local-install.sh frontend
```

**输出目录：** `web/dist`

## 依赖要求

### 开发环境

- **Go** >= 1.21
- **Node.js** >= 16
- **npm** >= 8

### 生产环境

- **Docker** >= 20.10
- **Docker Compose** >= 2.0

### 可选依赖

- **MySQL** >= 8.0（生产环境）
- **Redis** >= 6.0（生产环境）
- **Git**（克隆代码）

## 安装依赖

### Ubuntu/Debian

```bash
# 安装 Go
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# 安装 Node.js
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs

# 安装 Docker
curl -fsSL https://get.docker.com | sh
```

### CentOS/RHEL

```bash
# 安装 Go
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# 安装 Node.js
curl -fsSL https://rpm.nodesource.com/setup_18.x | sudo bash -
sudo yum install -y nodejs

# 安装 Docker
curl -fsSL https://get.docker.com | sh
```

### macOS

```bash
# 使用 Homebrew
brew install go node

# 安装 Docker Desktop
# 从 https://www.docker.com/products/docker-desktop 下载
```

### Windows

1. 安装 Go: https://go.dev/dl/
2. 安装 Node.js: https://nodejs.org/
3. 安装 Docker Desktop: https://www.docker.com/products/docker-desktop
4. 使用 Git Bash 或 WSL2 运行脚本

## 配置说明

### 配置文件位置

- 主配置: `configs/config.yaml`
- 环境变量: `.env`（生产环境）
- 密码文件: `.passwords`（生产环境）

### 修改配置

编辑 `configs/config.yaml`：

```yaml
app:
  name: "XBoard"
  url: "http://your-domain.com"  # 修改为你的域名
  debug: false
  jwt_secret: "your-secret-key"

database:
  type: "mysql"  # 或 "sqlite"
  host: "localhost"
  port: 3306
  username: "xboard"
  password: "your-password"
  database: "xboard"

mail:
  driver: "smtp"
  host: "smtp.gmail.com"
  port: 587
  username: "your-email@gmail.com"
  password: "your-app-password"
  from_address: "your-email@gmail.com"
  from_name: "XBoard"
```

## 常见问题

### Q1: 提示缺少 Go 或 Node.js？

**A:** 请先安装依赖，参考上面的"安装依赖"部分。

### Q2: 数据库连接失败？

**A:** 检查配置文件中的数据库信息是否正确：
- 主机地址
- 端口号
- 用户名和密码
- 数据库名称

### Q3: 端口被占用？

**A:** 修改配置文件中的端口号，或停止占用端口的程序。

```bash
# 查看端口占用
lsof -i :8080

# 停止占用端口的程序
kill -9 <PID>
```

### Q4: Docker 启动失败？

**A:** 检查 Docker 服务是否运行：

```bash
# 启动 Docker
sudo systemctl start docker

# 查看 Docker 状态
sudo systemctl status docker
```

### Q5: 前端构建失败？

**A:** 清理缓存后重试：

```bash
cd web
rm -rf node_modules package-lock.json
npm install
npm run build
```

### Q6: 如何重置数据库？

**A:** 

```bash
# SQLite
rm xboard.db

# MySQL
mysql -u root -p -e "DROP DATABASE xboard; CREATE DATABASE xboard;"

# 重新运行迁移
bash local-install.sh migrate
```

## 目录结构

```
xboard-go/
├── cmd/                    # 命令行工具
│   ├── server/            # 主程序
│   └── migrate/           # 迁移工具
├── configs/               # 配置文件
│   └── config.yaml
├── internal/              # 内部代码
│   ├── handler/          # HTTP 处理器
│   ├── service/          # 业务逻辑
│   ├── model/            # 数据模型
│   └── repository/       # 数据访问
├── migrations/            # 数据库迁移
├── web/                   # 前端代码
├── agent/                 # 节点 Agent
├── build/                 # 编译输出
├── docker-compose.yaml    # Docker 配置
├── Dockerfile            # Docker 镜像
├── Makefile              # Make 命令
├── local-install.sh      # 本地安装脚本
└── README.md             # 项目说明
```

## 下一步

安装完成后，你可以：

1. **登录后台**
   - 访问 `/admin` 路径
   - 使用默认账户登录
   - 修改默认密码

2. **配置节点**
   - 在后台添加节点
   - 获取节点 Token
   - 部署 Agent

3. **创建套餐**
   - 设置流量和价格
   - 配置用户组
   - 关联节点

4. **测试功能**
   - 注册测试用户
   - 购买套餐
   - 获取订阅链接

## 获取帮助

- 文档: [docs/](../docs/)
- Issues: https://github.com/ZYHUO/xboard-go/issues
- 讨论: https://github.com/ZYHUO/xboard-go/discussions

## 许可证

MIT License
