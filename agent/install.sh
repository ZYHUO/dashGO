#!/bin/bash

# dashGO Agent 一键安装脚本
# 用法: curl -sL https://raw.githubusercontent.com/ZYHUO/dashGO/main/agent/install.sh | bash -s -- <面板地址> <Token>
# 或者: bash install.sh <面板地址> <Token>

set -e

VERSION='v1.2.0'

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 配置
PANEL_URL=$1
TOKEN=$2
GITHUB_REPO="ZYHUO/dashGO"
GH_PROXY='https://hub.glowp.xyz/'
DOWNLOAD_BASE_URL="https://download.sharon.wiki"
INSTALL_DIR="/opt/dashgo-agent"
SERVICE_NAME="dashgo-agent"
SINGBOX_DIR="/etc/sing-box"
TEMP_DIR="/tmp/dashgo-install"
SINGBOX_DEFAULT_VERSION="1.12.0"
TLS_SERVER_DEFAULT="addons.mozilla.org"

# 日志函数
log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_hint() { echo -e "${BLUE}[HINT]${NC} $1"; }

# 清理函数
cleanup() {
    rm -rf "$TEMP_DIR" >/dev/null 2>&1 || true
}
trap cleanup EXIT

# 检查 root 权限
check_root() {
    if [ "$EUID" -ne 0 ]; then
        log_error "请使用 root 用户运行此脚本"
        exit 1
    fi
}

# 检查参数
check_args() {
    if [ -z "$PANEL_URL" ] || [ -z "$TOKEN" ]; then
        echo ""
        echo "=========================================="
        echo "   dashGO Agent 安装脚本 $VERSION"
        echo "=========================================="
        echo ""
        echo "用法: $0 <面板地址> <Token>"
        echo ""
        echo "示例: $0 https://your-panel.com abc123def456"
        echo ""
        echo "参数说明:"
        echo "  面板地址: dashGO 面板的完整 URL"
        echo "  Token:    在面板后台 -> 节点管理 -> 添加节点时获取"
        echo ""
        exit 1
    fi
    
    # 去除 URL 末尾的斜杠
    PANEL_URL="${PANEL_URL%/}"
}

# 检测系统架构
detect_arch() {
    ARCH=$(uname -m)
    case $ARCH in
        x86_64|amd64)
            ARCH="amd64"
            SINGBOX_ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            SINGBOX_ARCH="arm64"
            ;;
        armv7l)
            ARCH="armv7"
            SINGBOX_ARCH="armv7"
            ;;
        *)
            log_error "不支持的架构: $ARCH"
            exit 1
            ;;
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
    
    # 设置包管理器
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
            log_warn "未知的包管理器，将尝试通用方法"
            ;;
    esac
}

# 检查 GitHub CDN 可用性
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
    log_info "检查依赖..."
    
    local deps_needed=""
    
    command -v curl >/dev/null 2>&1 || deps_needed="$deps_needed curl"
    command -v wget >/dev/null 2>&1 || deps_needed="$deps_needed wget"
    command -v tar >/dev/null 2>&1 || deps_needed="$deps_needed tar"
    
    if [ -n "$deps_needed" ]; then
        log_info "安装依赖: $deps_needed"
        $PKG_UPDATE >/dev/null 2>&1 || true
        $PKG_INSTALL $deps_needed >/dev/null 2>&1 || {
            log_error "安装依赖失败"
            exit 1
        }
    fi
}

# 获取 sing-box 最新版本
get_singbox_version() {
    local version=""
    
    # 尝试从 API 获取
    version=$(curl -s --connect-timeout 5 "https://api.github.com/repos/SagerNet/sing-box/releases/latest" 2>/dev/null | grep '"tag_name"' | sed -E 's/.*"v([^"]+)".*/\1/')
    
    if [ -z "$version" ]; then
        version="$SINGBOX_DEFAULT_VERSION"
        log_warn "无法获取最新版本，使用默认版本: $version"
    fi
    
    echo "$version"
}

# 安装 sing-box
install_singbox() {
    log_info "安装 sing-box..."
    
    # 检查是否已安装
    if command -v sing-box &>/dev/null; then
        local current_version=$(sing-box version 2>/dev/null | head -1 | awk '{print $3}')
        log_info "sing-box 已安装，版本: $current_version"
        
        # 检查版本是否需要更新
        local latest_version=$(get_singbox_version)
        if [ "$current_version" = "$latest_version" ]; then
            log_info "sing-box 已是最新版本"
            return 0
        fi
        log_info "发现新版本: $latest_version，正在更新..."
    fi
    
    mkdir -p "$TEMP_DIR"
    cd "$TEMP_DIR"
    
    local SINGBOX_VERSION=$(get_singbox_version)
    local SINGBOX_URL="${GH_PROXY}https://github.com/SagerNet/sing-box/releases/download/v${SINGBOX_VERSION}/sing-box-${SINGBOX_VERSION}-linux-${SINGBOX_ARCH}.tar.gz"
    
    log_info "下载 sing-box v${SINGBOX_VERSION}..."
    
    if ! wget -q --show-progress -O sing-box.tar.gz "$SINGBOX_URL"; then
        log_error "下载 sing-box 失败"
        exit 1
    fi
    
    tar -xzf sing-box.tar.gz
    mv sing-box-*/sing-box /usr/local/bin/
    chmod +x /usr/local/bin/sing-box
    
    # 创建配置目录
    mkdir -p "$SINGBOX_DIR"
    mkdir -p "$SINGBOX_DIR/conf"
    
    # 创建默认配置（如果不存在）
    if [ ! -f "$SINGBOX_DIR/config.json" ]; then
        create_singbox_config
    fi
    
    # 创建 systemd 服务
    create_singbox_service
    
    log_info "sing-box v${SINGBOX_VERSION} 安装完成"
}

# 创建 sing-box 默认配置
create_singbox_config() {
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
    log_info "创建 sing-box 默认配置"
}

# 创建 sing-box systemd 服务
create_singbox_service() {
    if [ "$OS" = "alpine" ]; then
        # Alpine 使用 OpenRC
        cat > /etc/init.d/sing-box << 'EOF'
#!/sbin/openrc-run

name="sing-box"
description="sing-box service"
command="/usr/local/bin/sing-box"
command_args="run -c /etc/sing-box/config.json"
pidfile="/run/${RC_SVCNAME}.pid"
command_background="yes"

depend() {
    need net
    after firewall
}
EOF
        chmod +x /etc/init.d/sing-box
        rc-update add sing-box default 2>/dev/null || true
    else
        # 其他系统使用 systemd
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
    fi
}

# 安装 Agent
install_agent() {
    log_info "安装 dashGO Agent..."
    
    mkdir -p "$INSTALL_DIR"
    mkdir -p "$TEMP_DIR"
    cd "$TEMP_DIR"
    
    # 下载 Agent
    local AGENT_URL="${DOWNLOAD_BASE_URL}/agent/dashgo-agent-linux-${ARCH}"
    
    log_info "下载 Agent (${ARCH})..."
    log_info "下载地址: $AGENT_URL"
    if wget --show-progress -O "$INSTALL_DIR/dashgo-agent" "$AGENT_URL" 2>&1; then
        log_info "Agent 下载完成"
    else
        log_warn "下载预编译版本失败，尝试从源码构建..."
        rm -f "$INSTALL_DIR/dashgo-agent" 2>/dev/null
        build_agent_from_source
    fi
    
    chmod +x "$INSTALL_DIR/dashgo-agent"
    
    log_info "Agent 安装完成"
}

# 从源码构建 Agent
build_agent_from_source() {
    if ! command -v go &>/dev/null; then
        log_info "安装 Go..."
        install_golang
    fi
    
    cd "$TEMP_DIR"
    git clone --depth 1 "https://github.com/${GITHUB_REPO}.git" dashgo 2>/dev/null || {
        git clone --depth 1 "${GH_PROXY}https://github.com/${GITHUB_REPO}.git" dashgo
    }
    
    cd dashgo/agent
    go build -ldflags="-s -w" -o "$INSTALL_DIR/dashgo-agent" .
    
    log_info "从源码构建完成"
}

# 安装 Go
install_golang() {
    local GO_VERSION="1.21.5"
    local GO_URL="https://go.dev/dl/go${GO_VERSION}.linux-${ARCH}.tar.gz"
    
    cd "$TEMP_DIR"
    wget -q -O go.tar.gz "$GO_URL"
    rm -rf /usr/local/go
    tar -C /usr/local -xzf go.tar.gz
    
    export PATH=$PATH:/usr/local/go/bin
    export GOPATH=/root/go
    
    log_info "Go ${GO_VERSION} 安装完成"
}

# 创建 Agent 服务
create_agent_service() {
    log_info "创建 Agent 服务..."
    
    if [ "$OS" = "alpine" ]; then
        # Alpine 使用 OpenRC
        cat > /etc/init.d/${SERVICE_NAME} << EOF
#!/sbin/openrc-run

name="dashgo-agent"
description="dashGO Agent service"
command="${INSTALL_DIR}/dashgo-agent"
command_args="-panel ${PANEL_URL} -token ${TOKEN} -auto-update=true -update-check-interval=3600"
pidfile="/run/\${RC_SVCNAME}.pid"
command_background="yes"
output_log="/var/log/dashgo-agent.log"
error_log="/var/log/dashgo-agent.err"

depend() {
    need net
    after sing-box
}
EOF
        chmod +x /etc/init.d/${SERVICE_NAME}
        rc-update add ${SERVICE_NAME} default 2>/dev/null || true
        rc-service ${SERVICE_NAME} start 2>/dev/null || {
            log_error "Agent 启动失败，查看日志: tail -f /var/log/dashgo-agent.err"
            exit 1
        }
    else
        # 其他系统使用 systemd
        cat > /etc/systemd/system/${SERVICE_NAME}.service << EOF
[Unit]
Description=dashGO Agent
Documentation=https://github.com/${GITHUB_REPO}
After=network.target sing-box.service

[Service]
Type=simple
ExecStart=${INSTALL_DIR}/dashgo-agent -panel ${PANEL_URL} -token ${TOKEN} -auto-update=true -update-check-interval=3600
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
LimitNOFILE=infinity

[Install]
WantedBy=multi-user.target
EOF
        systemctl daemon-reload
        systemctl enable ${SERVICE_NAME}
        systemctl start ${SERVICE_NAME}
    fi
    
    log_info "Agent 服务已启动"
}

# 显示安装结果
show_status() {
    echo ""
    echo "=========================================="
    echo -e "${GREEN}dashGO Agent 安装完成！${NC}"
    echo "=========================================="
    echo ""
    echo "版本: $VERSION"
    echo "面板地址: $PANEL_URL"
    echo "安装目录: $INSTALL_DIR"
    echo "sing-box 目录: $SINGBOX_DIR"
    echo ""
    echo "常用命令:"
    if [ "$OS" = "alpine" ]; then
        echo "  查看 Agent 状态: rc-service ${SERVICE_NAME} status"
        echo "  查看 Agent 日志: tail -f /var/log/${SERVICE_NAME}.log"
        echo "  重启 Agent: rc-service ${SERVICE_NAME} restart"
        echo "  手动触发更新: ${INSTALL_DIR}/dashgo-agent -panel ${PANEL_URL} -token ${TOKEN} -update"
        echo "  查看 sing-box 状态: rc-service sing-box status"
    else
        echo "  查看 Agent 状态: systemctl status ${SERVICE_NAME}"
        echo "  查看 Agent 日志: journalctl -u ${SERVICE_NAME} -f"
        echo "  重启 Agent: systemctl restart ${SERVICE_NAME}"
        echo "  手动触发更新: ${INSTALL_DIR}/dashgo-agent -panel ${PANEL_URL} -token ${TOKEN} -update"
        echo "  查看 sing-box 状态: systemctl status sing-box"
    fi
    echo ""
    echo "自动更新: 已启用 (每小时检查一次)"
    echo "Reality 默认 SNI: $TLS_SERVER_DEFAULT"
    echo ""
    
    # 显示服务状态
    if [ "$OS" = "alpine" ]; then
        rc-service ${SERVICE_NAME} status 2>/dev/null || true
    else
        systemctl status ${SERVICE_NAME} --no-pager 2>/dev/null || true
    fi
}

# 卸载函数
uninstall() {
    log_info "卸载 dashGO Agent..."
    
    if [ "$OS" = "alpine" ]; then
        rc-service ${SERVICE_NAME} stop 2>/dev/null || true
        rc-update del ${SERVICE_NAME} 2>/dev/null || true
        rm -f /etc/init.d/${SERVICE_NAME}
    else
        systemctl stop ${SERVICE_NAME} 2>/dev/null || true
        systemctl disable ${SERVICE_NAME} 2>/dev/null || true
        rm -f /etc/systemd/system/${SERVICE_NAME}.service
        systemctl daemon-reload
    fi
    
    rm -rf "$INSTALL_DIR"
    
    log_info "dashGO Agent 已卸载"
    
    read -p "是否同时卸载 sing-box? [y/N]: " uninstall_singbox
    if [ "$uninstall_singbox" = "y" ] || [ "$uninstall_singbox" = "Y" ]; then
        if [ "$OS" = "alpine" ]; then
            rc-service sing-box stop 2>/dev/null || true
            rc-update del sing-box 2>/dev/null || true
            rm -f /etc/init.d/sing-box
        else
            systemctl stop sing-box 2>/dev/null || true
            systemctl disable sing-box 2>/dev/null || true
            rm -f /etc/systemd/system/sing-box.service
            systemctl daemon-reload
        fi
        rm -f /usr/local/bin/sing-box
        rm -rf "$SINGBOX_DIR"
        log_info "sing-box 已卸载"
    fi
}

# 主函数
main() {
    echo ""
    echo "=========================================="
    echo "   dashGO Agent 安装脚本 $VERSION"
    echo "=========================================="
    echo ""
    
    # 处理卸载参数
    if [ "$1" = "-u" ] || [ "$1" = "--uninstall" ]; then
        check_root
        detect_os
        uninstall
        exit 0
    fi
    
    check_root
    check_args
    detect_arch
    detect_os
    check_cdn
    install_deps
    install_singbox
    install_agent
    create_agent_service
    show_status
}

main "$@"
