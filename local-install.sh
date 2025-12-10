#!/bin/bash

# XBoard 本地安装脚本
# 用于从 Git 克隆后的本地目录安装
# 用法: bash local-install.sh

set -e

VERSION='v1.0.0'
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INSTALL_DIR="/opt/xboard"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# 日志函数
log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_hint() { echo -e "${BLUE}[HINT]${NC} $1"; }
log_success() { echo -e "${PURPLE}[SUCCESS]${NC} $1"; }

# 显示 Banner
show_banner() {
    echo -e "${CYAN}"
    cat << 'EOF'
 ██╗  ██╗██████╗  ██████╗  █████╗ ██████╗ ██████╗ 
 ╚██╗██╔╝██╔══██╗██╔═══██╗██╔══██╗██╔══██╗██╔══██╗
  ╚███╔╝ ██████╔╝██║   ██║███████║██████╔╝██║  ██║
  ██╔██╗ ██╔══██╗██║   ██║██╔══██║██╔══██╗██║  ██║
 ██╔╝ ██╗██████╔╝╚██████╔╝██║  ██║██║  ██║██████╔╝
 ╚═╝  ╚═╝╚═════╝  ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═════╝ 
EOF
    echo -e "${NC}"
    echo -e "${GREEN}XBoard 本地安装脚本 ${VERSION}${NC}"
    echo -e "${BLUE}从本地源码安装${NC}"
    echo ""
}

# 显示菜单
show_menu() {
    echo "请选择安装选项:"
    echo ""
    echo "  1) 开发环境安装 (本地运行)"
    echo "  2) 生产环境安装 (Docker)"
    echo "  3) 编译二进制文件"
    echo "  4) 运行数据库迁移"
    echo "  5) 构建前端"
    echo "  0) 退出"
    echo ""
    read -p "请输入选项 [0-5]: " choice
    
    case $choice in
        1) install_dev ;;
        2) install_production ;;
        3) build_binary ;;
        4) run_migrations ;;
        5) build_frontend ;;
        0) exit 0 ;;
        *) log_error "无效选项"; show_menu ;;
    esac
}

# 检查 root 权限
check_root() {
    if [ "$EUID" -ne 0 ]; then
        log_warn "建议使用 root 用户运行此脚本"
        read -p "是否继续? [y/N]: " continue_install
        if [ "$continue_install" != "y" ] && [ "$continue_install" != "Y" ]; then
            exit 1
        fi
    fi
}

# 检测操作系统
detect_os() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$ID
        OS_VERSION=$VERSION_ID
    else
        OS="unknown"
    fi
    log_info "检测到系统: $OS $OS_VERSION"
}

# 检查依赖
check_dependencies() {
    log_info "检查依赖..."
    
    local missing_deps=""
    
    # 检查 Go
    if ! command -v go &>/dev/null; then
        missing_deps="$missing_deps go"
    else
        local go_version=$(go version | awk '{print $3}' | sed 's/go//')
        log_info "Go 版本: $go_version"
    fi
    
    # 检查 Node.js
    if ! command -v node &>/dev/null; then
        missing_deps="$missing_deps nodejs"
    else
        local node_version=$(node --version)
        log_info "Node.js 版本: $node_version"
    fi
    
    # 检查 npm
    if ! command -v npm &>/dev/null; then
        missing_deps="$missing_deps npm"
    fi
    
    # 检查 MySQL 客户端
    if ! command -v mysql &>/dev/null; then
        log_warn "未检测到 MySQL 客户端"
    fi
    
    if [ -n "$missing_deps" ]; then
        log_error "缺少依赖:$missing_deps"
        log_hint "请先安装缺少的依赖"
        
        case $OS in
            ubuntu|debian)
                log_hint "运行: sudo apt-get install$missing_deps"
                ;;
            centos|rhel|rocky|alma)
                log_hint "运行: sudo yum install$missing_deps"
                ;;
        esac
        
        exit 1
    fi
    
    log_success "依赖检查通过"
}

# ==================== 开发环境安装 ====================

install_dev() {
    log_info "开始安装开发环境..."
    
    cd "$SCRIPT_DIR"
    
    # 检查依赖
    check_dependencies
    
    # 创建配置文件
    if [ ! -f "configs/config.yaml" ]; then
        log_info "创建配置文件..."
        create_dev_config
    else
        log_info "配置文件已存在，跳过创建"
    fi
    
    # 安装 Go 依赖
    log_info "安装 Go 依赖..."
    go mod download
    go mod tidy
    
    # 安装前端依赖
    if [ -d "web" ]; then
        log_info "安装前端依赖..."
        cd web
        npm install
        cd ..
    fi
    
    # 运行数据库迁移
    log_info "准备数据库..."
    setup_database
    
    log_success "开发环境安装完成！"
    
    # 提示运行迁移
    echo ""
    read -p "是否立即运行数据库迁移? [Y/n]: " run_migrate
    if [ "$run_migrate" != "n" ] && [ "$run_migrate" != "N" ]; then
        log_info "运行数据库迁移..."
        bash migrate.sh auto
    fi
    
    show_dev_info
}

# 创建开发配置
create_dev_config() {
    local JWT_SECRET=$(openssl rand -base64 32 2>/dev/null || echo "dev-secret-key-change-in-production")
    
    cat > configs/config.yaml << EOF
app:
  name: "XBoard"
  url: "http://localhost:8080"
  debug: true
  jwt_secret: "${JWT_SECRET}"
  
server:
  host: "0.0.0.0"
  port: 8080
  
database:
  type: "sqlite"
  database: "xboard.db"
  
redis:
  host: "localhost"
  port: 6379
  password: ""
  database: 0
  
mail:
  driver: "smtp"
  host: "smtp.gmail.com"
  port: 587
  username: ""
  password: ""
  encryption: "tls"
  from_address: ""
  from_name: "XBoard"
  
telegram:
  bot_token: ""
  
subscribe:
  single_mode: false
EOF
    
    log_info "配置文件已创建: configs/config.yaml"
    log_hint "使用 SQLite 数据库，适合开发测试"
}

# 设置数据库
setup_database() {
    log_info "设置数据库..."
    
    # 读取配置
    local db_type=$(grep "type:" configs/config.yaml | head -1 | awk '{print $2}' | tr -d '"')
    
    if [ "$db_type" = "mysql" ]; then
        log_info "检测到 MySQL 配置"
        
        local db_host=$(grep "host:" configs/config.yaml | grep -A 5 "database:" | grep "host:" | awk '{print $2}' | tr -d '"')
        local db_user=$(grep "username:" configs/config.yaml | awk '{print $2}' | tr -d '"')
        local db_pass=$(grep "password:" configs/config.yaml | grep -A 5 "database:" | grep "password:" | awk '{print $2}' | tr -d '"')
        local db_name=$(grep "database:" configs/config.yaml | grep -A 5 "database:" | tail -1 | awk '{print $2}' | tr -d '"')
        
        log_hint "数据库信息:"
        log_hint "  主机: $db_host"
        log_hint "  用户: $db_user"
        log_hint "  数据库: $db_name"
        
        read -p "是否运行数据库迁移? [Y/n]: " run_migration
        if [ "$run_migration" != "n" ] && [ "$run_migration" != "N" ]; then
            run_migrations
        fi
    else
        log_info "使用 SQLite 数据库"
        log_hint "数据库文件将在首次运行时自动创建"
    fi
}

# 显示开发环境信息
show_dev_info() {
    echo ""
    echo "=========================================="
    echo -e "${GREEN}开发环境安装完成！${NC}"
    echo "=========================================="
    echo ""
    echo "启动命令:"
    echo ""
    echo "  # 启动后端"
    echo "  go run ./cmd/server -config configs/config.yaml"
    echo ""
    echo "  # 或使用 Makefile"
    echo "  make run"
    echo ""
    if [ -d "web" ]; then
        echo "  # 启动前端开发服务器 (另一个终端)"
        echo "  cd web && npm run dev"
        echo ""
    fi
    echo "访问地址:"
    echo "  后端 API: http://localhost:8080"
    if [ -d "web" ]; then
        echo "  前端界面: http://localhost:3000"
    fi
    echo ""
    echo "默认管理员账户:"
    echo "  邮箱: admin@xboard.local"
    echo "  密码: admin123"
    echo ""
    echo -e "${YELLOW}提示: 首次运行会自动创建数据库表${NC}"
    echo ""
}

# ==================== 生产环境安装 ====================

install_production() {
    log_info "开始安装生产环境..."
    
    cd "$SCRIPT_DIR"
    
    # 检查 Docker
    if ! command -v docker &>/dev/null; then
        log_error "未检测到 Docker"
        log_hint "请先安装 Docker: curl -fsSL https://get.docker.com | sh"
        exit 1
    fi
    
    # 创建生产配置
    if [ ! -f "configs/config.yaml" ]; then
        log_info "创建生产配置..."
        create_prod_config
    fi
    
    # 创建 .env 文件
    if [ ! -f ".env" ]; then
        create_env_file
    fi
    
    # 构建前端
    if [ -d "web" ]; then
        log_info "构建前端..."
        build_frontend
    fi
    
    # 启动 Docker 服务
    log_info "启动 Docker 服务..."
    docker compose up -d --build
    
    log_success "生产环境安装完成！"
    show_prod_info
}

# 创建生产配置
create_prod_config() {
    local DB_PASS=$(openssl rand -base64 16 | tr -dc 'a-zA-Z0-9' | head -c 16)
    local REDIS_PASS=$(openssl rand -base64 16 | tr -dc 'a-zA-Z0-9' | head -c 16)
    local JWT_SECRET=$(openssl rand -base64 32)
    
    cat > configs/config.yaml << EOF
app:
  name: "XBoard"
  url: "http://localhost:8080"
  debug: false
  jwt_secret: "${JWT_SECRET}"
  
server:
  host: "0.0.0.0"
  port: 8080
  
database:
  type: "mysql"
  host: "mysql"
  port: 3306
  username: "xboard"
  password: "${DB_PASS}"
  database: "xboard"
  
redis:
  host: "redis"
  port: 6379
  password: "${REDIS_PASS}"
  database: 0
  
mail:
  driver: "smtp"
  host: "smtp.gmail.com"
  port: 587
  username: ""
  password: ""
  encryption: "tls"
  from_address: ""
  from_name: "XBoard"
  
telegram:
  bot_token: ""
  
subscribe:
  single_mode: false
EOF
    
    # 保存密码信息
    cat > .passwords << EOF
数据库密码: ${DB_PASS}
Redis 密码: ${REDIS_PASS}
JWT Secret: ${JWT_SECRET}
EOF
    chmod 600 .passwords
    
    log_info "配置文件已创建: configs/config.yaml"
    log_hint "密码信息已保存到 .passwords 文件"
}

# 创建 .env 文件
create_env_file() {
    local DB_PASS=$(grep "password:" configs/config.yaml | grep -A 5 "database:" | grep "password:" | awk '{print $2}' | tr -d '"')
    local REDIS_PASS=$(grep "password:" configs/config.yaml | grep -A 5 "redis:" | grep "password:" | awk '{print $2}' | tr -d '"')
    
    cat > .env << EOF
MYSQL_ROOT_PASSWORD=root_${DB_PASS}
MYSQL_DATABASE=xboard
MYSQL_USER=xboard
MYSQL_PASSWORD=${DB_PASS}
REDIS_PASSWORD=${REDIS_PASS}
EOF
    
    log_info ".env 文件已创建"
}

# 显示生产环境信息
show_prod_info() {
    local IP=$(curl -s4 ip.sb 2>/dev/null || curl -s4 ifconfig.me 2>/dev/null || echo "YOUR_IP")
    
    echo ""
    echo "=========================================="
    echo -e "${GREEN}生产环境安装完成！${NC}"
    echo "=========================================="
    echo ""
    echo "访问地址: http://${IP}:80"
    echo "后台地址: http://${IP}:80/admin"
    echo ""
    echo "默认管理员账户:"
    echo "  邮箱: admin@xboard.local"
    echo "  密码: admin123"
    echo ""
    echo "常用命令:"
    echo "  查看状态: docker compose ps"
    echo "  查看日志: docker compose logs -f"
    echo "  重启服务: docker compose restart"
    echo "  停止服务: docker compose down"
    echo ""
    echo -e "${YELLOW}请及时修改默认密码！${NC}"
    echo -e "${YELLOW}密码信息保存在 .passwords 文件中${NC}"
    echo ""
}

# ==================== 编译二进制 ====================

build_binary() {
    log_info "开始编译二进制文件..."
    
    cd "$SCRIPT_DIR"
    
    # 检查 Go
    if ! command -v go &>/dev/null; then
        log_error "未检测到 Go"
        exit 1
    fi
    
    # 创建输出目录
    mkdir -p build
    
    # 编译服务器
    log_info "编译服务器..."
    go build -ldflags="-s -w" -o build/xboard ./cmd/server
    
    # 编译迁移工具
    if [ -d "cmd/migrate" ]; then
        log_info "编译迁移工具..."
        go build -ldflags="-s -w" -o build/xboard-migrate ./cmd/migrate
    fi
    
    # 编译 Agent
    if [ -d "agent" ]; then
        log_info "编译 Agent..."
        cd agent
        go build -ldflags="-s -w" -o ../build/xboard-agent .
        cd ..
    fi
    
    # 复制配置文件
    cp -r configs build/ 2>/dev/null || true
    
    log_success "编译完成！"
    echo ""
    echo "输出目录: $SCRIPT_DIR/build"
    echo ""
    echo "文件列表:"
    ls -lh build/
    echo ""
    echo "运行命令:"
    echo "  ./build/xboard -config build/configs/config.yaml"
    echo ""
}

# ==================== 数据库迁移 ====================

run_migrations() {
    log_info "运行数据库迁移..."
    
    cd "$SCRIPT_DIR"
    
    if [ ! -d "migrations" ]; then
        log_warn "未找到 migrations 目录"
        return
    fi
    
    # 读取数据库配置
    local db_type=$(grep "type:" configs/config.yaml | head -1 | awk '{print $2}' | tr -d '"')
    
    if [ "$db_type" = "mysql" ]; then
        local db_host=$(grep "host:" configs/config.yaml | grep -A 5 "database:" | grep "host:" | awk '{print $2}' | tr -d '"')
        local db_user=$(grep "username:" configs/config.yaml | awk '{print $2}' | tr -d '"')
        local db_pass=$(grep "password:" configs/config.yaml | grep -A 5 "database:" | grep "password:" | awk '{print $2}' | tr -d '"')
        local db_name=$(grep "database:" configs/config.yaml | grep -A 5 "database:" | tail -1 | awk '{print $2}' | tr -d '"')
        
        log_info "执行 MySQL 迁移..."
        
        # 检查 MySQL 客户端
        if ! command -v mysql &>/dev/null; then
            log_error "未检测到 MySQL 客户端"
            log_hint "请先安装 MySQL 客户端"
            return
        fi
        
        # 执行迁移文件
        for migration in migrations/*.sql; do
            if [[ $migration == *"rollback"* ]]; then
                continue
            fi
            
            log_info "执行: $(basename $migration)"
            mysql -h"$db_host" -u"$db_user" -p"$db_pass" "$db_name" < "$migration" 2>/dev/null || {
                log_warn "迁移失败: $(basename $migration)"
            }
        done
        
        log_success "数据库迁移完成"
    else
        log_info "SQLite 数据库将在首次运行时自动创建表"
    fi
}

# ==================== 构建前端 ====================

build_frontend() {
    log_info "构建前端..."
    
    cd "$SCRIPT_DIR"
    
    if [ ! -d "web" ]; then
        log_warn "未找到 web 目录"
        return
    fi
    
    cd web
    
    # 检查 Node.js
    if ! command -v node &>/dev/null; then
        log_error "未检测到 Node.js"
        exit 1
    fi
    
    # 安装依赖
    if [ ! -d "node_modules" ]; then
        log_info "安装前端依赖..."
        npm install
    fi
    
    # 构建
    log_info "构建前端资源..."
    npm run build
    
    cd ..
    
    log_success "前端构建完成"
    log_info "输出目录: web/dist"
}

# ==================== 主函数 ====================

main() {
    show_banner
    
    # 检查是否在项目目录
    if [ ! -f "go.mod" ]; then
        log_error "请在项目根目录运行此脚本"
        exit 1
    fi
    
    detect_os
    
    # 处理命令行参数
    case "${1:-}" in
        dev)
            install_dev
            ;;
        prod|production)
            check_root
            install_production
            ;;
        build)
            build_binary
            ;;
        migrate)
            run_migrations
            ;;
        frontend)
            build_frontend
            ;;
        -h|--help)
            echo "用法: $0 [命令]"
            echo ""
            echo "命令:"
            echo "  dev         安装开发环境"
            echo "  prod        安装生产环境 (Docker)"
            echo "  build       编译二进制文件"
            echo "  migrate     运行数据库迁移"
            echo "  frontend    构建前端"
            echo ""
            echo "示例:"
            echo "  $0 dev"
            echo "  $0 prod"
            echo "  $0 build"
            ;;
        *)
            show_menu
            ;;
    esac
}

main "$@"
