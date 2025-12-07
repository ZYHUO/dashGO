
# XBoard Go

这是一个用 Go 写的代理面板，反正够用就对了。

## 致谢

本项目的开发离不开以下开源项目和工具的支持：

- [Xboard](https://github.com/cedar2025/Xboard) - 感谢 cedar2025 提供的 Xboard 原版项目，本项目参考了其设计理念和数据库结构
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

### 1. 配置文件

先复制一份配置文件：

```bash
cp configs/config.example.yaml configs/config.yaml
```

然后改 `configs/config.yaml`，填数据库、Redis、JWT 这些。

### 2. 编译运行

```bash
# 后端
go build -o xboard ./cmd/server

# 前端
cd web
npm install
npm run build
cd ..

# 启动
./xboard -config configs/config.yaml
```

### 3. 管理员账号

在 `configs/config.yaml` 里写：

```yaml
admin:
  email: "admin@example.com"
  password: "your_password"
```

启动后自动创建管理员。

---

## 配置说明（简单说）

- 数据库：MySQL 或 SQLite
- Redis：填地址和密码
- JWT：随便搞个随机字符串当 secret

---

## API（给开发用的）

用户端：
- `POST /api/v1/guest/register` 注册
- `POST /api/v1/guest/login` 登录
- `GET /api/v1/guest/plans` 套餐列表
- `GET /api/v1/user/info` 用户信息
- `GET /api/v1/user/subscribe` 订阅信息

管理员端：
- `GET /api/v2/admin/stats/overview` 概览
- `GET /api/v2/admin/users` 用户列表
- `GET /api/v2/admin/servers` 节点列表

---

## 项目结构（大概长这样）

```
xboard-go/
├── cmd/server/      # 主程序入口
├── configs/         # 配置文件
├── internal/        # 后端逻辑
│   ├── handler/     # 接口
│   ├── service/     # 业务逻辑
│   ├── model/       # 数据模型
│   └── protocol/    # 订阅生成
├── pkg/             # 工具类
└── web/             # 前端
```
