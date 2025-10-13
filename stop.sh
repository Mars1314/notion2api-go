#!/bin/bash

# notion-2api 停止脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_header() {
    echo ""
    echo -e "${BLUE}╔════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║${NC}  ${GREEN}notion-2api${NC} 停止脚本              ${BLUE}║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════╝${NC}"
    echo ""
}

command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# 停止 Docker 服务
stop_docker() {
    print_info "停止 Docker 容器..."
    if docker-compose ps 2>/dev/null | grep -q "Up"; then
        docker-compose down
        print_success "Docker 容器已停止"
        return 0
    else
        print_warning "未检测到运行中的 Docker 容器"
        return 1
    fi
}

# 停止本地进程
stop_local() {
    print_info "停止本地进程..."
    if pgrep -f "notion-2api" > /dev/null 2>&1; then
        pkill -f "notion-2api"
        sleep 1
        if pgrep -f "notion-2api" > /dev/null 2>&1; then
            print_warning "进程未能正常停止，使用强制停止..."
            pkill -9 -f "notion-2api"
        fi
        print_success "本地进程已停止"
        return 0
    else
        print_warning "未检测到运行中的本地进程"
        return 1
    fi
}

main() {
    print_header
    
    stopped=0
    
    # 尝试停止 Docker
    if command_exists docker-compose; then
        if stop_docker; then
            stopped=1
        fi
    fi
    
    # 尝试停止本地进程
    if stop_local; then
        stopped=1
    fi
    
    if [ $stopped -eq 0 ]; then
        print_warning "未检测到运行中的服务"
    else
        echo ""
        print_success "✅ 服务已完全停止"
    fi
    
    echo ""
}

main