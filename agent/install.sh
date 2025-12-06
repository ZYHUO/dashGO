#!/bin/bash

# XBoard Agent 一键安装脚本
# 用法: curl -sL https://raw.githubusercontent.com/ZYHUO/xboard-go/main/agent/install.sh | bash -s -- <面板地址> <Token>

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PANEL_URL=$1
TOKEN=$2
GITHUB_REPO="ZYHUO/xboard-go"
INSTALL_DIR="/opt/xboard-agent"
SERVICE_NAME="xboard-agent"

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

check_root() {
    if [ "$EUID" -ne 0 ]; then
        log_error "请使用 root 用户运行此脚本"
        exit 1
    fi
}

check_args() {
    if [ -z "$PANEL_URL" ] || [ -z "$TOKEN" ]; then
        echo "XBoard Agent 安装脚本"
        echo ""
        echo "用法: $0 <面板地址> <Token>"
        echo ""
        echo "示例: $0 https://your-panel.com abc123def456"
        exit 1
    fi
}

detect_arch() {
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
    log_info "检测到架构: $ARCH"
}

detect_os() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$ID
    else
        log_error "无法检测操作系统"
        exit 1
    fi
    log_info "检测到系统: $OS"
}

install_singbox() {
    log_info "安装 sing-box..."
    
    # 检查是否已安装
    if command -v sing-box &> /dev/null; then
        log_info "sing-box 已安装"
        return
    fi

    # 使用官方安装脚本
    bash <(curl -fsSL https://sing-box.app/deb-install.sh) || {
        # 备用方案：手动下载
        log_warn "官方脚本失败，尝试手动安装..."
        
        SINGBOX_VERSION=$(curl -s https://api.github.com/repos/SagerNet/sing-box/releases/latest | grep tag_name | cut -d'"' -f4 | sed 's/v//')
        SINGBOX_URL="https://github.com/SagerNet/sing-box/releases/download/v${SINGBOX_VERSION}/sing-box-${SINGBOX_VERSION}-linux-${ARCH}.tar.gz"
        
        cd /tmp
        curl -Lo sing-box.tar.gz "$SINGBOX_URL"
        tar -xzf sing-box.tar.gz
        mv sing-box-*/sing-box /usr/local/bin/
        chmod +x /usr/local/bin/sing-box
        rm -rf sing-box.tar.gz sing-box-*
    }

    # 创建配置目录
    mkdir -p /etc/sing-box
    
    log_info "sing-box 安装完成"
}

install_agent() {
    log_info "安装 XBoard Agent..."
    
    mkdir -p $INSTALL_DIR
    
    # 下载 Agent
    AGENT_URL="https://github.com/${GITHUB_REPO}/releases/download/1.1/xboard-agent-linux-${ARCH}"
    
    log_info "下载 Agent: $AGENT_URL"
    curl -Lo $INSTALL_DIR/xboard-agent "$AGENT_URL" || {
        # 备用：从源码构建
        log_warn "下载失败，尝试从源码构建..."
        
        if ! command -v go &> /dev/null; then
            log_error "需要安装 Go 来构建 Agent"
            exit 1
        fi
        
        cd /tmp
        git clone --depth 1 https://github.com/${GITHUB_REPO}.git
        cd xboard-go/agent
        go build -ldflags="-s -w" -o $INSTALL_DIR/xboard-agent .
        cd /
        rm -rf /tmp/xboard-go
    }
    
    chmod +x $INSTALL_DIR/xboard-agent
    
    log_info "Agent 安装完成"
}

create_service() {
    log_info "创建 systemd 服务..."
    
    cat > /etc/systemd/system/${SERVICE_NAME}.service << EOF
[Unit]
Description=XBoard Agent
After=network.target

[Service]
Type=simple
ExecStart=${INSTALL_DIR}/xboard-agent -panel ${PANEL_URL} -token ${TOKEN}
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable ${SERVICE_NAME}
    systemctl start ${SERVICE_NAME}
    
    log_info "服务已启动"
}

show_status() {
    echo ""
    echo "=========================================="
    echo -e "${GREEN}XBoard Agent 安装完成！${NC}"
    echo "=========================================="
    echo ""
    echo "面板地址: $PANEL_URL"
    echo "安装目录: $INSTALL_DIR"
    echo ""
    echo "常用命令:"
    echo "  查看状态: systemctl status ${SERVICE_NAME}"
    echo "  查看日志: journalctl -u ${SERVICE_NAME} -f"
    echo "  重启服务: systemctl restart ${SERVICE_NAME}"
    echo "  停止服务: systemctl stop ${SERVICE_NAME}"
    echo ""
    
    # 显示服务状态
    systemctl status ${SERVICE_NAME} --no-pager || true
}

main() {
    echo ""
    echo "=========================================="
    echo "   XBoard Agent 一键安装脚本"
    echo "=========================================="
    echo ""
    
    check_root
    check_args
    detect_arch
    detect_os
    install_singbox
    install_agent
    create_service
    show_status
}

main
