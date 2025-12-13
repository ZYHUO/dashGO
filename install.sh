#!/bin/bash

# dashGO 一键安装脚本
# 支持面板和 Agent 的完整部署
# 支持预编译二进制文件和源码构建两种方式
# 预编译文件下载地址: https://download.sharon.wiki/
# 用法: curl -sL https://raw.githubusercontent.com/ZYHUO/dashgo/main/install.sh | bash
# 或者: bash install.sh [panel|agent|all]

set -e

VERSION='v1.2.0'
GITHUB_REPO="ZYHUO/dashGO"
GH_PROXY='https://hub.glowp.xyz/'
INSTALL_DIR="/opt/dashgo"
AGENT_DIR="/opt/dashgo-agent"
TEMP_DIR="/tmp/dashgo-install"
SINGBOX_DIR="/etc/sing-box"
SINGBOX_DEFAULT_VERSION="1.12.0"

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

# 清理函数
cleanup() {
    rm -rf "$TEMP_DIR" >/dev/null 2>&1 || true
}
trap cleanup EXIT

# 显示 Banner
show_banner() {
    echo -e "${CYAN}"
    cat << 'EOF'
 ██████╗  █████╗ ███████╗██╗  ██╗ ██████╗  ██████╗ 
 ██╔══██╗██╔══██╗██╔════╝██║  ██║██╔════╝ ██╔═══██╗
 ██║  ██║███████║███████╗███████║██║  ███╗██║   ██║
 ██║  ██║██╔══██║╚════██║██╔══██║██║   ██║██║   ██║
 ██████╔╝██║  ██║███████║██║  ██║╚██████╔╝╚██████╔╝
 ╚═════╝ ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝ ╚═════╝  ╚═════╝ 
EOF
    echo -e "${NC}"
    echo -e "${GREEN}dashGO 一键安装脚本 ${VERSION}${NC}"
    echo -e "${BLUE}现代化的个人面板解决方案${NC}"
    echo ""
}

# 显示菜单
show_menu() {
    echo "请选择安装选项:"
    echo ""
    echo "  1) 安装面板 (Panel)"
    echo "  2) 安装节点 (Agent)"
    echo "  3) 完整安装 (Panel + Agent)"
    echo "  4) 卸载面板"
    echo "  5) 卸载节点"
    echo "  6) 更新面板"
    echo "  7) 更新节点"
    echo "  0) 退出"
    echo ""
    read -p "请输入选项 [0-7]: " choice
    
    case $choice in
        1) install_panel ;;
        2) install_agent_interactive ;;
        3) install_all ;;
        4) uninstall_panel ;;
        5) uninstall_agent ;;
        6) update_panel ;;
        7) update_agent ;;
        0) exit 0 ;;
        *) log_error "无效选项"; show_menu ;;
    esac
}

# 检查 root 权限
check_root() {
    if [ "$EUID" -ne 0 ]; then
        log_error "请使用 root 用户运行此脚本"
        exit 1
    fi
}

# 检测系统架构
detect_arch() {
    ARCH=$(uname -m)
    case $ARCH in
        x86_64|amd64) ARCH="amd64"; SINGBOX_ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64"; SINGBOX_ARCH="arm64" ;;
        armv7l) ARCH="armv7"; SINGBOX_ARCH="armv7" ;;
        *) log_error "不支持的架构: $ARCH"; exit 1 ;;
    esac
    log_info "检测到架构: $ARCH"
}

# 检测操作系统
detect_os() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$ID
        OS_VERSION=$VERSION_ID
    elif [ -f /etc/redhat-release ]; then
        OS="centos"
    else
        log_error "无法检测操作系统"
        exit 1
    fi
    log_info "检测到系统: $OS $OS_VERSION"
    
    case $OS in
        debian|ubuntu)
            PKG_UPDATE="apt-get update -qq"
            PKG_INSTALL="apt-get install -y -qq"
            ;;
        centos|rhel|rocky|alma|fedora)
            PKG_UPDATE="yum makecache -q"
            PKG_INSTALL="yum install -y -q"
            ;;
        alpine)
            PKG_UPDATE="apk update"
            PKG_INSTALL="apk add --no-cache"
            ;;
        *)
            log_warn "未知的包管理器"
            ;;
    esac
}

# 检查 GitHub CDN
check_cdn() {
    if wget --spider --quiet --timeout=3 "${GH_PROXY}https://raw.githubusercontent.com/SagerNet/sing-box/main/README.md" 2>/dev/null; then
        log_info "使用 GitHub 代理: $GH_PROXY"
    else
        GH_PROXY=""
        log_info "直连 GitHub"
    fi
}

# 安装依赖
install_deps() {
    log_info "安装系统依赖..."
    local deps_needed=""
    command -v curl >/dev/null 2>&1 || deps_needed="$deps_needed curl"
    command -v wget >/dev/null 2>&1 || deps_needed="$deps_needed wget"
    command -v tar >/dev/null 2>&1 || deps_needed="$deps_needed tar"
    command -v unzip >/dev/null 2>&1 || deps_needed="$deps_needed unzip"
    command -v git >/dev/null 2>&1 || deps_needed="$deps_needed git"
    command -v rsync >/dev/null 2>&1 || deps_needed="$deps_needed rsync"
    
    if [ -n "$deps_needed" ]; then
        log_info "安装: $deps_needed"
        $PKG_UPDATE >/dev/null 2>&1 || true
        $PKG_INSTALL $deps_needed >/dev/null 2>&1 || log_warn "部分依赖安装失败"
    fi
}

# 安装 Node.js 和 npm
install_nodejs() {
    if command -v node &>/dev/null && command -v npm &>/dev/null; then
        local node_version=$(node -v 2>/dev/null)
        log_info "Node.js 已安装: $node_version"
        return 0
    fi
    
    log_info "安装 Node.js 和 npm..."
    
    case $OS in
        debian|ubuntu)
            # 使用 NodeSource 仓库安装最新 LTS 版本
            curl -fsSL https://deb.nodesource.com/setup_lts.x | bash -
            $PKG_INSTALL nodejs
            ;;
        centos|rhel|rocky|alma|fedora)
            # 使用 NodeSource 仓库
            curl -fsSL https://rpm.nodesource.com/setup_lts.x | bash -
            $PKG_INSTALL nodejs
            ;;
        alpine)
            $PKG_INSTALL nodejs npm
            ;;
        *)
            log_warn "未知系统，尝试通用安装方法..."
            # 使用 nvm 安装
            curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
            export NVM_DIR="$HOME/.nvm"
            [ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
            nvm install --lts
            ;;
    esac
    
    if command -v node &>/dev/null; then
        log_success "Node.js 安装完成: $(node -v)"
        log_success "npm 安装完成: $(npm -v)"
    else
        log_error "Node.js 安装失败"
        return 1
    fi
}

# 构建前端
build_frontend() {
    local web_dir="$1"
    
    if [ ! -d "$web_dir" ]; then
        log_warn "前端目录不存在: $web_dir"
        return 0
    fi
    
    log_info "开始构建前端..."
    cd "$web_dir"
    
    # 检查是否有 package.json
    if [ ! -f "package.json" ]; then
        log_warn "未找到 package.json，跳过前端构建"
        return 0
    fi
    
    # 安装依赖
    log_info "安装前端依赖 (这可能需要几分钟)..."
    if command -v pnpm &>/dev/null; then
        log_info "使用 pnpm 安装依赖..."
        pnpm install --frozen-lockfile 2>&1 | grep -v "^npm WARN" || npm install --legacy-peer-deps
    elif command -v yarn &>/dev/null; then
        log_info "使用 yarn 安装依赖..."
        yarn install --frozen-lockfile 2>&1 | grep -v "^npm WARN" || npm install --legacy-peer-deps
    else
        log_info "使用 npm 安装依赖..."
        npm install --legacy-peer-deps 2>&1 | grep -v "^npm WARN"
    fi
    
    if [ $? -ne 0 ]; then
        log_warn "依赖安装可能有警告，继续构建..."
    fi
    
    # 构建
    log_info "构建前端资源..."
    if npm run build 2>&1 | tee /tmp/build.log | grep -v "^npm WARN"; then
        log_success "前端构建完成"
        
        # 检查构建产物
        if [ -d "dist" ]; then
            log_info "构建产物位于: $web_dir/dist"
        else
            log_warn "未找到 dist 目录，构建可能失败"
        fi
    else
        log_error "前端构建失败，查看日志: /tmp/build.log"
        log_warn "将继续安装，但前端可能无法正常显示"
    fi
    
    cd - >/dev/null
}

# 安装 Docker
install_docker() {
    if command -v docker &>/dev/null; then
        log_info "Docker 已安装"
        return 0
    fi
    
    log_info "安装 Docker..."
    curl -fsSL https://get.docker.com | sh
    systemctl enable docker
    systemctl start docker
    log_success "Docker 安装完成"
}

# 安装 Docker Compose
install_docker_compose() {
    if docker compose version &>/dev/null 2>&1; then
        log_info "Docker Compose 已安装"
        return 0
    fi
    
    log_info "安装 Docker Compose..."
    local COMPOSE_VERSION="2.24.0"
    local COMPOSE_URL="https://github.com/docker/compose/releases/download/v${COMPOSE_VERSION}/docker-compose-linux-${ARCH}"
    
    if [ -n "$GH_PROXY" ]; then
        COMPOSE_URL="${GH_PROXY}${COMPOSE_URL}"
    fi
    
    curl -L "$COMPOSE_URL" -o /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose
    log_success "Docker Compose 安装完成"
}


# ==================== 面板安装 ====================

# 安装面板
install_panel() {
    log_info "开始安装 dashGO 面板..."
    
    # 询问安装方式
    echo ""
    echo "请选择安装方式:"
    echo "  1) 使用预编译版本 (推荐)"
    echo "  2) 从源码构建"
    echo ""
    read -p "请选择 [1-2]: " install_type
    install_type=${install_type:-1}
    
    install_docker
    install_docker_compose
    
    mkdir -p "$INSTALL_DIR"
    mkdir -p "$TEMP_DIR"
    cd "$TEMP_DIR"
    
    if [ "$install_type" = "1" ]; then
        # 使用预编译版本
        log_info "下载预编译面板..."
        local PANEL_URL="https://download.sharon.wiki/server/dashgo-server-linux-${ARCH}"
        log_info "下载地址: $PANEL_URL"
        
        if wget --show-progress -O "$INSTALL_DIR/dashgo-server" "$PANEL_URL" 2>&1; then
            chmod +x "$INSTALL_DIR/dashgo-server"
            log_success "预编译面板下载完成"
        else
            log_warn "下载预编译版本失败 (HTTP错误或网络问题)"
            log_info "切换到源码构建..."
            rm -f "$INSTALL_DIR/dashgo-server" 2>/dev/null
            install_type="2"
        fi
        
        if [ "$install_type" = "1" ]; then
            # 下载配置模板和静态文件
            cd "$TEMP_DIR"
            local REPO_URL="${GH_PROXY}https://github.com/${GITHUB_REPO}/archive/refs/heads/main.zip"
            log_info "下载配置模板..."
            if wget --show-progress -O dashgo.zip "$REPO_URL" 2>&1; then
                log_info "解压配置模板..."
                unzip -q dashgo.zip
                # 自动检测解压后的目录
                local extracted_dir=$(ls -d *-main 2>/dev/null | head -1)
                if [ -n "$extracted_dir" ]; then
                    log_info "检测到解压目录: $extracted_dir"
                fi
                if [ -n "$extracted_dir" ] && [ -d "$extracted_dir" ]; then
                    cp -r "$extracted_dir/configs" "$INSTALL_DIR/" 2>/dev/null || mkdir -p "$INSTALL_DIR/configs"
                    cp -r "$extracted_dir/web/dist" "$INSTALL_DIR/web/" 2>/dev/null || mkdir -p "$INSTALL_DIR/web/dist"
                    cp "$extracted_dir/docker-compose.yaml" "$INSTALL_DIR/" 2>/dev/null || true
                    cp "$extracted_dir/Dockerfile" "$INSTALL_DIR/" 2>/dev/null || true
                    log_success "配置模板下载完成 (从 $extracted_dir)"
                else
                    log_warn "未找到解压目录 (*-main)"
                    mkdir -p "$INSTALL_DIR/configs"
                    mkdir -p "$INSTALL_DIR/web/dist"
                fi
            else
                log_warn "配置模板下载失败，将使用默认配置"
                mkdir -p "$INSTALL_DIR/configs"
                mkdir -p "$INSTALL_DIR/web/dist"
            fi
        fi
    fi
    
    if [ "$install_type" = "2" ]; then
        # 从源码构建
        log_info "从源码构建..."
        
        # 询问是否构建前端
        read -p "是否需要构建前端? (需要 Node.js) [Y/n]: " build_fe
        build_fe=${build_fe:-Y}
        
        if [ "$build_fe" = "Y" ] || [ "$build_fe" = "y" ]; then
            install_nodejs
        fi
        
        # 下载源码
        local REPO_URL="${GH_PROXY}https://github.com/${GITHUB_REPO}/archive/refs/heads/main.zip"
        
        log_info "下载源码..."
        log_info "下载地址: $REPO_URL"
        cd "$TEMP_DIR"
        if wget --show-progress -O dashgo.zip "$REPO_URL" 2>&1; then
            log_info "解压源码..."
            unzip -q dashgo.zip
            # 自动检测解压后的目录
            local extracted_dir=$(ls -d *-main 2>/dev/null | head -1)
            log_info "检测到的目录: $extracted_dir"
            if [ -n "$extracted_dir" ] && [ -d "$extracted_dir" ]; then
                cp -r "$extracted_dir"/* "$INSTALL_DIR/"
                log_success "源码下载完成 (从 $extracted_dir)"
            else
                log_error "未找到解压目录 (*-main)"
                exit 1
            fi
        else
            log_error "源码下载失败，请检查网络连接"
            exit 1
        fi
    fi
    
    cd "$INSTALL_DIR"
    
    # 处理前端构建 (仅源码构建时)
    if [ "$install_type" = "2" ]; then
        if [ "$build_fe" = "Y" ] || [ "$build_fe" = "y" ]; then
            if [ -d "web" ]; then
                build_frontend "$INSTALL_DIR/web"
            else
                log_warn "未找到 web 目录，跳过前端构建"
            fi
        else
            log_info "跳过前端构建"
            log_hint "如需前端，请手动构建: cd $INSTALL_DIR/web && npm install && npm run build"
        fi
    fi
    
    # 创建必要目录
    mkdir -p data
    mkdir -p web/dist
    
    # 创建配置文件
    if [ ! -f "configs/config.yaml" ]; then
        create_panel_config
    fi
    
    # 创建 Docker Compose 文件
    create_docker_compose
    
    # 创建初始化 SQL
    create_init_sql
    
    # 创建 Nginx 配置
    create_nginx_config
    
    # 启动服务
    log_info "启动面板服务..."
    docker compose up -d --build
    
    # 等待服务启动
    log_info "等待服务启动..."
    sleep 10
    
    # 检查服务状态
    if docker compose ps | grep -q "Up"; then
        log_success "面板安装完成！"
        show_panel_info
    else
        log_error "服务启动失败，查看日志:"
        docker compose logs --tail=50
        echo ""
        log_hint "尝试手动启动: cd $INSTALL_DIR && docker compose up -d"
    fi
}

# 创建面板配置
create_panel_config() {
    log_info "创建配置文件..."
    
    # 生成随机密码
    local DB_PASS=$(openssl rand -base64 16 | tr -dc 'a-zA-Z0-9' | head -c 16)
    local REDIS_PASS=$(openssl rand -base64 16 | tr -dc 'a-zA-Z0-9' | head -c 16)
    local JWT_SECRET=$(openssl rand -base64 32)
    local NODE_TOKEN=$(openssl rand -base64 32 | tr -dc 'a-zA-Z0-9' | head -c 32)
    
    # 创建 configs 目录
    mkdir -p configs
    
    cat > configs/config.yaml << EOF
# dashGO Configuration for Docker
app:
  name: "dashGO"
  mode: "release"
  listen: ":8080"

database:
  driver: "mysql"
  host: "mysql"
  port: 3306
  username: "root"
  password: "${DB_PASS}"
  database: "dashgo"

redis:
  host: "redis"
  port: 6379
  password: "${REDIS_PASS}"
  db: 0

jwt:
  secret: "${JWT_SECRET}"
  expire_hour: 24

node:
  token: "${NODE_TOKEN}"
  push_interval: 60
  pull_interval: 60
  enable_sync: false

mail:
  host: "smtp.example.com"
  port: 587
  username: ""
  password: ""
  from_name: "dashGO"
  from_addr: "noreply@example.com"
  encryption: "tls"

telegram:
  bot_token: ""
  chat_id: ""

admin:
  email: "admin@example.com"
  password: "admin123456"
EOF
    
    # 保存密码到环境文件
    cat > .env << EOF
MYSQL_ROOT_PASSWORD=${DB_PASS}
MYSQL_DATABASE=dashgo
REDIS_PASSWORD=${REDIS_PASS}
JWT_SECRET=${JWT_SECRET}
NODE_TOKEN=${NODE_TOKEN}
EOF
    
    log_info "配置文件已创建: configs/config.yaml"
    log_hint "数据库密码: ${DB_PASS}"
    log_hint "Redis 密码: ${REDIS_PASS}"
    log_hint "节点 Token: ${NODE_TOKEN}"
}

# 创建 Docker Compose 文件
create_docker_compose() {
    # 检查是否存在预编译二进制文件
    if [ -f "dashgo-server" ]; then
        # 为预编译版本创建简单的 Dockerfile
        cat > Dockerfile << 'EOF'
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY dashgo-server /app/dashgo-server
COPY configs/config.yaml /app/config.yaml
COPY web/dist /app/web/dist

RUN chmod +x /app/dashgo-server

EXPOSE 8080

CMD ["/app/dashgo-server"]
EOF
    fi

    cat > docker-compose.yaml << 'EOF'
services:
  dashgo:
    build: .
    container_name: dashgo
    ports:
      - "8080:8080"
    volumes:
      - ./configs/config.yaml:/app/config.yaml
      - ./data:/app/data
      - ./web/dist:/app/web/dist
    depends_on:
      mysql:
        condition: service_healthy
      redis:
        condition: service_started
    restart: unless-stopped
    environment:
      - TZ=Asia/Shanghai
    networks:
      - dashgo-net

  mysql:
    image: mysql:8.0
    container_name: dashgo-mysql
    env_file:
      - .env
    volumes:
      - mysql_data:/var/lib/mysql
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql
    ports:
      - "3306:3306"
    restart: unless-stopped
    command: --default-authentication-plugin=mysql_native_password --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - dashgo-net

  redis:
    image: redis:7-alpine
    container_name: dashgo-redis
    command: redis-server --requirepass ${REDIS_PASSWORD:-}
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"
    restart: unless-stopped
    networks:
      - dashgo-net

  nginx:
    image: nginx:alpine
    container_name: dashgo-nginx
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/nginx/ssl
    depends_on:
      - dashgo
    restart: unless-stopped
    networks:
      - dashgo-net

volumes:
  mysql_data:
  redis_data:

networks:
  dashgo-net:
    driver: bridge
EOF
}

# 创建初始化 SQL
create_init_sql() {
    cat > init.sql << 'EOF'
-- dashGO 初始化 SQL
SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- 创建管理员账户 (密码: admin123)
INSERT INTO `users` (`email`, `password`, `is_admin`, `is_staff`, `balance`, `created_at`, `updated_at`) 
VALUES ('admin@dashgo.local', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 1, 1, 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP())
ON DUPLICATE KEY UPDATE `is_admin` = 1;

SET FOREIGN_KEY_CHECKS = 1;
EOF
}

# 创建 Nginx 配置
create_nginx_config() {
    cat > nginx.conf << 'EOF'
events {
    worker_connections 1024;
}

http {
    include       /etc/nginx/mime.types;
    default_type  application/octet-stream;
    
    sendfile        on;
    keepalive_timeout  65;
    client_max_body_size 50m;
    
    # Gzip
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml;

    upstream dashgo {
        server dashgo:8080;
    }

    server {
        listen 80;
        server_name _;
        
        location / {
            proxy_pass http://dashgo;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection "upgrade";
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            proxy_connect_timeout 60s;
            proxy_send_timeout 60s;
            proxy_read_timeout 60s;
        }
    }
    
    # HTTPS 配置 (取消注释并配置证书后使用)
    # server {
    #     listen 443 ssl http2;
    #     server_name your-domain.com;
    #     
    #     ssl_certificate /etc/nginx/ssl/cert.pem;
    #     ssl_certificate_key /etc/nginx/ssl/key.pem;
    #     ssl_protocols TLSv1.2 TLSv1.3;
    #     ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256;
    #     
    #     location / {
    #         proxy_pass http://dashgo;
    #         proxy_http_version 1.1;
    #         proxy_set_header Upgrade $http_upgrade;
    #         proxy_set_header Connection "upgrade";
    #         proxy_set_header Host $host;
    #         proxy_set_header X-Real-IP $remote_addr;
    #         proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    #         proxy_set_header X-Forwarded-Proto $scheme;
    #     }
    # }
}
EOF
}

# 显示面板信息
show_panel_info() {
    local IP=$(curl -s4 ip.sb 2>/dev/null || curl -s4 ifconfig.me 2>/dev/null || echo "YOUR_IP")
    
    echo ""
    echo "=========================================="
    echo -e "${GREEN}dashGO 面板安装完成！${NC}"
    echo "=========================================="
    echo ""
    echo "访问地址: http://${IP}:80"
    echo "后台地址: http://${IP}:80/admin"
    echo ""
    echo "默认管理员账户:"
    echo "  邮箱: admin@dashgo.local"
    echo "  密码: admin123"
    echo ""
    echo "安装目录: $INSTALL_DIR"
    echo ""
    echo "常用命令:"
    echo "  查看状态: cd $INSTALL_DIR && docker compose ps"
    echo "  查看日志: cd $INSTALL_DIR && docker compose logs -f"
    echo "  重启服务: cd $INSTALL_DIR && docker compose restart"
    echo "  停止服务: cd $INSTALL_DIR && docker compose down"
    echo ""
    echo -e "${YELLOW}请及时修改默认密码！${NC}"
    echo ""
}


# ==================== Agent 安装 ====================

# 安装 sing-box
install_singbox() {
    log_info "安装 sing-box..."
    
    if command -v sing-box &>/dev/null; then
        local current_version=$(sing-box version 2>/dev/null | head -1 | awk '{print $3}')
        log_info "sing-box 已安装，版本: $current_version"
        return 0
    fi
    
    mkdir -p "$TEMP_DIR"
    cd "$TEMP_DIR"
    
    local SINGBOX_VERSION="$SINGBOX_DEFAULT_VERSION"
    local SINGBOX_URL="${GH_PROXY}https://github.com/SagerNet/sing-box/releases/download/v${SINGBOX_VERSION}/sing-box-${SINGBOX_VERSION}-linux-${SINGBOX_ARCH}.tar.gz"
    
    log_info "下载 sing-box v${SINGBOX_VERSION}..."
    
    if ! wget -q --show-progress -O sing-box.tar.gz "$SINGBOX_URL"; then
        log_error "下载 sing-box 失败"
        exit 1
    fi
    
    tar -xzf sing-box.tar.gz
    mv sing-box-*/sing-box /usr/local/bin/
    chmod +x /usr/local/bin/sing-box
    
    mkdir -p "$SINGBOX_DIR"
    mkdir -p "$SINGBOX_DIR/conf"
    
    # 创建默认配置
    if [ ! -f "$SINGBOX_DIR/config.json" ]; then
        cat > "$SINGBOX_DIR/config.json" << 'EOF'
{
    "log": {
        "level": "info",
        "timestamp": true
    },
    "inbounds": [],
    "outbounds": [
        {
            "type": "direct",
            "tag": "direct"
        }
    ]
}
EOF
    fi
    
    # 创建 systemd 服务
    cat > /etc/systemd/system/sing-box.service << 'EOF'
[Unit]
Description=sing-box service
Documentation=https://sing-box.sagernet.org
After=network.target nss-lookup.target

[Service]
Type=simple
ExecStart=/usr/local/bin/sing-box run -c /etc/sing-box/config.json
Restart=on-failure
RestartSec=10
LimitNOFILE=infinity

[Install]
WantedBy=multi-user.target
EOF
    
    systemctl daemon-reload
    log_info "sing-box v${SINGBOX_VERSION} 安装完成"
}

# 交互式安装 Agent
install_agent_interactive() {
    echo ""
    read -p "请输入面板地址 (如 https://your-panel.com): " PANEL_URL
    read -p "请输入节点 Token: " TOKEN
    
    if [ -z "$PANEL_URL" ] || [ -z "$TOKEN" ]; then
        log_error "面板地址和 Token 不能为空"
        exit 1
    fi
    
    PANEL_URL="${PANEL_URL%/}"
    install_agent "$PANEL_URL" "$TOKEN"
}

# 安装 Agent
install_agent() {
    local panel_url="${1:-$PANEL_URL}"
    local token="${2:-$TOKEN}"
    
    log_info "开始安装 dashGO Agent..."
    
    install_singbox
    
    mkdir -p "$AGENT_DIR"
    mkdir -p "$TEMP_DIR"
    cd "$TEMP_DIR"
    
    # 下载 Agent (使用新的下载地址)
    local AGENT_URL="https://download.sharon.wiki/agent/dashgo-agent-linux-${ARCH}"
    
    log_info "下载 Agent..."
    log_info "下载地址: $AGENT_URL"
    if wget --show-progress -O "$AGENT_DIR/dashgo-agent" "$AGENT_URL" 2>&1; then
        log_success "Agent 下载完成"
    else
        log_warn "下载预编译版本失败 (HTTP错误或网络问题)"
        log_info "尝试从源码构建..."
        rm -f "$AGENT_DIR/dashgo-agent" 2>/dev/null
        build_agent_from_source
    fi
    
    chmod +x "$AGENT_DIR/dashgo-agent"
    
    # 创建服务
    create_agent_service "$panel_url" "$token"
    
    log_success "Agent 安装完成！"
    show_agent_info
}

# 从源码构建 Agent
build_agent_from_source() {
    if ! command -v go &>/dev/null; then
        log_info "安装 Go..."
        local GO_VERSION="1.21.5"
        local GO_URL="https://go.dev/dl/go${GO_VERSION}.linux-${ARCH}.tar.gz"
        
        cd "$TEMP_DIR"
        wget -q -O go.tar.gz "$GO_URL"
        rm -rf /usr/local/go
        tar -C /usr/local -xzf go.tar.gz
        export PATH=$PATH:/usr/local/go/bin
        export GOPATH=/root/go
    fi
    
    cd "$TEMP_DIR"
    git clone --depth 1 "${GH_PROXY}https://github.com/${GITHUB_REPO}.git" dashgo 2>/dev/null || \
    git clone --depth 1 "https://github.com/${GITHUB_REPO}.git" dashgo
    
    cd dashgo/agent
    go build -ldflags="-s -w" -o "$AGENT_DIR/dashgo-agent" .
    
    log_info "从源码构建完成"
}

# 创建 Agent 服务
create_agent_service() {
    local panel_url="$1"
    local token="$2"
    
    log_info "创建 Agent 服务..."
    
    cat > /etc/systemd/system/dashgo-agent.service << EOF
[Unit]
Description=dashGO Agent
Documentation=https://github.com/${GITHUB_REPO}
After=network.target sing-box.service

[Service]
Type=simple
ExecStart=${AGENT_DIR}/dashgo-agent -panel ${panel_url} -token ${token}
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
LimitNOFILE=infinity

[Install]
WantedBy=multi-user.target
EOF
    
    systemctl daemon-reload
    systemctl enable dashgo-agent
    systemctl start dashgo-agent
    
    log_info "Agent 服务已启动"
}

# 显示 Agent 信息
show_agent_info() {
    echo ""
    echo "=========================================="
    echo -e "${GREEN}dashGO Agent 安装完成！${NC}"
    echo "=========================================="
    echo ""
    echo "安装目录: $AGENT_DIR"
    echo "sing-box 目录: $SINGBOX_DIR"
    echo ""
    echo "常用命令:"
    echo "  查看 Agent 状态: systemctl status dashgo-agent"
    echo "  查看 Agent 日志: journalctl -u dashgo-agent -f"
    echo "  重启 Agent: systemctl restart dashgo-agent"
    echo "  查看 sing-box 状态: systemctl status sing-box"
    echo "  查看 sing-box 日志: journalctl -u sing-box -f"
    echo ""
    
    systemctl status dashgo-agent --no-pager 2>/dev/null || true
}


# ==================== 完整安装 ====================

# 完整安装 (面板 + Agent)
install_all() {
    log_info "开始完整安装..."
    
    install_panel
    
    echo ""
    read -p "是否在本机同时安装 Agent? [y/N]: " install_local_agent
    
    if [ "$install_local_agent" = "y" ] || [ "$install_local_agent" = "Y" ]; then
        local IP=$(curl -s4 ip.sb 2>/dev/null || curl -s4 ifconfig.me 2>/dev/null || echo "localhost")
        
        echo ""
        log_hint "请先在面板后台添加节点并获取 Token"
        read -p "请输入节点 Token: " TOKEN
        
        if [ -n "$TOKEN" ]; then
            install_agent "http://${IP}:8080" "$TOKEN"
        else
            log_warn "未输入 Token，跳过 Agent 安装"
            log_hint "稍后可以运行: bash install.sh agent"
        fi
    fi
    
    log_success "安装完成！"
}

# ==================== 卸载功能 ====================

# 卸载面板
uninstall_panel() {
    log_info "卸载 dashGO 面板..."
    
    if [ -d "$INSTALL_DIR" ]; then
        cd "$INSTALL_DIR"
        docker compose down -v 2>/dev/null || true
        cd /
        
        read -p "是否删除数据目录? [y/N]: " delete_data
        if [ "$delete_data" = "y" ] || [ "$delete_data" = "Y" ]; then
            rm -rf "$INSTALL_DIR"
            log_info "数据目录已删除"
        else
            log_info "保留数据目录: $INSTALL_DIR"
        fi
    fi
    
    log_success "面板已卸载"
}

# 卸载 Agent
uninstall_agent() {
    log_info "卸载 dashGO Agent..."
    
    systemctl stop dashgo-agent 2>/dev/null || true
    systemctl disable dashgo-agent 2>/dev/null || true
    rm -f /etc/systemd/system/dashgo-agent.service
    systemctl daemon-reload
    
    rm -rf "$AGENT_DIR"
    
    log_info "Agent 已卸载"
    
    read -p "是否同时卸载 sing-box? [y/N]: " uninstall_sb
    if [ "$uninstall_sb" = "y" ] || [ "$uninstall_sb" = "Y" ]; then
        systemctl stop sing-box 2>/dev/null || true
        systemctl disable sing-box 2>/dev/null || true
        rm -f /etc/systemd/system/sing-box.service
        systemctl daemon-reload
        rm -f /usr/local/bin/sing-box
        rm -rf "$SINGBOX_DIR"
        log_info "sing-box 已卸载"
    fi
    
    log_success "Agent 已卸载"
}

# ==================== 更新功能 ====================

# 更新面板
update_panel() {
    log_info "更新 dashGO 面板..."
    
    if [ ! -d "$INSTALL_DIR" ]; then
        log_error "面板未安装"
        exit 1
    fi
    
    # 询问更新方式
    echo ""
    echo "请选择更新方式:"
    echo "  1) 使用预编译版本 (推荐)"
    echo "  2) 从源码更新"
    echo ""
    read -p "请选择 [1-2]: " update_type
    update_type=${update_type:-1}
    
    cd "$INSTALL_DIR"
    
    # 备份配置和数据
    log_info "备份配置文件..."
    cp configs/config.yaml configs/config.yaml.bak 2>/dev/null || cp config.yaml config.yaml.bak 2>/dev/null || true
    cp .env .env.bak 2>/dev/null || true
    
    # 备份前端构建产物
    if [ -d "web/dist" ]; then
        log_info "备份前端构建产物..."
        mv web/dist web/dist.bak
    fi
    
    # 停止服务
    log_info "停止服务..."
    docker compose down
    
    if [ "$update_type" = "1" ]; then
        # 使用预编译版本更新
        log_info "下载预编译面板..."
        mkdir -p "$TEMP_DIR"
        cd "$TEMP_DIR"
        
        local PANEL_URL="https://download.sharon.wiki/server/dashgo-server-linux-${ARCH}"
        log_info "下载地址: $PANEL_URL"
        
        if wget --show-progress -O dashgo-server "$PANEL_URL" 2>&1; then
            # 备份旧版本
            mv "$INSTALL_DIR/dashgo-server" "$INSTALL_DIR/dashgo-server.bak" 2>/dev/null || true
            # 安装新版本
            mv dashgo-server "$INSTALL_DIR/dashgo-server"
            chmod +x "$INSTALL_DIR/dashgo-server"
            log_success "预编译面板更新完成"
        else
            log_warn "下载预编译版本失败 (HTTP错误或网络问题)"
            log_info "切换到源码更新..."
            rm -f dashgo-server 2>/dev/null
            update_type="2"
        fi
    fi
    
    if [ "$update_type" = "2" ]; then
        # 从源码更新
        # 询问是否重新构建前端
        read -p "是否重新构建前端? [y/N]: " rebuild_fe
        
        # 下载新版本
        mkdir -p "$TEMP_DIR"
        cd "$TEMP_DIR"
        
        log_info "下载最新版本..."
        local REPO_URL="${GH_PROXY}https://github.com/${GITHUB_REPO}/archive/refs/heads/main.zip"
        log_info "下载地址: $REPO_URL"
        if wget --show-progress -O dashgo.zip "$REPO_URL" 2>&1; then
            log_info "解压源码..."
            unzip -q dashgo.zip
            # 自动检测解压后的目录
            local extracted_dir=$(ls -d *-main 2>/dev/null | head -1)
            log_info "检测到的目录: $extracted_dir"
            if [ -n "$extracted_dir" ] && [ -d "$extracted_dir" ]; then
                log_success "源码下载完成 (从 $extracted_dir)"
                # 更新文件 (保留配置和数据)
                log_info "更新文件..."
                rsync -av --exclude='configs/config.yaml' --exclude='config.yaml' --exclude='.env' \
                    --exclude='data' --exclude='storage' --exclude='ssl' --exclude='web/dist' \
                    "$extracted_dir"/* "$INSTALL_DIR/"
            else
                log_error "未找到解压目录 (*-main)"
                exit 1
            fi
        else
            log_error "源码下载失败，请检查网络连接"
            exit 1
        fi
        
        cd "$INSTALL_DIR"
        
        # 处理前端
        if [ "$rebuild_fe" = "y" ] || [ "$rebuild_fe" = "Y" ]; then
            log_info "重新构建前端..."
            install_nodejs
            if [ -d "web" ]; then
                build_frontend "$INSTALL_DIR/web"
            fi
        else
            # 恢复旧的前端构建
            if [ -d "web/dist.bak" ]; then
                log_info "恢复旧的前端构建..."
                mv web/dist.bak web/dist
            else
                log_warn "未找到前端构建产物"
            fi
        fi
    fi
    
    cd "$INSTALL_DIR"
    
    # 恢复配置
    log_info "恢复配置..."
    if [ -f "configs/config.yaml.bak" ]; then
        mv configs/config.yaml.bak configs/config.yaml
    elif [ -f "config.yaml.bak" ]; then
        mkdir -p configs
        mv config.yaml.bak configs/config.yaml
    fi
    mv .env.bak .env 2>/dev/null || true
    
    # 重新构建并启动
    log_info "重新启动服务..."
    docker compose up -d --build
    
    # 等待服务启动
    sleep 5
    
    # 检查状态
    if docker compose ps | grep -q "Up"; then
        log_success "面板更新完成！"
    else
        log_error "服务启动失败，查看日志:"
        docker compose logs --tail=30
    fi
}

# 更新 Agent
update_agent() {
    log_info "更新 dashGO Agent..."
    
    if [ ! -f "$AGENT_DIR/dashgo-agent" ]; then
        log_error "Agent 未安装"
        exit 1
    fi
    
    # 停止服务
    systemctl stop dashgo-agent
    
    # 下载新版本
    mkdir -p "$TEMP_DIR"
    cd "$TEMP_DIR"
    
    local AGENT_URL="https://download.sharon.wiki/agent/dashgo-agent-linux-${ARCH}"
    log_info "下载地址: $AGENT_URL"
    
    if wget --show-progress -O "$AGENT_DIR/dashgo-agent.new" "$AGENT_URL" 2>&1; then
        mv "$AGENT_DIR/dashgo-agent.new" "$AGENT_DIR/dashgo-agent"
        chmod +x "$AGENT_DIR/dashgo-agent"
        log_success "Agent 更新完成"
    else
        log_warn "下载失败 (HTTP错误或网络问题)"
        log_info "尝试从源码构建..."
        rm -f "$AGENT_DIR/dashgo-agent.new" 2>/dev/null
        build_agent_from_source
    fi
    
    # 重启服务
    systemctl start dashgo-agent
    
    log_success "Agent 更新完成！"
}

# ==================== 主函数 ====================

main() {
    show_banner
    check_root
    detect_arch
    detect_os
    check_cdn
    install_deps
    
    # 处理命令行参数
    case "${1:-}" in
        panel)
            install_panel
            ;;
        agent)
            if [ -n "$2" ] && [ -n "$3" ]; then
                install_agent "$2" "$3"
            else
                install_agent_interactive
            fi
            ;;
        all)
            install_all
            ;;
        uninstall-panel)
            uninstall_panel
            ;;
        uninstall-agent)
            uninstall_agent
            ;;
        update-panel)
            update_panel
            ;;
        update-agent)
            update_agent
            ;;
        -h|--help)
            echo "用法: $0 [命令]"
            echo ""
            echo "命令:"
            echo "  panel              安装面板"
            echo "  agent [url] [token] 安装节点"
            echo "  all                完整安装"
            echo "  uninstall-panel    卸载面板"
            echo "  uninstall-agent    卸载节点"
            echo "  update-panel       更新面板"
            echo "  update-agent       更新节点"
            echo ""
            echo "示例:"
            echo "  $0 panel"
            echo "  $0 agent https://panel.example.com abc123"
            echo "  $0 all"
            ;;
        *)
            show_menu
            ;;
    esac
}

main "$@"
