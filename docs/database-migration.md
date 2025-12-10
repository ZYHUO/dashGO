# 数据库迁移指南

## 概述

XBoard 提供了两种数据库迁移方式：

1. **SQL 迁移** - 使用 SQL 文件进行精确控制（推荐生产环境）
2. **自动迁移** - 根据 Go 模型自动生成表结构（推荐开发环境）

## 快速开始

### 方式 1: 使用迁移脚本（推荐）

```bash
# 执行所有待执行的迁移
bash migrate.sh up

# 查看迁移状态
bash migrate.sh status

# 自动迁移（开发环境）
bash migrate.sh auto
```

### 方式 2: 使用 Go 工具

```bash
# 执行迁移
go run ./cmd/migrate -config configs/config.yaml -action up

# 查看状态
go run ./cmd/migrate -config configs/config.yaml -action status

# 自动迁移
go run ./cmd/migrate -config configs/config.yaml -action auto
```

### 方式 3: 使用 Makefile

```bash
# 执行迁移
make migrate

# 查看状态
make migrate-status

# 自动迁移
make migrate-auto
```

## SQL 迁移

### 迁移文件结构

```
migrations/
├── 001_add_host_id_to_servers.sql
├── 001_add_host_id_to_servers_rollback.sql
├── 002_add_user_fields.sql
├── 003_create_user_group.sql
├── 003_create_user_group_rollback.sql
└── 004_simplify_user_group.sql
```

### 命名规范

- 格式：`{序号}_{描述}.sql`
- 回滚文件：`{序号}_{描述}_rollback.sql`
- 示例：
  - `001_create_users_table.sql`
  - `001_create_users_table_rollback.sql`

### 创建新迁移

#### 使用脚本创建

```bash
bash migrate.sh create add_new_field
```

这会创建两个文件：
- `{timestamp}_add_new_field.sql`
- `{timestamp}_add_new_field_rollback.sql`

#### 手动创建

```sql
-- migrations/005_add_email_verified.sql

-- 添加邮箱验证字段
ALTER TABLE v2_user ADD COLUMN email_verified BOOLEAN DEFAULT FALSE COMMENT '邮箱是否已验证';

-- 添加验证码字段
ALTER TABLE v2_user ADD COLUMN verification_code VARCHAR(64) COMMENT '验证码';
```

回滚文件：

```sql
-- migrations/005_add_email_verified_rollback.sql

-- 删除邮箱验证字段
ALTER TABLE v2_user DROP COLUMN IF EXISTS email_verified;
ALTER TABLE v2_user DROP COLUMN IF EXISTS verification_code;
```

### 执行迁移

```bash
# 执行所有待执行的迁移
bash migrate.sh up

# 或使用 Go 工具
go run ./cmd/migrate -action up
```

### 回滚迁移

```bash
# 交互式选择回滚文件
bash migrate.sh down

# 或手动执行
mysql -u root -p xboard < migrations/005_add_email_verified_rollback.sql
```

## 自动迁移

### 适用场景

- ✅ 开发环境快速迭代
- ✅ 本地测试
- ✅ 原型开发
- ❌ 生产环境（不推荐）

### 使用方法

```bash
# 使用脚本
bash migrate.sh auto

# 或使用 Go 工具
go run ./cmd/migrate -action auto

# 或使用 Makefile
make migrate-auto
```

### 工作原理

自动迁移会根据 Go 模型定义自动创建/修改表结构：

```go
// internal/model/user.go
type User struct {
    ID        int64  `gorm:"primaryKey"`
    Email     string `gorm:"uniqueIndex"`
    Password  string
    // ...
}
```

GORM 会自动：
- 创建不存在的表
- 添加缺失的字段
- 创建索引

**注意：** 自动迁移不会删除字段或修改字段类型！

## 迁移状态

### 查看状态

```bash
bash migrate.sh status
```

输出示例：

```
迁移状态:
----------------------------------------
[✓] 已执行  001_add_host_id_to_servers.sql
[✓] 已执行  002_add_user_fields.sql
[✓] 已执行  003_create_user_group.sql
[ ] 待执行  004_simplify_user_group.sql
```

### 迁移记录表

迁移记录保存在 `migrations` 表中：

```sql
SELECT * FROM migrations;
```

| id | name | executed_at |
|----|------|-------------|
| 1  | 001_add_host_id_to_servers.sql | 1702345678 |
| 2  | 002_add_user_fields.sql | 1702345679 |

## 数据库重置

### 完全重置

```bash
# 警告：会删除所有数据！
bash migrate.sh reset
```

这会：
1. 删除数据库
2. 重新创建数据库
3. 执行所有迁移

### 部分重置

```bash
# 1. 回滚特定迁移
bash migrate.sh down

# 2. 删除迁移记录
mysql -u root -p xboard -e "DELETE FROM migrations WHERE name='xxx.sql';"

# 3. 重新执行
bash migrate.sh up
```

## 生产环境最佳实践

### 1. 使用 SQL 迁移

```bash
# ✅ 推荐
bash migrate.sh up

# ❌ 不推荐
bash migrate.sh auto
```

### 2. 备份数据库

```bash
# 执行迁移前备份
mysqldump -u root -p xboard > backup_$(date +%Y%m%d_%H%M%S).sql

# 执行迁移
bash migrate.sh up

# 如果出错，恢复备份
mysql -u root -p xboard < backup_20231210_120000.sql
```

### 3. 测试迁移

```bash
# 1. 在测试环境执行
bash migrate.sh up

# 2. 验证数据
mysql -u root -p xboard -e "SHOW TABLES; DESCRIBE v2_user;"

# 3. 测试应用
go run ./cmd/server

# 4. 确认无误后在生产环境执行
```

### 4. 版本控制

```bash
# 提交迁移文件到 Git
git add migrations/
git commit -m "Add migration: xxx"
git push
```

### 5. 部署流程

```bash
# 1. 拉取最新代码
git pull

# 2. 查看待执行的迁移
bash migrate.sh status

# 3. 备份数据库
mysqldump -u root -p xboard > backup.sql

# 4. 执行迁移
bash migrate.sh up

# 5. 重启应用
docker compose restart
```

## 常见问题

### Q1: 迁移执行失败怎么办？

**A:** 检查错误信息，常见原因：

1. **字段已存在**
   ```sql
   -- 使用 IF NOT EXISTS
   ALTER TABLE v2_user ADD COLUMN IF NOT EXISTS new_field VARCHAR(255);
   ```

2. **外键约束**
   ```sql
   -- 先删除外键
   ALTER TABLE v2_order DROP FOREIGN KEY fk_user_id;
   -- 再修改
   ALTER TABLE v2_user MODIFY COLUMN id BIGINT;
   -- 重新添加外键
   ALTER TABLE v2_order ADD CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES v2_user(id);
   ```

3. **数据类型不兼容**
   ```sql
   -- 先备份数据
   -- 再修改类型
   ALTER TABLE v2_user MODIFY COLUMN status INT;
   ```

### Q2: 如何跳过某个迁移？

**A:** 手动添加迁移记录：

```sql
INSERT INTO migrations (name, executed_at) VALUES ('xxx.sql', UNIX_TIMESTAMP());
```

### Q3: 如何重新执行某个迁移？

**A:** 删除迁移记录后重新执行：

```sql
DELETE FROM migrations WHERE name='xxx.sql';
```

然后：

```bash
bash migrate.sh up
```

### Q4: SQLite 和 MySQL 迁移有什么区别？

**A:** 

- **SQLite**: 使用自动迁移，不支持某些 ALTER TABLE 操作
- **MySQL**: 使用 SQL 迁移，支持完整的 DDL 语法

### Q5: 如何在 Docker 中执行迁移？

**A:** 

```bash
# 方式 1: 进入容器执行
docker compose exec xboard bash migrate.sh up

# 方式 2: 直接执行
docker compose exec xboard go run ./cmd/migrate -action up

# 方式 3: 在启动时自动执行（修改 Dockerfile）
```

## 迁移工具源码

迁移工具位于 `cmd/migrate/main.go`，支持：

- ✅ SQL 文件执行
- ✅ 迁移记录管理
- ✅ 自动迁移模型
- ✅ 迁移状态查看
- ✅ 错误处理

## 相关命令

```bash
# 迁移脚本
bash migrate.sh up          # 执行迁移
bash migrate.sh down        # 回滚迁移
bash migrate.sh status      # 查看状态
bash migrate.sh auto        # 自动迁移
bash migrate.sh reset       # 重置数据库
bash migrate.sh create xxx  # 创建迁移文件

# Makefile
make migrate                # 执行迁移
make migrate-status         # 查看状态
make migrate-auto           # 自动迁移

# Go 工具
go run ./cmd/migrate -action up      # 执行迁移
go run ./cmd/migrate -action status  # 查看状态
go run ./cmd/migrate -action auto    # 自动迁移
```

## 总结

- **开发环境**: 使用 `bash migrate.sh auto` 快速迭代
- **生产环境**: 使用 `bash migrate.sh up` 执行 SQL 迁移
- **回滚**: 使用 `bash migrate.sh down` 或手动执行回滚文件
- **备份**: 执行迁移前务必备份数据库
- **测试**: 先在测试环境验证迁移

## 下一步

- [本地安装指南](local-installation.md)
- [快速开始](../QUICK_INSTALL.md)
- [用户组设计](user-group-design.md)
