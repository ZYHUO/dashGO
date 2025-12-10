# 数据库迁移快速指南

## 🚀 快速命令

```bash
# 执行迁移
bash migrate.sh up

# 查看状态
bash migrate.sh status

# 自动迁移（开发环境）
bash migrate.sh auto

# 回滚
bash migrate.sh down

# 创建新迁移
bash migrate.sh create add_new_field

# 升级现有数据库（不清除数据）
bash upgrade.sh
```

## 📋 使用 Makefile

```bash
make migrate              # 执行迁移
make migrate-status       # 查看状态
make migrate-auto         # 自动迁移
make migrate-down         # 回滚
make migrate-reset        # 重置数据库
make migrate-create name=xxx  # 创建迁移
```

## 🔄 迁移流程

### 开发环境

```bash
# 1. 修改模型
vim internal/model/user.go

# 2. 自动迁移
bash migrate.sh auto

# 3. 测试
go run ./cmd/server
```

### 生产环境

```bash
# 1. 创建迁移文件
bash migrate.sh create add_user_field

# 2. 编写 SQL
vim migrations/xxx_add_user_field.sql

# 3. 测试迁移
bash migrate.sh up

# 4. 提交代码
git add migrations/
git commit -m "Add migration: add_user_field"
```

## 📝 迁移文件示例

### 创建表

```sql
-- migrations/001_create_users.sql
CREATE TABLE IF NOT EXISTS v2_user (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
) COMMENT='用户表';
```

### 添加字段

```sql
-- migrations/002_add_user_fields.sql
ALTER TABLE v2_user ADD COLUMN IF NOT EXISTS phone VARCHAR(20) COMMENT '手机号';
ALTER TABLE v2_user ADD COLUMN IF NOT EXISTS avatar VARCHAR(255) COMMENT '头像';
```

### 修改字段

```sql
-- migrations/003_modify_user_email.sql
ALTER TABLE v2_user MODIFY COLUMN email VARCHAR(320) NOT NULL COMMENT '邮箱（支持更长的邮箱地址）';
```

### 添加索引

```sql
-- migrations/004_add_user_indexes.sql
CREATE INDEX idx_user_phone ON v2_user(phone);
CREATE INDEX idx_user_created_at ON v2_user(created_at);
```

### 回滚文件

```sql
-- migrations/002_add_user_fields_rollback.sql
ALTER TABLE v2_user DROP COLUMN IF EXISTS phone;
ALTER TABLE v2_user DROP COLUMN IF EXISTS avatar;
```

## ⚠️ 注意事项

### 生产环境

- ✅ 使用 SQL 迁移
- ✅ 执行前备份数据库
- ✅ 先在测试环境验证
- ❌ 不要使用自动迁移

### 开发环境

- ✅ 可以使用自动迁移
- ✅ 快速迭代
- ⚠️ 自动迁移不会删除字段

## 🔧 常用操作

### 查看迁移状态

```bash
bash migrate.sh status
```

### 备份数据库

```bash
# MySQL
mysqldump -u root -p xboard > backup_$(date +%Y%m%d).sql

# 恢复
mysql -u root -p xboard < backup_20231210.sql
```

### 重置数据库

```bash
# 警告：会删除所有数据！
bash migrate.sh reset
```

### 跳过某个迁移

```sql
-- 手动标记为已执行
INSERT INTO migrations (name, executed_at) 
VALUES ('xxx.sql', UNIX_TIMESTAMP());
```

### 重新执行迁移

```sql
-- 删除迁移记录
DELETE FROM migrations WHERE name='xxx.sql';
```

然后重新执行：

```bash
bash migrate.sh up
```

## 📚 详细文档

查看完整文档：[docs/database-migration.md](docs/database-migration.md)

## 🆘 常见问题

### Q: 迁移失败怎么办？

1. 查看错误信息
2. 检查 SQL 语法
3. 确认数据库连接
4. 查看迁移记录表

### Q: 如何回滚？

```bash
bash migrate.sh down
```

### Q: 自动迁移和 SQL 迁移的区别？

| 特性 | 自动迁移 | SQL 迁移 |
|------|---------|---------|
| 适用环境 | 开发 | 生产 |
| 精确控制 | ❌ | ✅ |
| 删除字段 | ❌ | ✅ |
| 回滚支持 | ❌ | ✅ |
| 速度 | 快 | 慢 |

### Q: 如何在 Docker 中执行？

```bash
docker compose exec xboard bash migrate.sh up
```

## 🎯 最佳实践

1. **版本控制**: 所有迁移文件提交到 Git
2. **命名规范**: 使用序号和描述性名称
3. **测试优先**: 先在测试环境验证
4. **备份数据**: 执行前务必备份
5. **原子操作**: 每个迁移只做一件事
6. **可回滚**: 提供回滚文件

## 📞 获取帮助

```bash
bash migrate.sh --help
```

或查看：
- [完整迁移文档](docs/database-migration.md)
- [本地安装指南](docs/local-installation.md)
- [快速开始](QUICK_INSTALL.md)


## 🔄 升级现有数据库

如果你有旧版本的数据库，需要在**不清除数据**的情况下升级：

### 使用升级脚本（推荐）

```bash
bash upgrade.sh
```

升级脚本会自动：
1. 备份数据库
2. 检查数据完整性
3. 显示待执行的迁移
4. 执行迁移
5. 验证升级结果
6. 提供配置建议

### 手动升级

```bash
# 1. 备份数据库
mysqldump -u root -p xboard > backup_$(date +%Y%m%d).sql

# 2. 查看待执行的迁移
bash migrate.sh status

# 3. 执行迁移
bash migrate.sh up

# 4. 验证数据
mysql -u root -p xboard -e "SELECT COUNT(*) FROM v2_user;"
```

### 升级特点

- ✅ 保留所有现有数据
- ✅ 只执行新的迁移
- ✅ 自动跳过已执行的迁移
- ✅ 支持回滚
- ✅ 安全可靠

**详细升级指南：** [docs/upgrade-guide.md](docs/upgrade-guide.md)
