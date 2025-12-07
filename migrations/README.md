# 数据库迁移脚本

本目录包含数据库迁移脚本，用于更新数据库结构。

## 迁移列表

### 001_add_host_id_to_servers.sql
- **日期**: 2024-12-07
- **描述**: 为 `v2_server` 表添加 `host_id` 字段，用于绑定节点到主机
- **变更**:
  - 添加 `host_id` 字段（BIGINT, NULL, 默认 NULL）
  - 添加 `host_id` 索引
  - 可选：添加外键约束

## 使用方法

### 执行迁移
```bash
# 连接到 MySQL 数据库
mysql -u username -p database_name < migrations/001_add_host_id_to_servers.sql
```

### 回滚迁移
```bash
mysql -u username -p database_name < migrations/001_add_host_id_to_servers_rollback.sql
```

### 使用 Docker
```bash
docker-compose exec mysql mysql -u root -p xboard < migrations/001_add_host_id_to_servers.sql
```

## 数据模型说明

### 关系说明

```
Host (主机)                    Server (节点)
┌─────────────────┐           ┌─────────────────┐
│ id              │◄──────────│ host_id         │
│ name            │           │ id              │
│ token           │           │ name            │
│ ip              │           │ type            │
│ status          │           │ host (入口地址)  │
│ ...             │           │ port (入口端口)  │
└─────────────────┘           │ server_port     │
                              │ ...             │
                              └─────────────────┘
```

- **Host（主机）**: 物理服务器，安装了 Agent 和 sing-box
- **Server（节点）**: 代理节点配置，可以绑定到某个主机上自动部署
- 一个主机可以运行多个节点
- 节点通过 `host_id` 绑定到主机

### 绑定后的效果

当 Server 绑定了 Host 后：
1. Agent 会自动在该主机上部署节点配置
2. sing-box 配置会自动同步到主机
3. 用户列表会自动同步
4. 节点的 `host` 字段是用户连接的入口地址（可以是 CDN 或中转）
5. 节点的 `server_port` 是 sing-box 实际监听的端口

### 使用场景

1. **直连模式**: 
   - `host` = 主机 IP
   - `port` = `server_port`

2. **中转模式**:
   - `host` = 中转服务器地址
   - `port` = 中转端口
   - `server_port` = 主机上 sing-box 监听端口

3. **CDN 模式**:
   - `host` = CDN 域名
   - `port` = 443
   - `server_port` = 主机上 sing-box 监听端口

## 迁移后验证

```sql
-- 查看表结构
DESC v2_server;

-- 验证字段是否添加成功
SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_DEFAULT 
FROM INFORMATION_SCHEMA.COLUMNS 
WHERE TABLE_NAME = 'v2_server' AND COLUMN_NAME = 'host_id';

-- 查看索引
SHOW INDEX FROM v2_server WHERE Key_name LIKE '%host_id%';

-- 查看绑定关系
SELECT s.id, s.name, s.type, s.host_id, h.name as host_name, h.ip as host_ip
FROM v2_server s
LEFT JOIN v2_host h ON s.host_id = h.id;
```

## 注意事项

1. **备份数据库**: 执行迁移前务必备份
2. **测试环境**: 建议先在测试环境执行
3. **权限**: 确保数据库用户有 ALTER TABLE 权限
4. **兼容性**: 未绑定主机的节点仍可正常使用（手动配置模式）
