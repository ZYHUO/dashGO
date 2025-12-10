#!/bin/bash

# XBoard 数据库升级脚本
# 用于在不清除数据的情况下升级数据库
# 用法: bash upgrade.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_FILE="${CONFIG_FILE:-configs/config.yaml}"
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

# 显示 Banner
show_banner() {
    echo -e "${BLUE}"
    cat << 'EOF'
 ██╗   ██╗██████╗  ██████╗ ██████╗  █████╗ ██████╗ ███████╗
 ██║   ██║██╔══██╗██╔════╝ ██╔══██╗██╔══██╗██╔══██╗██╔════╝
 ██║   ██║██████╔╝██║  ███╗██████╔╝███████║██║  ██║█████╗  
 ██║   ██║██╔═══╝ ██║   ██║██╔══██╗██╔══██║██║  ██║██╔══╝  
 ╚██████╔╝██║     ╚██████╔╝██║  ██║██║  ██║██████╔╝███████╗
  ╚═════╝ ╚═╝      ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═════╝ ╚══════╝
EOF
    echo -e "${NC}"
    echo -e "${GREEN}XBoard 数据库升级工具${NC}"
    echo -e "${BLUE}安全升级，保留所有数据${NC}"
    echo ""
}

# 检查配置文件
check_config() {
    if [ ! -f "$CONFIG_FILE" ]; then
        log_error "配置文件不存在: $CONFIG_FILE"
        exit 1
    fi
}

# 读取数据库配置
read_db_config() {
    DB_TYPE=$(grep "type:" "$CONFIG_FILE" | head -1 | awk '{print $2}' | tr -d '"')
    
    if [ "$DB_TYPE" = "mysql" ]; then
        DB_HOST=$(grep "host:" "$CONFIG_FILE" | grep -A 5 "database:" | grep "host:" | awk '{print $2}' | tr -d '"')
        DB_USER=$(grep "username:" "$CONFIG_FILE" | awk '{print $2}' | tr -d '"')
        DB_PASS=$(grep "password:" "$CONFIG_FILE" | grep -A 5 "database:" | grep "password:" | awk '{print $2}' | tr -d '"')
        DB_NAME=$(grep "database:" "$CONFIG_FILE" | grep -A 5 "database:" | tail -1 | awk '{print $2}' | tr -d '"')
    else
        DB_FILE=$(grep "database:" "$CONFIG_FILE" | grep -A 5 "database:" | tail -1 | awk '{print $2}' | tr -d '"')
    fi
}

# 备份数据库
backup_database() {
    log_info "备份数据库..."
    
    mkdir -p "$BACKUP_DIR"
    local timestamp=$(date +%Y%m%d_%H%M%S)
    
    if [ "$DB_TYPE" = "mysql" ]; then
        local backup_file="$BACKUP_DIR/backup_before_upgrade_${timestamp}.sql"
        
        if ! command -v mysqldump &>/dev/null; then
            log_error "未检测到 mysqldump 命令"
            log_hint "请先安装 MySQL 客户端"
            exit 1
        fi
        
        log_info "备份到: $backup_file"
        mysqldump -h"$DB_HOST" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" > "$backup_file" 2>/dev/null || {
            log_error "备份失败"
            exit 1
        }
        
        local size=$(du -h "$backup_file" | cut -f1)
        log_success "备份完成！文件大小: $size"
        echo "$backup_file" > "$BACKUP_DIR/latest_backup.txt"
        
    else
        local backup_file="$BACKUP_DIR/xboard.db.backup_${timestamp}"
        cp "$DB_FILE" "$backup_file"
        local size=$(du -h "$backup_file" | cut -f1)
        log_success "备份完成！文件大小: $size"
        echo "$backup_file" > "$BACKUP_DIR/latest_backup.txt"
    fi
}

# 检查数据完整性
check_data_integrity() {
    log_info "检查数据完整性..."
    
    if [ "$DB_TYPE" = "mysql" ]; then
        # 统计各表记录数
        local user_count=$(mysql -h"$DB_HOST" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -sN -e "SELECT COUNT(*) FROM v2_user;" 2>/dev/null || echo "0")
        local plan_count=$(mysql -h"$DB_HOST" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -sN -e "SELECT COUNT(*) FROM v2_plan;" 2>/dev/null || echo "0")
        local order_count=$(mysql -h"$DB_HOST" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -sN -e "SELECT COUNT(*) FROM v2_order;" 2>/dev/null || echo "0")
        
        log_info "当前数据统计:"
        log_info "  用户数: $user_count"
        log_info "  套餐数: $plan_count"
        log_info "  订单数: $order_count"
        
        # 保存到文件
        cat > "$BACKUP_DIR/data_before_upgrade.txt" << EOF
升级前数据统计
时间: $(date)
用户数: $user_count
套餐数: $plan_count
订单数: $order_count
EOF
    fi
}

# 显示待执行的迁移
show_pending_migrations() {
    log_info "检查待执行的迁移..."
    echo ""
    bash migrate.sh status
    echo ""
}

# 执行升级
run_upgrade() {
    log_info "开始执行数据库升级..."
    
    # 执行迁移
    bash migrate.sh up
    
    log_success "数据库升级完成！"
}

# 验证升级结果
verify_upgrade() {
    log_info "验证升级结果..."
    
    if [ "$DB_TYPE" = "mysql" ]; then
        # 统计升级后的记录数
        local user_count=$(mysql -h"$DB_HOST" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -sN -e "SELECT COUNT(*) FROM v2_user;" 2>/dev/null || echo "0")
        local plan_count=$(mysql -h"$DB_HOST" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -sN -e "SELECT COUNT(*) FROM v2_plan;" 2>/dev/null || echo "0")
        local order_count=$(mysql -h"$DB_HOST" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -sN -e "SELECT COUNT(*) FROM v2_order;" 2>/dev/null || echo "0")
        
        # 检查新表
        local group_exists=$(mysql -h"$DB_HOST" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -sN -e "SHOW TABLES LIKE 'v2_user_group';" 2>/dev/null || echo "")
        
        log_info "升级后数据统计:"
        log_info "  用户数: $user_count"
        log_info "  套餐数: $plan_count"
        log_info "  订单数: $order_count"
        
        if [ -n "$group_exists" ]; then
            local group_count=$(mysql -h"$DB_HOST" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -sN -e "SELECT COUNT(*) FROM v2_user_group;" 2>/dev/null || echo "0")
            log_info "  用户组数: $group_count"
        fi
        
        # 保存到文件
        cat > "$BACKUP_DIR/data_after_upgrade.txt" << EOF
升级后数据统计
时间: $(date)
用户数: $user_count
套餐数: $plan_count
订单数: $order_count
用户组数: ${group_count:-0}
EOF
    fi
}

# 显示升级后的配置建议
show_post_upgrade_tips() {
    echo ""
    echo "=========================================="
    echo -e "${GREEN}升级完成！${NC}"
    echo "=========================================="
    echo ""
    
    if [ "$DB_TYPE" = "mysql" ]; then
        local group_exists=$(mysql -h"$DB_HOST" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -sN -e "SHOW TABLES LIKE 'v2_user_group';" 2>/dev/null || echo "")
        
        if [ -n "$group_exists" ]; then
            echo -e "${YELLOW}重要：升级后需要配置用户组${NC}"
            echo ""
            echo "1. 配置用户组的节点权限:"
            echo "   UPDATE v2_user_group SET server_ids = '[1,2,3]' WHERE id = 2;"
            echo ""
            echo "2. 配置用户组的套餐权限:"
            echo "   UPDATE v2_user_group SET plan_ids = '[1,2,3]' WHERE id = 2;"
            echo ""
            echo "3. 配置套餐的升级组:"
            echo "   UPDATE v2_plan SET upgrade_group_id = 2 WHERE id = 1;"
            echo ""
            echo "详细说明请查看: docs/user-group-design.md"
            echo ""
        fi
    fi
    
    echo "备份文件位置:"
    if [ -f "$BACKUP_DIR/latest_backup.txt" ]; then
        cat "$BACKUP_DIR/latest_backup.txt"
    fi
    echo ""
    
    echo "数据对比:"
    if [ -f "$BACKUP_DIR/data_before_upgrade.txt" ]; then
        echo "升级前:"
        cat "$BACKUP_DIR/data_before_upgrade.txt" | grep -E "用户数|套餐数|订单数"
    fi
    if [ -f "$BACKUP_DIR/data_after_upgrade.txt" ]; then
        echo "升级后:"
        cat "$BACKUP_DIR/data_after_upgrade.txt" | grep -E "用户数|套餐数|订单数|用户组数"
    fi
    echo ""
    
    echo "下一步:"
    echo "1. 启动服务: docker compose up -d"
    echo "2. 查看日志: docker compose logs -f"
    echo "3. 测试功能: 访问后台管理"
    echo ""
    
    echo -e "${YELLOW}如果遇到问题，可以使用备份恢复:${NC}"
    if [ -f "$BACKUP_DIR/latest_backup.txt" ]; then
        local backup_file=$(cat "$BACKUP_DIR/latest_backup.txt")
        if [ "$DB_TYPE" = "mysql" ]; then
            echo "  mysql -h$DB_HOST -u$DB_USER -p$DB_PASS $DB_NAME < $backup_file"
        else
            echo "  cp $backup_file $DB_FILE"
        fi
    fi
    echo ""
}

# 主函数
main() {
    show_banner
    
    cd "$SCRIPT_DIR"
    
    # 检查配置
    check_config
    read_db_config
    
    # 显示当前配置
    log_info "数据库类型: $DB_TYPE"
    if [ "$DB_TYPE" = "mysql" ]; then
        log_info "数据库地址: $DB_HOST"
        log_info "数据库名称: $DB_NAME"
    else
        log_info "数据库文件: $DB_FILE"
    fi
    echo ""
    
    # 确认升级
    log_warn "此操作将升级数据库结构，但不会删除任何数据"
    read -p "是否继续? [y/N]: " confirm
    
    if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
        log_info "已取消升级"
        exit 0
    fi
    
    echo ""
    
    # 检查数据完整性
    check_data_integrity
    echo ""
    
    # 显示待执行的迁移
    show_pending_migrations
    
    # 再次确认
    read -p "确认执行以上迁移? [y/N]: " confirm2
    
    if [ "$confirm2" != "y" ] && [ "$confirm2" != "Y" ]; then
        log_info "已取消升级"
        exit 0
    fi
    
    echo ""
    
    # 备份数据库
    backup_database
    echo ""
    
    # 执行升级
    run_upgrade
    echo ""
    
    # 验证结果
    verify_upgrade
    echo ""
    
    # 显示升级后的提示
    show_post_upgrade_tips
}

main "$@"
