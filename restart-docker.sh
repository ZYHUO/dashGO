#!/bin/bash

# dashGO Docker 重启脚本
# 用于快速重启和查看日志

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }

# 检查配置文件
if [ ! -f "configs/config.yaml" ]; then
    log_error "配置文件不存在: configs/config.yaml"
    log_info "正在从示例文件创建..."
    
    if [ -f "configs/config.example.yaml" ]; then
        cp configs/config.example.yaml configs/config.yaml
        log_info "已创建配置文件，请根据需要修改"
    else
        log_error "示例配置文件也不存在！"
        exit 1
    fi
fi

# 创建必要的目录
mkdir -p data

log_info "停止现有容器..."
docker compose down

log_info "清理旧的镜像（可选）..."
read -p "是否重新构建镜像? [y/N]: " rebuild
if [ "$rebuild" = "y" ] || [ "$rebuild" = "Y" ]; then
    docker compose build --no-cache
fi

log_info "启动服务..."
docker compose up -d

log_info "等待服务启动..."
sleep 5

log_info "查看容器状态..."
docker compose ps

echo ""
log_info "查看 dashgo 日志 (Ctrl+C 退出)..."
echo ""
docker compose logs -f dashgo
