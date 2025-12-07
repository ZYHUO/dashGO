# 数据库迁移

## 迁移方式

### 方式一：使用迁移工具（推荐）

```bash
# 查看迁移状态
make migrate-status

# 执行 SQL 迁移
make migrate

# 自动迁移模型结构（仅添加新字段，不执行 SQL）
make migrate-auto
```

或者直接运行：

```bash
# 执行迁移
go run ./cmd/migrate -action up

# 查看状态
go run ./cmd/migrate -action status

# 自动迁移
go run ./cmd/migrate -action auto
```

### 方式二：手动执行 SQL

```bash
# MySQL
mysql -u root -p xboard < migrations/002_add_user_fields.sql

# 或登录 MySQL 后
mysql> source /path/to/migrations/002_add_user_fields.sql
```

### 方式三：程序自动迁移

程序启动时会自动通过 GORM 的 `AutoMigrate` 更新表结构，但不会执行 SQL 文件中的数据插入。

## 迁移文件说明

| 文件 | 说明 |
|------|------|
| 001_add_host_id_to_servers.sql | 添加主机绑定字段 |
| 002_add_user_fields.sql | 添加用户字段和站点设置 |

## 创建新迁移

1. 在 `migrations/` 目录创建新的 SQL 文件
2. 文件名格式：`XXX_description.sql`（XXX 为序号）
3. 运行 `make migrate` 执行迁移

## 注意事项

- 迁移工具会记录已执行的迁移，不会重复执行
- `AutoMigrate` 只会添加新字段，不会删除或修改现有字段
- 生产环境建议先备份数据库再执行迁移
