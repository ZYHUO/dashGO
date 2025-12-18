#!/bin/bash

# 前端清理和重新构建脚本
# 解决新旧版本冲突问题

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}前端清理和重新构建${NC}"
echo -e "${GREEN}========================================${NC}"

# 1. 清理 dist 目录
echo -e "\n${YELLOW}[1/5] 清理 dist 目录...${NC}"
if [ -d "dist" ]; then
    rm -rf dist
    echo -e "${GREEN}✓ dist 目录已删除${NC}"
else
    echo -e "${GREEN}✓ dist 目录不存在，跳过${NC}"
fi

# 2. 清理 Vite 缓存
echo -e "\n${YELLOW}[2/5] 清理 Vite 缓存...${NC}"
if [ -d "node_modules/.vite" ]; then
    rm -rf node_modules/.vite
    echo -e "${GREEN}✓ Vite 缓存已清理${NC}"
else
    echo -e "${GREEN}✓ Vite 缓存不存在，跳过${NC}"
fi

# 3. 清理 node_modules（可选，但推荐）
echo -e "\n${YELLOW}[3/5] 清理 node_modules...${NC}"
read -p "是否清理 node_modules？这会重新安装所有依赖 (y/N): " clean_node_modules
if [ "$clean_node_modules" = "y" ] || [ "$clean_node_modules" = "Y" ]; then
    if [ -d "node_modules" ]; then
        echo -e "${YELLOW}正在删除 node_modules（这可能需要一些时间）...${NC}"
        rm -rf node_modules
        echo -e "${GREEN}✓ node_modules 已删除${NC}"
    fi
    
    echo -e "${YELLOW}正在重新安装依赖...${NC}"
    npm install
    echo -e "${GREEN}✓ 依赖安装完成${NC}"
else
    echo -e "${GREEN}✓ 跳过 node_modules 清理${NC}"
fi

# 4. 清理 package-lock.json 缓存
echo -e "\n${YELLOW}[4/5] 检查 package-lock.json...${NC}"
if [ -f "package-lock.json" ]; then
    echo -e "${GREEN}✓ package-lock.json 存在${NC}"
else
    echo -e "${YELLOW}! package-lock.json 不存在，将在安装时生成${NC}"
fi

# 5. 重新构建
echo -e "\n${YELLOW}[5/5] 重新构建前端...${NC}"
npm run build

if [ $? -eq 0 ]; then
    echo -e "\n${GREEN}========================================${NC}"
    echo -e "${GREEN}✓ 构建成功！${NC}"
    echo -e "${GREEN}========================================${NC}"
    
    # 显示构建产物信息
    if [ -d "dist" ]; then
        echo -e "\n${CYAN}构建产物信息:${NC}"
        dist_size=$(du -sh dist | cut -f1)
        file_count=$(find dist -type f | wc -l)
        echo -e "${CYAN}  大小: ${dist_size}${NC}"
        echo -e "${CYAN}  文件数: ${file_count}${NC}"
    fi
    
    echo -e "\n${YELLOW}提示:${NC}"
    echo -e "${YELLOW}  1. 如果浏览器仍显示旧版本，请按 Ctrl+Shift+R 强制刷新${NC}"
    echo -e "${YELLOW}  2. 或者清除浏览器缓存后重新访问${NC}"
else
    echo -e "\n${RED}========================================${NC}"
    echo -e "${RED}✗ 构建失败${NC}"
    echo -e "${RED}========================================${NC}"
    echo -e "${RED}请检查上面的错误信息${NC}"
    exit 1
fi
