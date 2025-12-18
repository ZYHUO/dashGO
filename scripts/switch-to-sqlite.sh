#!/bin/bash

# 切换到 SQLite 数据库

echo "切换 dashGO 到 SQLite 数据库..."

# 1. 停止服务
docker-compose down

# 2. 备份配置
cp configs/config.yaml configs/config.yaml.mysql.backup

# 3. 修改配置为 SQLite
cat > configs/config.yaml << 'EOF'
# dashGO Configuration (SQLite)
app:
  name: "dashGO"
  mode: "release"
  listen: ":8080"

database:
  driver: "sqlite"
  dsn: "data/dashgo.db"

redis:
  host: "dashgo-redis"
  port: 6379
  password: ""
  db: 0

jwt:
  secret: "your-jwt-secret-change-this"
  expire_hour: 24

node:
  token: "your-node-token-change-this"
  push_interval: 60
  pull_interval: 60
  enable_sync: false

admin:
  email: "admin@example.com"
  password: "admin123456"
EOF

# 4. 确保数据目录存在
mkdir -p data
chmod 755 data

# 5. 重启服务
docker-compose up -d

echo ""
echo "已切换到 SQLite 数据库"
echo "查看日志: docker-compose logs -f dashgo"
