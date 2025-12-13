#!/bin/bash

# dashGO 完整编译脚本
# 用于在 Linux 上编译 Dashboard 和全架构的 Agent

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 版本信息
VERSION=${VERSION:-"1.0.0"}
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 输出目录
OUTPUT_DIR="dist"
AGENT_OUTPUT_DIR="${OUTPUT_DIR}/agent"
SERVER_OUTPUT_DIR="${OUTPUT_DIR}/server"
WEB_OUTPUT_DIR="${OUTPUT_DIR}/web"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}dashGO 完整编译脚本${NC}"
echo -e "${GREEN}版本: ${VERSION}${NC}"
echo -e "${GREEN}构建时间: ${BUILD_TIME}${NC}"
echo -e "${GREEN}Git Commit: ${GIT_COMMIT}${NC}"
echo -e "${GREEN}========================================${NC}"

# 清理旧的构建文件
clean() {
    echo -e "${YELLOW}清理旧的构建文件...${NC}"
    rm -rf ${OUTPUT_DIR}
    mkdir -p ${AGENT_OUTPUT_DIR}
    mkdir -p ${SERVER_OUTPUT_DIR}
    mkdir -p ${WEB_OUTPUT_DIR}
    echo -e "${GREEN}✓ 清理完成${NC}"
}

# 构建前端
build_frontend() {
    echo -e "${YELLOW}开始构建前端...${NC}"
    
    if [ ! -d "web/node_modules" ]; then
        echo -e "${YELLOW}安装前端依赖...${NC}"
        cd web
        npm install
        cd ..
    fi
    
    cd web
    npm run build
    cd ..
    
    # 复制构建产物
    cp -r web/dist ${WEB_OUTPUT_DIR}/
    
    echo -e "${GREEN}✓ 前端构建完成${NC}"
}

# 构建 Server (Dashboard)
build_server() {
    echo -e "${YELLOW}开始构建 Server (支持 SQLite，需要 CGO)...${NC}"
    
    # 构建 Linux amd64 (启用 CGO 以支持 SQLite)
    echo -e "${YELLOW}构建 Server (Linux amd64 with SQLite)...${NC}"
    CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
        -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
        -o ${SERVER_OUTPUT_DIR}/dashgo-server-linux-amd64 \
        ./cmd/server
    
    # 构建 Linux arm64 (启用 CGO 以支持 SQLite)
    echo -e "${YELLOW}构建 Server (Linux arm64 with SQLite)...${NC}"
    CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build \
        -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
        -o ${SERVER_OUTPUT_DIR}/dashgo-server-linux-arm64 \
        ./cmd/server
    
    # 构建 Windows amd64
    echo -e "${YELLOW}构建 Server (Windows amd64)...${NC}"
    CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build \
        -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
        -o ${SERVER_OUTPUT_DIR}/dashgo-server-windows-amd64.exe \
        ./cmd/server
    
    # 构建 macOS amd64
    echo -e "${YELLOW}构建 Server (macOS amd64)...${NC}"
    CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build \
        -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
        -o ${SERVER_OUTPUT_DIR}/dashgo-server-darwin-amd64 \
        ./cmd/server
    
    # 构建 macOS arm64 (Apple Silicon)
    echo -e "${YELLOW}构建 Server (macOS arm64)...${NC}"
    CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build \
        -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
        -o ${SERVER_OUTPUT_DIR}/dashgo-server-darwin-arm64 \
        ./cmd/server
    
    echo -e "${GREEN}✓ Server 构建完成${NC}"
}

# 构建 Agent (全架构)
build_agent() {
    echo -e "${YELLOW}开始构建 Agent (全架构)...${NC}"
    
    cd agent
    
    # Linux amd64
    echo -e "${YELLOW}构建 Agent (Linux amd64)...${NC}"
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
        -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
        -o ../${AGENT_OUTPUT_DIR}/dashgo-agent-linux-amd64 .
    
    # Linux arm64
    echo -e "${YELLOW}构建 Agent (Linux arm64)...${NC}"
    CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build \
        -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
        -o ../${AGENT_OUTPUT_DIR}/dashgo-agent-linux-arm64 .
    
    # Linux 386
    echo -e "${YELLOW}构建 Agent (Linux 386)...${NC}"
    CGO_ENABLED=0 GOOS=linux GOARCH=386 go build \
        -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
        -o ../${AGENT_OUTPUT_DIR}/dashgo-agent-linux-386 .
    
    # Windows amd64
    echo -e "${YELLOW}构建 Agent (Windows amd64)...${NC}"
    CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build \
        -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
        -o ../${AGENT_OUTPUT_DIR}/dashgo-agent-windows-amd64.exe .
    
    # Windows 386
    echo -e "${YELLOW}构建 Agent (Windows 386)...${NC}"
    CGO_ENABLED=0 GOOS=windows GOARCH=386 go build \
        -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
        -o ../${AGENT_OUTPUT_DIR}/dashgo-agent-windows-386.exe .
    
    # macOS amd64
    echo -e "${YELLOW}构建 Agent (macOS amd64)...${NC}"
    CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build \
        -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
        -o ../${AGENT_OUTPUT_DIR}/dashgo-agent-darwin-amd64 .
    
    # macOS arm64 (Apple Silicon)
    echo -e "${YELLOW}构建 Agent (macOS arm64)...${NC}"
    CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build \
        -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
        -o ../${AGENT_OUTPUT_DIR}/dashgo-agent-darwin-arm64 .
    
    # FreeBSD amd64
    echo -e "${YELLOW}构建 Agent (FreeBSD amd64)...${NC}"
    CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 go build \
        -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
        -o ../${AGENT_OUTPUT_DIR}/dashgo-agent-freebsd-amd64 .
    
    cd ..
    
    echo -e "${GREEN}✓ Agent 构建完成${NC}"
}

# 构建 Migrate 工具
build_migrate() {
    echo -e "${YELLOW}开始构建 Migrate 工具...${NC}"
    
    # Linux amd64
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
        -ldflags="-s -w" \
        -o ${SERVER_OUTPUT_DIR}/migrate-linux-amd64 \
        ./cmd/migrate
    
    # Linux arm64
    CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build \
        -ldflags="-s -w" \
        -o ${SERVER_OUTPUT_DIR}/migrate-linux-arm64 \
        ./cmd/migrate
    
    echo -e "${GREEN}✓ Migrate 工具构建完成${NC}"
}

# 生成校验和
generate_checksums() {
    echo -e "${YELLOW}生成校验和文件...${NC}"
    
    cd ${OUTPUT_DIR}
    
    # Agent 校验和
    cd agent
    sha256sum * > SHA256SUMS
    cd ..
    
    # Server 校验和
    cd server
    sha256sum * > SHA256SUMS
    cd ..
    
    cd ..
    
    echo -e "${GREEN}✓ 校验和文件生成完成${NC}"
}

# 创建版本信息文件
create_version_info() {
    echo -e "${YELLOW}创建版本信息文件...${NC}"
    
    cat > ${OUTPUT_DIR}/VERSION.txt << EOF
dashGO Build Information
========================

Version: ${VERSION}
Build Time: ${BUILD_TIME}
Git Commit: ${GIT_COMMIT}

Server Binaries:
- dashgo-server-linux-amd64
- dashgo-server-linux-arm64
- dashgo-server-windows-amd64.exe
- dashgo-server-darwin-amd64
- dashgo-server-darwin-arm64

Agent Binaries:
- dashgo-agent-linux-amd64
- dashgo-agent-linux-arm64
- dashgo-agent-linux-386
- dashgo-agent-windows-amd64.exe
- dashgo-agent-windows-386.exe
- dashgo-agent-darwin-amd64
- dashgo-agent-darwin-arm64
- dashgo-agent-freebsd-amd64

Tools:
- migrate-linux-amd64
- migrate-linux-arm64

Frontend:
- web/dist (Vue.js build)
EOF
    
    echo -e "${GREEN}✓ 版本信息文件创建完成${NC}"
}

# 显示构建结果
show_results() {
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}构建完成！${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo -e "${YELLOW}构建产物位置: ${OUTPUT_DIR}/${NC}"
    echo ""
    echo -e "${YELLOW}文件大小:${NC}"
    du -sh ${OUTPUT_DIR}/*
    echo ""
    echo -e "${YELLOW}Agent 文件:${NC}"
    ls -lh ${AGENT_OUTPUT_DIR}/
    echo ""
    echo -e "${YELLOW}Server 文件:${NC}"
    ls -lh ${SERVER_OUTPUT_DIR}/
    echo ""
    echo -e "${GREEN}所有二进制文件已生成到 ${OUTPUT_DIR} 目录${NC}"
}

# 主函数
main() {
    # 检查参数
    case "${1}" in
        clean)
            clean
            ;;
        frontend)
            clean
            build_frontend
            ;;
        server)
            clean
            build_server
            build_migrate
            ;;
        agent)
            clean
            build_agent
            ;;
        all|"")
            clean
            build_frontend
            build_server
            build_agent
            build_migrate
            generate_checksums
            create_version_info
            show_results
            ;;
        *)
            echo -e "${RED}未知参数: ${1}${NC}"
            echo ""
            echo "用法: $0 [clean|frontend|server|agent|all]"
            echo ""
            echo "  clean     - 仅清理构建文件"
            echo "  frontend  - 仅构建前端"
            echo "  server    - 仅构建 Server"
            echo "  agent     - 仅构建 Agent (全架构)"
            echo "  all       - 构建所有组件 (默认)"
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"
