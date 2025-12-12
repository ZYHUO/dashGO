#!/bin/bash

# XBoard 前端构建脚本
# 用于单独构建前端资源

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_success() { echo -e "${BLUE}[SUCCESS]${NC} $1"; }

WEB_DIR="${1:-web}"

if [ ! -d "$WEB_DIR" ]; then
    log_error "前端目录不存在: $WEB_DIR"
    exit 1
fi

log_info "开始构建前端..."
cd "$WEB_DIR"

# 检查 Node.js
if ! command -v node &>/dev/null; then
    log_error "Node.js 未安装"
    log_info "请先安装 Node.js: https://nodejs.org/"
    exit 1
fi

log_info "Node.js 版本: $(node -v)"
log_info "npm 版本: $(npm -v)"

# 检查 package.json
if [ ! -f "package.json" ]; then
    log_error "未找到 package.json"
    exit 1
fi

# 清理旧的构建
if [ -d "dist" ]; then
    log_info "清理旧的构建产物..."
    rm -rf dist
fi

if [ -d "node_modules" ]; then
    log_warn "node_modules 已存在"
    read -p "是否重新安装依赖? [y/N]: " reinstall
    if [ "$reinstall" = "y" ] || [ "$reinstall" = "Y" ]; then
        log_info "删除 node_modules..."
        rm -rf node_modules
    fi
fi

# 安装依赖
if [ ! -d "node_modules" ]; then
    log_info "安装依赖 (这可能需要几分钟)..."
    
    # 检测包管理器
    if command -v pnpm &>/dev/null; then
        log_info "使用 pnpm..."
        pnpm install
    elif command -v yarn &>/dev/null; then
        log_info "使用 yarn..."
        yarn install
    else
        log_info "使用 npm..."
        npm install --legacy-peer-deps
    fi
    
    if [ $? -ne 0 ]; then
        log_error "依赖安装失败"
        exit 1
    fi
    log_success "依赖安装完成"
else
    log_info "依赖已安装，跳过"
fi

# 构建
log_info "开始构建..."
echo ""

if npm run build; then
    echo ""
    log_success "构建完成！"
    
    if [ -d "dist" ]; then
        log_info "构建产物: $(pwd)/dist"
        log_info "文件大小:"
        du -sh dist
        echo ""
        log_info "文件列表:"
        ls -lh dist/
    else
        log_warn "未找到 dist 目录"
    fi
else
    echo ""
    log_error "构建失败"
    exit 1
fi

echo ""
log_info "提示:"
echo "  - 如果使用 Docker，请重启容器: docker compose restart xboard"
echo "  - 如果直接运行，前端资源已更新"
echo ""
