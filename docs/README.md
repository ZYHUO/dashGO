# XBoard Go

XBoard 的 Go 语言重写版本，后端节点使用 sing-box server。

## ✨ 特性

- 🚀 Go 语言重写，高性能低资源占用
- 🎨 Vue3 + TypeScript 现代化前端，马卡龙配色主题
- 📦 支持 sing-box、Clash、Surge、Quantumult X 等多种订阅格式
- 🔐 支持 AnyTLS、SS2022、VMess、VLESS、Trojan 等协议
- 📧 邮件通知系统
- 🤖 Telegram Bot 集成
- 🎫 完整工单系统
- 💾 兼容原版 XBoard 数据库，可平滑迁移

## 项目结构

```
dashGO/
├── cmd/server/main.go          # 入口
├── internal/
│   ├── config/                 # 配置管理
│   ├── model/                  # 数据模型
│   ├── repository/             # 数据访问层
│   ├── service/                # 业务逻辑
│   ├── handler/                # HTTP 处理器
│   ├── middleware/             # 中间件
│   └── protocol/               # 订阅协议生成
├── pkg/
│   ├── cache/                  # Redis 缓存
│   ├── database/               # 数据库连接
│   └── utils/                  # 工具函数
├── web/                        # Vue3 前端
│   ├── src/
│   │   ├── views/              # 页面组件
│   │   ├── layouts/            # 布局组件
│   │   ├── stores/             # Pinia 状态管理
│   │   ├── router/             # 路由配置
│   │   └── api/                # API 封装
│   └── ...
├── configs/                    # 配置文件
└── docs/                       # 文档
```

## 快速开始

### 使用 Docker Compose（推荐）

```bash
cd dashGO
cp configs/config.example.yaml configs/config.yaml
# 编辑 config.yaml 配置数据库等信息
docker-compose up -d
```

### 手动部署

#### 1. 配置数据库

编辑 `configs/config.yaml`:

```yaml
database:
  driver: "mysql"
  host: "127.0.0.1"
  port: 3306
  database: "xboard"
  username: "root"
  password: "your_password"

redis:
  host: "127.0.0.1"
  port: 6379
  password: ""
  db: 0
```

#### 2. 配置邮件（可选）

```yaml
mail:
  host: "smtp.example.com"
  port: 587
  username: "your_email@example.com"
  password: "your_password"
  from_name: "XBoard"
  from_addr: "noreply@example.com"
  encryption: "tls"
```

#### 3. 配置 Telegram Bot（可选）

```yaml
telegram:
  bot_token: "your_bot_token"
  chat_id: "admin_chat_id"
```

#### 4. 启动后端

```bash
cd dashGO
go mod tidy
go build -o dashgo cmd/server/main.go
./xboard -config configs/config.yaml
```

#### 5. 构建前端

```bash
cd web
npm install
npm run build
```

## 节点部署 (sing-box server)

详细文档请参考 [sing-box 集成指南](singbox-integration.md)

### 快速配置

1. 安装 sing-box（需要 ssmapi 支持）
2. 使用 `configs/singbox-server.json` 作为配置模板
3. 启动节点：`./sing-box run -c config.json`

## API 接口

### 用户端 API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/v1/guest/register | 用户注册 |
| POST | /api/v1/guest/login | 用户登录 |
| GET | /api/v1/guest/plans | 获取套餐列表 |
| GET | /api/v1/user/info | 获取用户信息 |
| GET | /api/v1/user/subscribe | 获取订阅信息 |
| GET | /api/v1/user/tickets | 获取工单列表 |
| POST | /api/v1/user/ticket/create | 创建工单 |
| GET | /api/v1/client/subscribe | 获取订阅配置 |

### 节点通信 API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/server/UniProxy/config | 获取节点配置 |
| GET | /api/v1/server/UniProxy/user | 获取用户列表 |
| POST | /api/v1/server/UniProxy/push | 流量上报 |
| POST | /api/v1/server/UniProxy/alive | 在线状态上报 |

### 管理端 API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v2/admin/servers | 获取服务器列表 |
| POST | /api/v2/admin/server | 创建服务器 |
| GET | /api/v2/admin/users | 获取用户列表 |
| GET | /api/v2/admin/plans | 获取套餐列表 |
| GET | /api/v2/admin/orders | 获取订单列表 |
| GET | /api/v2/admin/tickets | 获取工单列表 |
| GET | /api/v2/admin/settings | 获取系统设置 |

## 订阅格式支持

| 客户端 | 格式 | 参数 |
|--------|------|------|
| sing-box / Hiddify | JSON | `?format=singbox` |
| Clash / Clash Meta (mihomo) | YAML | `?format=clash` |
| Surge | 配置文件 | `?format=surge` |
| Quantumult X | 配置片段 | `?format=quantumultx` |
| 通用 | Base64 | 默认 |

## 支持的协议

- Shadowsocks 2022 (ss2022)
- VMess
- VLESS (含 Reality)
- Trojan
- Hysteria2
- AnyTLS

## 与原版 XBoard 的兼容性

- ✅ 数据库表结构完全兼容
- ✅ API 接口兼容（可平滑迁移）
- ✅ 节点通信协议兼容

## 功能完成度

- [x] 核心数据模型
- [x] 用户认证 (JWT)
- [x] 节点管理
- [x] 订阅生成（多格式）
- [x] 节点通信 API
- [x] 工单系统
- [x] 邮件通知
- [x] Telegram Bot
- [x] 前端界面（用户端 + 管理端）
- [x] 支付集成（易支付）
- [x] 优惠券系统
- [x] 邀请返利系统
- [x] 公告系统
- [x] 知识库系统
- [x] 定时任务（流量重置、到期提醒）
- [x] 统计报表

## 技术栈

### 后端
- Go 1.22+
- Gin (Web 框架)
- GORM (ORM)
- Redis (缓存)
- JWT (认证)

### 前端
- Vue 3
- TypeScript
- TailwindCSS
- Pinia
- Vue Router

## License

MIT
