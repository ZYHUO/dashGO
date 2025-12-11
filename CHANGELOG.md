# XBoard-Go v1.0.0 更新日志

## 发布日期：2024-12-11

---

## 🎉 重要变更

### 默认使用 SQLite 数据库

**功能描述**：现在默认使用 SQLite 数据库，无需安装 MySQL 即可快速启动

**优势**：
- ✅ 零配置，开箱即用
- ✅ 无需安装数据库服务器
- ✅ 适合开发、测试和小规模部署
- ✅ 数据库文件便于备份和迁移
- ✅ 自动创建 data 目录

**配置示例**：
```yaml
database:
  driver: "sqlite"
  database: "data/xboard.db"
```

**切换到 MySQL**：
```yaml
database:
  driver: "mysql"
  database: "xboard"
  host: "127.0.0.1"
  port: 3306
  username: "root"
  password: "your_password"
```

**相关文档**：
- [SQLite 快速开始](QUICK_START_SQLITE.md)
- [完整安装指南](README_SETUP.md)

---

## 📦 部署优化

### 预编译二进制文件

**功能描述**：提供预编译二进制文件，无需本地编译

**实现**：
- ✅ 编译 `xboard-server-linux-amd64` - 主服务器程序
- ✅ 编译 `migrate-linux-amd64` - 数据库迁移工具
- ✅ 编译 `xboard-agent-linux-amd64` - 节点代理程序
- ✅ 更新 `setup.sh` 支持自动下载二进制
- ✅ 更新 `agent/install.sh` 支持自动下载二进制
- ✅ 支持自动检测系统架构 (amd64/arm64)

**下载地址**：
```
https://download.sharon.wiki/xboard-server-linux-amd64
https://download.sharon.wiki/migrate-linux-amd64
https://download.sharon.wiki/xboard-agent-linux-amd64
```

**使用方式**：
```bash
# 安装 Dashboard（自动下载二进制）
bash setup.sh

# 安装 Agent（自动下载二进制）
curl -sL https://raw.githubusercontent.com/ZYHUO/xboard-go/main/agent/install.sh | bash -s -- <面板地址> <Token>
```

**优势**：
- 🚀 安装速度更快（无需编译）
- 💾 节省磁盘空间（无需 Go 环境）
- 🔧 降低安装门槛（无需配置编译环境）
- ✨ 支持多架构（amd64/arm64）

**相关文件**：
- `setup.sh` - 添加 `download_binaries()` 函数
- `agent/install.sh` - 更新下载 URL
- `BINARIES_UPLOAD.md` - 上传说明文档

---

## 🎉 新功能

### 1. Agent 自动更新功能

**功能描述**：Agent 支持自动检测新版本并自动更新，无需手动干预

**实现**：
- ✅ 版本管理和语义化版本支持（SemVer）
- ✅ 定期检查更新（可配置间隔）
- ✅ 安全下载（HTTPS only，SHA256 验证）
- ✅ 原子更新和自动回滚
- ✅ 更新策略控制（auto/manual）
- ✅ 更新历史记录
- ✅ 零停机更新（sing-box 继续运行）

**使用方式**：
```bash
# 启用自动更新（默认）
xboard-agent -panel https://panel.example.com -token abc123

# 自定义检查间隔（每30分钟）
xboard-agent -panel https://panel.example.com -token abc123 -update-check-interval=1800

# 禁用自动更新
xboard-agent -panel https://panel.example.com -token abc123 -auto-update=false

# 手动触发更新
xboard-agent -panel https://panel.example.com -token abc123 -update
```

**安全机制**：
- 🔒 HTTPS 强制验证
- 🔒 SHA256 文件完整性验证
- 🔒 Token 认证（最小16字符）
- 🔒 路径遍历防护
- 🔒 文件权限验证
- 🔒 原子操作替换
- 🔒 自动回滚机制

**相关文件**：
- `agent/version.go` - 版本管理
- `agent/update_checker.go` - 更新检查
- `agent/downloader.go` - 文件下载
- `agent/verifier.go` - 文件验证
- `agent/updater.go` - 更新执行
- `agent/security.go` - 安全验证
- `agent/update_history.go` - 历史记录
- `agent/update_error.go` - 错误处理
- `agent/update_notifier.go` - 通知服务
- `docs/agent-auto-update.md` - 完整文档

---

### 2. 节点-主机绑定功能

**功能描述**：节点可以选择绑定到主机，实现自动部署

**实现**：
- ✅ Server 模型添加 `host_id` 字段
- ✅ AdminCreateServer 支持 `host_id` 参数
- ✅ AdminUpdateServer 支持修改绑定
- ✅ AdminListServers 返回主机名称
- ✅ 前端界面支持选择绑定主机

**使用场景**：
```
1. 创建主机（Host）
2. 创建节点（Server），选择绑定到主机
3. Agent 自动在主机上部署节点配置
4. 用户通过订阅获取节点信息
```

**API 示例**：
```json
POST /api/v2/admin/server
{
  "name": "香港节点1",
  "type": "shadowsocks",
  "host": "hk1.example.com",
  "port": "443",
  "host_id": 1,  // 绑定到主机ID=1
  "rate": 1.0,
  "show": true
}
```

**相关文件**：
- `internal/model/server.go`
- `internal/handler/admin.go`
- `internal/service/host.go`
- `web/src/views/admin/Servers.vue`
- `docs/server-host-binding.md`

---

### 2. 套餐购买数量限制

**功能描述**：套餐支持设置最大可售数量，实现库存管理

**实现**：
- ✅ Plan 模型添加 `sold_count` 字段
- ✅ `CanPurchase()` 方法检查是否可购买
- ✅ `GetRemainingCount()` 方法获取剩余数量
- ✅ PlanService 提供计数管理方法
- ✅ Repository 实现原子操作

**使用场景**：
```
1. 创建限量套餐（capacity_limit = 100）
2. 用户购买时自动增加 sold_count
3. 达到限制后自动停止销售
4. 用户退订时减少 sold_count
```

**API 响应**：
```json
{
  "id": 1,
  "name": "限量套餐",
  "capacity_limit": 100,
  "sold_count": 85,
  "remaining_count": 15,
  "can_purchase": true
}
```

**相关文件**：
- `internal/model/plan.go`
- `internal/service/plan.go`
- `internal/repository/plan.go`
- `migrations/005_add_plan_sold_count.sql`
- `docs/plan-purchase-limit.md`

---

### 3. 一键安装/升级/修复脚本

**功能描述**：统一的脚本管理所有安装、升级、修复操作

**功能**：
- ✅ 全新安装（本地开发）
- ✅ 安装到现有 MySQL 数据库
- ✅ 升级现有数据库（保留数据）
- ✅ 修复迁移问题
- ✅ 查看迁移状态
- ✅ 生成配置文件

**使用**：
```bash
bash setup.sh
```

**相关文件**：
- `setup.sh` - 一键脚本
- `README_SETUP.md` - 完整文档
- `QUICK_START.md` - 快速开始

---

## 🐛 Bug 修复

### 1. 配置文件字段名称错误

**问题**：安装脚本生成的配置使用 `type: "mysql"`，但代码期望 `driver: "mysql"`

**影响**：导致数据库连接失败，报错 "unsupported database driver"

**修复**：
- ✅ 所有脚本改为生成 `driver` 字段
- ✅ 更新所有文档示例
- ✅ 创建修复指南

**相关文件**：
- `install-existing-db.sh`
- `local-install.sh`
- `install.sh`
- `upgrade.sh`
- `docs/local-installation.md`
- `QUICK_INSTALL.md`
- `UPGRADE_MYSQL.md`
- `FIX_CONFIG.md`

---

### 2. 迁移系统执行 rollback 文件

**问题**：迁移系统会执行所有 `.sql` 文件，包括 `_rollback.sql`，导致字段被删除

**影响**：
- `host_id` 字段被删除，保存节点失败
- `sold_count` 字段被删除，套餐库存功能失效

**修复**：
- ✅ 修改迁移系统，跳过 `_rollback.sql` 文件
- ✅ 创建修复脚本自动修复数据库
- ✅ 添加字段检查和自动添加逻辑

**相关文件**：
- `cmd/migrate/main.go`
- `setup.sh` (选项 4)
- `MIGRATION_FIX.md`

---

## 📚 文档更新

### 新增文档

1. **README_SETUP.md** - 完整的安装和升级指南
   - 详细的功能说明
   - 常见问题解答
   - 手动操作指南
   - 架构说明

2. **QUICK_START.md** - 快速开始指南
   - 一键安装命令
   - 常用场景
   - 快速排查

3. **docs/server-host-binding.md** - 节点-主机绑定设计文档
   - 设计目标
   - 实现方案
   - API 变更
   - 使用场景

4. **docs/plan-purchase-limit.md** - 套餐购买数量限制设计文档
   - 需求说明
   - 设计方案
   - 购买逻辑
   - 使用场景

5. **MIGRATION_FIX.md** - 迁移问题修复指南
   - 问题分析
   - 修复方案
   - 验证步骤

6. **ARCHITECTURE_CLARIFICATION.md** - 架构说明文档
   - 数据模型
   - 绑定关系
   - 核心方法
   - 工作流程

### 更新文档

1. **docs/local-installation.md** - 更新配置示例
2. **docs/database-migration.md** - 更新迁移说明
3. **QUICK_INSTALL.md** - 更新配置字段
4. **UPGRADE_MYSQL.md** - 更新升级步骤

---

## 🔧 技术改进

### 1. 迁移系统优化

**改进**：
- 跳过 `_rollback.sql` 文件
- 更好的错误处理
- 支持字段已存在的情况

**代码**：
```go
// 过滤并排序 SQL 文件（跳过 rollback 文件）
var sqlFiles []string
for _, f := range files {
    if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") && 
       !strings.Contains(f.Name(), "_rollback") {
        sqlFiles = append(sqlFiles, f.Name())
    }
}
```

### 2. 原子操作

**改进**：套餐购买数量使用原子操作，防止并发超卖

**代码**：
```go
// IncrementSoldCount 增加已售数量（原子操作）
func (r *PlanRepository) IncrementSoldCount(planID int64) error {
    return r.db.Model(&model.Plan{}).Where("id = ?", planID).
        UpdateColumn("sold_count", gorm.Expr("sold_count + ?", 1)).Error
}
```

### 3. 配置验证

**改进**：创建/更新节点时验证主机是否存在

**代码**：
```go
// 如果设置了 host_id，验证主机是否存在
if req.HostID != nil {
    if _, err := services.Host.GetByID(*req.HostID); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "host not found"})
        return
    }
}
```

---

## 📦 数据库变更

### 新增字段

#### v2_server.host_id
```sql
ALTER TABLE `v2_server` 
ADD COLUMN `host_id` BIGINT NULL DEFAULT NULL COMMENT '绑定的主机ID' AFTER `parent_id`;

ALTER TABLE `v2_server` 
ADD INDEX `idx_server_host_id` (`host_id`);
```

#### v2_plan.sold_count
```sql
ALTER TABLE `v2_plan` 
ADD COLUMN `sold_count` INT NOT NULL DEFAULT 0 COMMENT '已售出数量';

-- 初始化数据
UPDATE `v2_plan` p 
SET `sold_count` = (
    SELECT COUNT(*) 
    FROM `v2_user` u 
    WHERE u.`plan_id` = p.`id`
);

CREATE INDEX `idx_plan_capacity` ON `v2_plan`(`capacity_limit`, `sold_count`);
```

### 迁移文件

- `migrations/001_add_host_id_to_servers.sql`
- `migrations/005_add_plan_sold_count.sql`

---

## 🎯 API 变更

### 新增/修改的 API

#### 1. POST /api/v2/admin/server
**新增参数**：`host_id`

**请求示例**：
```json
{
  "name": "香港节点1",
  "type": "shadowsocks",
  "host": "hk1.example.com",
  "port": "443",
  "host_id": 1,  // 新增
  "rate": 1.0,
  "show": true
}
```

#### 2. PUT /api/v2/admin/server/:id
**新增参数**：`host_id`

**请求示例**：
```json
{
  "host_id": 2  // 修改绑定
}
```

#### 3. GET /api/v2/admin/servers
**新增响应字段**：`host_name`

**响应示例**：
```json
{
  "data": [
    {
      "id": 1,
      "name": "香港节点1",
      "host_id": 1,
      "host_name": "香港主机1"  // 新增
    }
  ]
}
```

#### 4. GET /api/v2/admin/plans
**新增响应字段**：`sold_count`, `remaining_count`, `can_purchase`

**响应示例**：
```json
{
  "data": [
    {
      "id": 1,
      "name": "基础套餐",
      "capacity_limit": 100,
      "sold_count": 85,           // 新增
      "remaining_count": 15,      // 新增
      "can_purchase": true        // 新增
    }
  ]
}
```

---

## ⚠️ 破坏性变更

### 配置文件字段名称

**旧版本**：
```yaml
database:
  type: "mysql"  # ❌ 不再支持
```

**新版本**：
```yaml
database:
  driver: "mysql"  # ✅ 必须使用
```

**迁移方法**：
```bash
# 方法1：使用脚本修复
bash setup.sh
# 选择 6 (重新生成配置)

# 方法2：手动修改
sed -i 's/type: "mysql"/driver: "mysql"/g' configs/config.yaml
```

---

## 📋 升级指南

### 从旧版本升级

#### 步骤1：备份数据

```bash
# MySQL
mysqldump -u root -p xboard > backup.sql

# SQLite
cp xboard.db xboard.db.backup
```

#### 步骤2：更新代码

```bash
git pull origin main
```

#### 步骤3：修改配置文件

```bash
# 将 type 改为 driver
sed -i 's/type: "mysql"/driver: "mysql"/g' configs/config.yaml
```

#### 步骤4：运行升级脚本

```bash
bash setup.sh
# 选择 3 (升级现有数据库)
```

#### 步骤5：重启服务

```bash
# Docker
docker compose restart

# Systemd
systemctl restart xboard

# 手动
./xboard-server
```

---

## 🧪 测试建议

### 1. 节点绑定功能

```bash
# 1. 创建主机
# 2. 创建节点，选择绑定主机
# 3. 编辑节点，修改绑定
# 4. 验证主机配置生成正确
```

### 2. 套餐库存功能

```bash
# 1. 创建限量套餐（capacity_limit = 10）
# 2. 购买套餐，验证 sold_count 增加
# 3. 购买到限制，验证 can_purchase = false
# 4. 用户退订，验证 sold_count 减少
```

### 3. 迁移修复

```bash
# 1. 运行修复脚本
bash setup.sh
# 选择 4

# 2. 验证字段存在
mysql -u root -p xboard -e "DESCRIBE v2_server;" | grep host_id
mysql -u root -p xboard -e "DESCRIBE v2_plan;" | grep sold_count

# 3. 验证迁移记录
mysql -u root -p xboard -e "SELECT * FROM migrations;"
```

---

## 🔮 未来计划

### v1.1.0

- [ ] 订单服务集成购买数量管理
- [ ] 前端套餐库存显示
- [ ] 管理后台库存预警
- [ ] 定时任务校验计数准确性

### v1.2.0

- [ ] 节点自动部署优化
- [ ] 主机监控和告警
- [ ] 批量操作支持
- [ ] API 文档完善

---

## 🙏 致谢

感谢所有贡献者和用户的支持！

---

## 📞 支持

- GitHub: https://github.com/ZYHUO/xboard-go
- Issues: https://github.com/ZYHUO/xboard-go/issues
- Discussions: https://github.com/ZYHUO/xboard-go/discussions

---

## 📄 许可证

MIT License
