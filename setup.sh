#!/bin/bash

# dashGO 一键安装/升级/修复脚本
# 支持：全新安装、现有数据库安装、升级、修复迁移问题

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_FILE="configs/config.yaml"
BACKUP_DIR="backups"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_success() { echo -e "${PURPLE}[SUCCESS]${NC} $1"; }

# 下载预编译二进制文件
DOWNLOAD_BASE_URL="https://download.sharon.wiki"

download_binaries() {
    log_info "下载预编译二进制文件..."
    
    # 检测架构
    ARCH=$(uname -m)
    case $ARCH in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            log_error "不支持的架构: $ARCH"
            exit 1
            ;;
    esac
    
    # 下载 server
    if [ ! -f "dashgo-server-linux-${ARCH}" ]; then
        log_info "下载 dashgo-server-linux-${ARCH}..."
        if ! wget -q --show-progress "${DOWNLOAD_BASE_URL}/server/dashgo-server-linux-${ARCH}" -O "dashgo-server-linux-${ARCH}"; then
            log_error "下载 server 失败"
            exit 1
        fi
        chmod +x "dashgo-server-linux-${ARCH}"
    fi
    
    # 下载 migrate
    if [ ! -f "migrate-linux-${ARCH}" ]; then
        log_info "下载 migrate-linux-${ARCH}..."
        if ! wget -q --show-progress "${DOWNLOAD_BASE_URL}/server/migrate-linux-${ARCH}" -O "migrate-linux-${ARCH}"; then
            log_error "下载 migrate 失败"
            exit 1
        fi
        chmod +x "migrate-linux-${ARCH}"
    fi
    
    # 创建符号链接
    ln -sf "dashgo-server-linux-${ARCH}" dashgo-server
    ln -sf "migrate-linux-${ARCH}" migrate
    
    log_success "二进制文件下载完成"
}

# 显示 Banner
show_banner() {
    clear
    echo -e "${BLUE}"
    cat << 'EOF'
 ██╗  ██╗██████╗  ██████╗  █████╗ ██████╗ ██████╗        ██████╗  ██████╗ 
 ╚██╗██╔╝██╔══██╗██╔═══██╗██╔══██╗██╔══██╗██╔══██╗      ██╔════╝ ██╔═══██╗
  ╚███╔╝ ██████╔╝██║   ██║███████║██████╔╝██║  ██║█████╗██║  ███╗██║   ██║
  ██╔██╗ ██╔══██╗██║   ██║██╔══██║██╔══██╗██║  ██║╚════╝██║   ██║██║   ██║
 ██╔╝ ██╗██████╔╝╚██████╔╝██║  ██║██║  ██║██████╔╝      ╚██████╔╝╚██████╔╝
 ╚═╝  ╚═╝╚═════╝  ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═════╝        ╚═════╝  ╚═════╝ 
EOF
    echo -e "${NC}"
    echo -e "${GREEN}dashGO 一键安装/升级/修复工具${NC}"
    echo -e "${BLUE}https://github.com/ZYHUO/dashGO${NC}"
    echo ""
}

# 主菜单
show_menu() {
    echo "请选择操作:"
    echo ""
    echo "  ${GREEN}1${NC}) 全新安装 (本地开发)"
    echo "  ${GREEN}2${NC}) 安装到现有 MySQL 数据库"
    echo "  ${GREEN}3${NC}) 升级现有数据库"
    echo "  ${GREEN}4${NC}) 修复迁移问题"
    echo "  ${GREEN}5${NC}) 查看迁移状态"
    echo "  ${GREEN}6${NC}) 生成配置文件"
    echo "  ${GREEN}0${NC}) 退出"
    echo ""
    read -p "请输入选项 [0-6]: " choice
    echo ""
}

# 1. 全新安装 (本地开发)
install_local() {
    log_info "开始本地开发环境安装..."
    
    # 检查依赖
    if ! command -v go &>/dev/null; then
        log_error "未检测到 Go，请先安装 Go 1.21+"
        exit 1
    fi
    
    # 选择数据库类型
    echo "选择数据库类型:"
    echo "  1) SQLite (推荐，无需配置)"
    echo "  2) MySQL"
    read -p "请选择 [1-2]: " db_choice
    
    if [ "$db_choice" = "1" ]; then
        # SQLite
        log_info "使用 SQLite 数据库..."
        cat > "$CONFIG_FILE" <<EOF
app:
  name: "dashGO"
  mode: "debug"
  listen: ":8080"

database:
  driver: "sqlite"
  database: "dashgo.db"

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

jwt:
  secret: "$(openssl rand -hex 32)"
  expire_hour: 24

node:
  token: "$(openssl rand -hex 32)"
  push_interval: 60
  pull_interval: 60
  enable_sync: false

admin:
  email: "admin@example.com"
  password: "admin123456"
EOF
    else
        # MySQL
        read -p "MySQL 主机 (默认: localhost): " mysql_host
        mysql_host=${mysql_host:-localhost}
        
        read -p "MySQL 端口 (默认: 3306): " mysql_port
        mysql_port=${mysql_port:-3306}
        
        read -p "数据库名 (默认: dashgo): " mysql_db
        mysql_db=${mysql_db:-dashgo}
        
        read -p "用户名 (默认: root): " mysql_user
        mysql_user=${mysql_user:-root}
        
        read -sp "密码: " mysql_pass
        echo ""
        
        cat > "$CONFIG_FILE" <<EOF
app:
  name: "dashGO"
  mode: "debug"
  listen: ":8080"

database:
  driver: "mysql"
  host: "$mysql_host"
  port: $mysql_port
  database: "$mysql_db"
  username: "$mysql_user"
  password: "$mysql_pass"

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

jwt:
  secret: "$(openssl rand -hex 32)"
  expire_hour: 24

node:
  token: "$(openssl rand -hex 32)"
  push_interval: 60
  pull_interval: 60
  enable_sync: false

admin:
  email: "admin@example.com"
  password: "admin123456"
EOF
    fi
    
    log_success "配置文件已生成: $CONFIG_FILE"
    
    # 下载二进制文件
    download_binaries
    
    # 运行迁移
    log_info "运行数据库迁移..."
    ./migrate -action up
    
    log_success "安装完成！"
    echo ""
    echo "启动服务:"
    echo "  ./dashgo-server"
    echo ""
    echo "默认管理员账号:"
    echo "  邮箱: admin@example.com"
    echo "  密码: admin123456"
    echo ""
}

# 2. 安装到现有 MySQL 数据库
install_existing_db() {
    log_info "安装到现有 MySQL 数据库..."
    
    # 输入数据库信息
    read -p "MySQL 主机 (默认: localhost): " DB_HOST
    DB_HOST=${DB_HOST:-localhost}
    
    read -p "MySQL 端口 (默认: 3306): " DB_PORT
    DB_PORT=${DB_PORT:-3306}
    
    read -p "数据库名: " DB_NAME
    if [ -z "$DB_NAME" ]; then
        log_error "数据库名不能为空"
        exit 1
    fi
    
    read -p "用户名 (默认: root): " DB_USER
    DB_USER=${DB_USER:-root}
    
    read -sp "密码: " DB_PASS
    echo ""
    
    # 测试连接
    log_info "测试数据库连接..."
    if ! mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -e "SELECT 1;" &>/dev/null; then
        log_error "无法连接到数据库"
        exit 1
    fi
    log_success "数据库连接成功"
    
    # 生成配置文件
    log_info "生成配置文件..."
    cat > "$CONFIG_FILE" <<EOF
app:
  name: "dashGO"
  mode: "release"
  listen: ":8080"

database:
  driver: "mysql"
  host: "$DB_HOST"
  port: $DB_PORT
  database: "$DB_NAME"
  username: "$DB_USER"
  password: "$DB_PASS"

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

jwt:
  secret: "$(openssl rand -hex 32)"
  expire_hour: 24

node:
  token: "$(openssl rand -hex 32)"
  push_interval: 60
  pull_interval: 60
  enable_sync: false

admin:
  email: "admin@example.com"
  password: "admin123456"
EOF
    
    log_success "配置文件已生成"
    
    # 下载二进制文件
    download_binaries
    
    # 运行迁移
    log_info "运行数据库迁移..."
    ./migrate -action up
    
    log_success "安装完成！"
    echo ""
    echo "启动服务:"
    echo "  ./dashgo-server"
    echo ""
}

# 3. 升级现有数据库
upgrade_database() {
    log_info "升级现有数据库..."
    
    # 检查配置文件
    if [ ! -f "$CONFIG_FILE" ]; then
        log_error "配置文件不存在: $CONFIG_FILE"
        log_info "请先运行选项 6 生成配置文件"
        exit 1
    fi
    
    # 读取数据库配置
    DB_DRIVER=$(grep "driver:" "$CONFIG_FILE" | head -1 | awk '{print $2}' | tr -d '"')
    
    if [ "$DB_DRIVER" = "mysql" ]; then
        DB_HOST=$(grep "host:" "$CONFIG_FILE" | grep -A 5 "database:" | grep "host:" | awk '{print $2}' | tr -d '"')
        DB_USER=$(grep "username:" "$CONFIG_FILE" | awk '{print $2}' | tr -d '"')
        DB_PASS=$(grep "password:" "$CONFIG_FILE" | grep -A 5 "database:" | grep "password:" | awk '{print $2}' | tr -d '"')
        DB_NAME=$(grep "database:" "$CONFIG_FILE" | grep -A 5 "database:" | tail -1 | awk '{print $2}' | tr -d '"')
        
        log_info "数据库类型: MySQL"
        log_info "数据库地址: $DB_HOST"
        log_info "数据库名称: $DB_NAME"
    else
        log_info "数据库类型: SQLite"
    fi
    
    echo ""
    log_warn "此操作将升级数据库结构，但不会删除任何数据"
    read -p "是否继续? [y/N]: " confirm
    
    if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
        log_info "已取消升级"
        exit 0
    fi
    
    # 备份数据库
    if [ "$DB_DRIVER" = "mysql" ]; then
        log_info "备份数据库..."
        mkdir -p "$BACKUP_DIR"
        local timestamp=$(date +%Y%m%d_%H%M%S)
        local backup_file="$BACKUP_DIR/backup_before_upgrade_${timestamp}.sql"
        
        mysqldump -h"$DB_HOST" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" > "$backup_file" 2>/dev/null || {
            log_error "备份失败"
            exit 1
        }
        
        log_success "备份完成: $backup_file"
    fi
    
    # 运行迁移
    log_info "运行数据库迁移..."
    ./migrate -action up
    
    log_success "升级完成！"
}

# 4. 修复迁移问题
fix_migration() {
    log_info "修复迁移问题..."
    
    # 检查配置文件
    if [ ! -f "$CONFIG_FILE" ]; then
        log_error "配置文件不存在: $CONFIG_FILE"
        exit 1
    fi
    
    # 读取数据库配置
    DB_DRIVER=$(grep "driver:" "$CONFIG_FILE" | head -1 | awk '{print $2}' | tr -d '"')
    
    if [ "$DB_DRIVER" != "mysql" ]; then
        log_error "此功能仅支持 MySQL 数据库"
        exit 1
    fi
    
    DB_HOST=$(grep "host:" "$CONFIG_FILE" | grep -A 5 "database:" | grep "host:" | awk '{print $2}' | tr -d '"')
    DB_PORT=$(grep "port:" "$CONFIG_FILE" | grep -A 5 "database:" | grep "port:" | awk '{print $2}')
    DB_USER=$(grep "username:" "$CONFIG_FILE" | awk '{print $2}' | tr -d '"')
    DB_PASS=$(grep "password:" "$CONFIG_FILE" | grep -A 5 "database:" | grep "password:" | awk '{print $2}' | tr -d '"')
    DB_NAME=$(grep "database:" "$CONFIG_FILE" | grep -A 5 "database:" | tail -1 | awk '{print $2}' | tr -d '"')
    
    log_info "数据库: $DB_NAME @ $DB_HOST:$DB_PORT"
    
    # 1. 清理错误的迁移记录
    log_info "清理错误的迁移记录..."
    mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" <<EOF
DELETE FROM migrations WHERE name LIKE '%_rollback%';
EOF
    log_success "清理完成"
    
    # 2. 检查并添加 host_id 字段
    log_info "检查 v2_server 表结构..."
    HAS_HOST_ID=$(mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -sN -e "SHOW COLUMNS FROM v2_server LIKE 'host_id';" | wc -l)
    
    if [ "$HAS_HOST_ID" -eq 0 ]; then
        log_warn "host_id 字段不存在，正在添加..."
        mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" <<EOF
ALTER TABLE \`v2_server\` 
ADD COLUMN \`host_id\` BIGINT NULL DEFAULT NULL COMMENT '绑定的主机ID' AFTER \`parent_id\`;

ALTER TABLE \`v2_server\` 
ADD INDEX \`idx_server_host_id\` (\`host_id\`);
EOF
        log_success "host_id 字段已添加"
    else
        log_success "host_id 字段已存在"
    fi
    
    # 3. 检查并添加 sold_count 字段
    log_info "检查 v2_plan 表结构..."
    HAS_SOLD_COUNT=$(mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -sN -e "SHOW COLUMNS FROM v2_plan LIKE 'sold_count';" | wc -l)
    
    if [ "$HAS_SOLD_COUNT" -eq 0 ]; then
        log_warn "sold_count 字段不存在，正在添加..."
        mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" <<EOF
ALTER TABLE \`v2_plan\` ADD COLUMN \`sold_count\` INT NOT NULL DEFAULT 0 COMMENT '已售出数量';

UPDATE \`v2_plan\` p 
SET \`sold_count\` = (
    SELECT COUNT(*) 
    FROM \`v2_user\` u 
    WHERE u.\`plan_id\` = p.\`id\`
);

CREATE INDEX \`idx_plan_capacity\` ON \`v2_plan\`(\`capacity_limit\`, \`sold_count\`);
EOF
        log_success "sold_count 字段已添加"
    else
        log_success "sold_count 字段已存在"
    fi
    
    # 4. 检查并添加 socks_outbound 字段
    log_info "检查 v2_host 表结构..."
    HAS_SOCKS_OUTBOUND=$(mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -sN -e "SHOW COLUMNS FROM v2_host LIKE 'socks_outbound';" | wc -l)
    
    if [ "$HAS_SOCKS_OUTBOUND" -eq 0 ]; then
        log_warn "socks_outbound 字段不存在，正在添加..."
        mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" <<EOF
ALTER TABLE \`v2_host\` 
ADD COLUMN \`socks_outbound\` TEXT NULL COMMENT 'SOCKS5 出口代理地址，格式：socks5://user:pass@host:port';
EOF
        log_success "socks_outbound 字段已添加"
    else
        log_success "socks_outbound 字段已存在"
    fi
    
    log_success "修复完成！"
}

# 5. 查看迁移状态
show_migration_status() {
    log_info "查看迁移状态..."
    ./migrate -action status
}

# 6. 生成配置文件
generate_config() {
    log_info "生成配置文件..."
    
    if [ -f "$CONFIG_FILE" ]; then
        log_warn "配置文件已存在: $CONFIG_FILE"
        read -p "是否覆盖? [y/N]: " confirm
        if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
            log_info "已取消"
            return
        fi
    fi
    
    # 选择数据库类型
    echo "选择数据库类型:"
    echo "  1) MySQL"
    echo "  2) SQLite"
    read -p "请选择 [1-2]: " db_choice
    
    if [ "$db_choice" = "1" ]; then
        # MySQL
        read -p "MySQL 主机 (默认: localhost): " mysql_host
        mysql_host=${mysql_host:-localhost}
        
        read -p "MySQL 端口 (默认: 3306): " mysql_port
        mysql_port=${mysql_port:-3306}
        
        read -p "数据库名: " mysql_db
        if [ -z "$mysql_db" ]; then
            log_error "数据库名不能为空"
            return
        fi
        
        read -p "用户名 (默认: root): " mysql_user
        mysql_user=${mysql_user:-root}
        
        read -sp "密码: " mysql_pass
        echo ""
        
        cat > "$CONFIG_FILE" <<EOF
app:
  name: "dashGO"
  mode: "release"
  listen: ":8080"

database:
  driver: "mysql"
  host: "$mysql_host"
  port: $mysql_port
  database: "$mysql_db"
  username: "$mysql_user"
  password: "$mysql_pass"

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

jwt:
  secret: "$(openssl rand -hex 32)"
  expire_hour: 24

node:
  token: "$(openssl rand -hex 32)"
  push_interval: 60
  pull_interval: 60
  enable_sync: false

admin:
  email: "admin@example.com"
  password: "admin123456"
EOF
    else
        # SQLite
        cat > "$CONFIG_FILE" <<EOF
app:
  name: "dashGO"
  mode: "release"
  listen: ":8080"

database:
  driver: "sqlite"
  database: "dashgo.db"

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

jwt:
  secret: "$(openssl rand -hex 32)"
  expire_hour: 24

node:
  token: "$(openssl rand -hex 32)"
  push_interval: 60
  pull_interval: 60
  enable_sync: false

admin:
  email: "admin@example.com"
  password: "admin123456"
EOF
    fi
    
    log_success "配置文件已生成: $CONFIG_FILE"
}

# 主程序
main() {
    show_banner
    
    while true; do
        show_menu
        
        case $choice in
            1)
                install_local
                ;;
            2)
                install_existing_db
                ;;
            3)
                upgrade_database
                ;;
            4)
                fix_migration
                ;;
            5)
                show_migration_status
                ;;
            6)
                generate_config
                ;;
            0)
                log_info "退出"
                exit 0
                ;;
            *)
                log_error "无效选项"
                ;;
        esac
        
        echo ""
        read -p "按回车键继续..."
    done
}

# 运行主程序
main "$@"
