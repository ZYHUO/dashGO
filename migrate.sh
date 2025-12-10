#!/bin/bash

# XBoard 数据库迁移脚本
# 用法: bash migrate.sh [up|down|status|auto|reset]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_FILE="${CONFIG_FILE:-configs/config.yaml}"
MIGRATIONS_DIR="migrations"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# 显示帮助
show_help() {
    echo "XBoard 数据库迁移工具"
    echo ""
    echo "用法: $0 [命令]"
    echo ""
    echo "命令:"
    echo "  up       - 执行所有待执行的迁移（默认）"
    echo "  down     - 回滚最后一次迁移"
    echo "  status   - 查看迁移状态"
    echo "  auto     - 自动迁移模型结构（开发用）"
    echo "  reset    - 重置数据库（危险！）"
    echo "  create   - 创建新的迁移文件"
    echo ""
    echo "示例:"
    echo "  $0 up              # 执行迁移"
    echo "  $0 status          # 查看状态"
    echo "  $0 auto            # 自动迁移"
    echo "  $0 create add_xxx  # 创建迁移文件"
    echo ""
}

# 检查配置文件
check_config() {
    if [ ! -f "$CONFIG_FILE" ]; then
        log_error "配置文件不存在: $CONFIG_FILE"
        log_info "请先创建配置文件或设置 CONFIG_FILE 环境变量"
        exit 1
    fi
}

# 执行迁移
run_up() {
    log_info "执行数据库迁移..."
    
    if [ ! -d "$MIGRATIONS_DIR" ]; then
        log_error "迁移目录不存在: $MIGRATIONS_DIR"
        exit 1
    fi
    
    # 使用 Go 迁移工具
    if [ -f "cmd/migrate/main.go" ]; then
        go run ./cmd/migrate -config "$CONFIG_FILE" -action up
    else
        log_error "迁移工具不存在"
        exit 1
    fi
}

# 查看状态
run_status() {
    log_info "查看迁移状态..."
    
    if [ -f "cmd/migrate/main.go" ]; then
        go run ./cmd/migrate -config "$CONFIG_FILE" -action status
    else
        log_error "迁移工具不存在"
        exit 1
    fi
}

# 自动迁移
run_auto() {
    log_warn "自动迁移将根据模型结构修改数据库"
    read -p "是否继续? [y/N]: " confirm
    
    if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
        log_info "已取消"
        exit 0
    fi
    
    log_info "执行自动迁移..."
    
    if [ -f "cmd/migrate/main.go" ]; then
        go run ./cmd/migrate -config "$CONFIG_FILE" -action auto
    else
        log_error "迁移工具不存在"
        exit 1
    fi
}

# 回滚迁移
run_down() {
    log_warn "回滚功能需要手动执行 rollback SQL 文件"
    
    # 列出可用的 rollback 文件
    echo ""
    echo "可用的回滚文件:"
    ls -1 "$MIGRATIONS_DIR"/*_rollback.sql 2>/dev/null || {
        log_warn "没有找到回滚文件"
        exit 0
    }
    
    echo ""
    read -p "请输入要执行的回滚文件名: " filename
    
    if [ -z "$filename" ]; then
        log_info "已取消"
        exit 0
    fi
    
    local filepath="$MIGRATIONS_DIR/$filename"
    if [ ! -f "$filepath" ]; then
        log_error "文件不存在: $filepath"
        exit 1
    fi
    
    log_warn "即将执行回滚: $filename"
    read -p "确认执行? [y/N]: " confirm
    
    if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
        log_info "已取消"
        exit 0
    fi
    
    # 读取数据库配置
    local db_type=$(grep "type:" "$CONFIG_FILE" | head -1 | awk '{print $2}' | tr -d '"')
    
    if [ "$db_type" = "mysql" ]; then
        local db_host=$(grep "host:" "$CONFIG_FILE" | grep -A 5 "database:" | grep "host:" | awk '{print $2}' | tr -d '"')
        local db_user=$(grep "username:" "$CONFIG_FILE" | awk '{print $2}' | tr -d '"')
        local db_pass=$(grep "password:" "$CONFIG_FILE" | grep -A 5 "database:" | grep "password:" | awk '{print $2}' | tr -d '"')
        local db_name=$(grep "database:" "$CONFIG_FILE" | grep -A 5 "database:" | tail -1 | awk '{print $2}' | tr -d '"')
        
        log_info "执行回滚..."
        mysql -h"$db_host" -u"$db_user" -p"$db_pass" "$db_name" < "$filepath"
        log_info "回滚完成"
    else
        log_error "SQLite 不支持自动回滚，请手动执行 SQL"
        log_info "文件路径: $filepath"
    fi
}

# 重置数据库
run_reset() {
    log_error "警告：此操作将删除所有数据！"
    read -p "确认重置数据库? 输入 'yes' 继续: " confirm
    
    if [ "$confirm" != "yes" ]; then
        log_info "已取消"
        exit 0
    fi
    
    log_warn "重置数据库..."
    
    # 读取数据库配置
    local db_type=$(grep "type:" "$CONFIG_FILE" | head -1 | awk '{print $2}' | tr -d '"')
    
    if [ "$db_type" = "mysql" ]; then
        local db_host=$(grep "host:" "$CONFIG_FILE" | grep -A 5 "database:" | grep "host:" | awk '{print $2}' | tr -d '"')
        local db_user=$(grep "username:" "$CONFIG_FILE" | awk '{print $2}' | tr -d '"')
        local db_pass=$(grep "password:" "$CONFIG_FILE" | grep -A 5 "database:" | grep "password:" | awk '{print $2}' | tr -d '"')
        local db_name=$(grep "database:" "$CONFIG_FILE" | grep -A 5 "database:" | tail -1 | awk '{print $2}' | tr -d '"')
        
        log_info "删除数据库..."
        mysql -h"$db_host" -u"$db_user" -p"$db_pass" -e "DROP DATABASE IF EXISTS $db_name; CREATE DATABASE $db_name CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
        
        log_info "重新执行迁移..."
        run_up
    else
        log_info "删除 SQLite 数据库文件..."
        local db_file=$(grep "database:" "$CONFIG_FILE" | grep -A 5 "database:" | tail -1 | awk '{print $2}' | tr -d '"')
        rm -f "$db_file"
        
        log_info "重新执行迁移..."
        run_auto
    fi
    
    log_info "数据库重置完成"
}

# 创建迁移文件
run_create() {
    local name="$1"
    
    if [ -z "$name" ]; then
        read -p "请输入迁移名称 (如: add_user_field): " name
    fi
    
    if [ -z "$name" ]; then
        log_error "迁移名称不能为空"
        exit 1
    fi
    
    # 生成文件名
    local timestamp=$(date +%Y%m%d%H%M%S)
    local filename="${timestamp}_${name}.sql"
    local rollback_filename="${timestamp}_${name}_rollback.sql"
    
    # 创建迁移文件
    cat > "$MIGRATIONS_DIR/$filename" << EOF
-- Migration: $name
-- Created at: $(date)

-- 在此添加你的 SQL 语句

EOF
    
    # 创建回滚文件
    cat > "$MIGRATIONS_DIR/$rollback_filename" << EOF
-- Rollback: $name
-- Created at: $(date)

-- 在此添加回滚 SQL 语句

EOF
    
    log_info "迁移文件已创建:"
    log_info "  $MIGRATIONS_DIR/$filename"
    log_info "  $MIGRATIONS_DIR/$rollback_filename"
}

# 主函数
main() {
    cd "$SCRIPT_DIR"
    
    local action="${1:-up}"
    
    case "$action" in
        up)
            check_config
            run_up
            ;;
        down|rollback)
            check_config
            run_down
            ;;
        status)
            check_config
            run_status
            ;;
        auto)
            check_config
            run_auto
            ;;
        reset)
            check_config
            run_reset
            ;;
        create)
            run_create "$2"
            ;;
        -h|--help|help)
            show_help
            ;;
        *)
            log_error "未知命令: $action"
            show_help
            exit 1
            ;;
    esac
}

main "$@"
