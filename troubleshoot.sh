#!/bin/bash

# dashGO 故障排查脚本

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}dashGO 故障排查工具${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# 1. 检查配置文件
echo -e "${GREEN}[1] 检查配置文件${NC}"
if [ -f "configs/config.yaml" ]; then
    echo -e "  ✓ configs/config.yaml 存在"
else
    echo -e "  ${RED}✗ configs/config.yaml 不存在${NC}"
    echo -e "  ${YELLOW}解决方案: cp configs/config.example.yaml configs/config.yaml${NC}"
fi
echo ""

# 2. 检查目录
echo -e "${GREEN}[2] 检查必要目录${NC}"
for dir in data configs web/dist; do
    if [ -d "$dir" ]; then
        echo -e "  ✓ $dir 存在"
    else
        echo -e "  ${YELLOW}! $dir 不存在 (将自动创建)${NC}"
        mkdir -p "$dir"
    fi
done
echo ""

# 3. 检查 Docker
echo -e "${GREEN}[3] 检查 Docker 服务${NC}"
if command -v docker &>/dev/null; then
    echo -e "  ✓ Docker 已安装"
    docker --version
else
    echo -e "  ${RED}✗ Docker 未安装${NC}"
fi

if docker compose version &>/dev/null 2>&1; then
    echo -e "  ✓ Docker Compose 已安装"
    docker compose version
else
    echo -e "  ${RED}✗ Docker Compose 未安装${NC}"
fi
echo ""

# 4. 检查容器状态
echo -e "${GREEN}[4] 检查容器状态${NC}"
if docker compose ps &>/dev/null 2>&1; then
    docker compose ps
else
    echo -e "  ${YELLOW}! 没有运行的容器${NC}"
fi
echo ""

# 5. 查看最近的日志
echo -e "${GREEN}[5] 查看 dashgo 最近日志${NC}"
if docker ps -a | grep -q dashgo; then
    docker compose logs --tail=50 dashgo
else
    echo -e "  ${YELLOW}! dashgo 容器不存在${NC}"
fi
echo ""

# 6. 检查端口占用
echo -e "${GREEN}[6] 检查端口占用${NC}"
for port in 8080 3306 6379; do
    if command -v netstat &>/dev/null; then
        if netstat -tuln | grep -q ":$port "; then
            echo -e "  ${YELLOW}! 端口 $port 已被占用${NC}"
        else
            echo -e "  ✓ 端口 $port 可用"
        fi
    elif command -v ss &>/dev/null; then
        if ss -tuln | grep -q ":$port "; then
            echo -e "  ${YELLOW}! 端口 $port 已被占用${NC}"
        else
            echo -e "  ✓ 端口 $port 可用"
        fi
    fi
done
echo ""

# 7. 数据库连接测试
echo -e "${GREEN}[7] 测试数据库连接${NC}"
if docker ps | grep -q dashgo-mysql; then
    if docker exec dashgo-mysql mysqladmin ping -h localhost -pdashgo_password &>/dev/null; then
        echo -e "  ✓ MySQL 连接正常"
    else
        echo -e "  ${RED}✗ MySQL 连接失败${NC}"
    fi
else
    echo -e "  ${YELLOW}! MySQL 容器未运行${NC}"
fi
echo ""

# 8. Redis 连接测试
echo -e "${GREEN}[8] 测试 Redis 连接${NC}"
if docker ps | grep -q dashgo-redis; then
    if docker exec dashgo-redis redis-cli ping &>/dev/null; then
        echo -e "  ✓ Redis 连接正常"
    else
        echo -e "  ${RED}✗ Redis 连接失败${NC}"
    fi
else
    echo -e "  ${YELLOW}! Redis 容器未运行${NC}"
fi
echo ""

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}常用修复命令:${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo "1. 重启所有服务:"
echo "   docker compose restart"
echo ""
echo "2. 完全重建:"
echo "   docker compose down && docker compose up -d --build"
echo ""
echo "3. 查看实时日志:"
echo "   docker compose logs -f dashgo"
echo ""
echo "4. 进入容器调试:"
echo "   docker exec -it dashgo sh"
echo ""
echo "5. 清理并重新开始:"
echo "   docker compose down -v && docker compose up -d"
echo ""
