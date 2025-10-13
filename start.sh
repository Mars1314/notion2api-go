#!/bin/bash

# notion-2api 一键启动脚本
# 自动检测环境并选择最佳启动方式

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
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
    echo -e "${BLUE}║${NC}  ${GREEN}notion-2api${NC} 一键启动脚本           ${BLUE}║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════╝${NC}"
    echo ""
}

# 检查命令是否存在
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# 检查 .env 文件
check_env_file() {
    if [ ! -f ".env" ]; then
        print_warning ".env 文件不存在，正在从 .env.example 创建..."
        if [ -f ".env.example" ]; then
            cp .env.example .env
            print_success ".env 文件已创建"
            print_warning "请编辑 .env 文件并填入您的 Notion 凭证："
            echo ""
            echo "  必填项："
            echo "    - NOTION_COOKIE"
            echo "    - NOTION_SPACE_ID"
            echo "    - NOTION_USER_ID"
            echo ""
            read -p "按 Enter 键打开编辑器..."
            
            # 尝试使用不同的编辑器
            if command_exists nano; then
                nano .env
            elif command_exists vim; then
                vim .env
            elif command_exists vi; then
                vi .env
            else
                print_error "未找到文本编辑器，请手动编辑 .env 文件"
                exit 1
            fi
        else
            print_error ".env.example 文件不存在"
            exit 1
        fi
    fi
    
    # 验证必需的配置项
    if ! grep -q "NOTION_COOKIE=" .env || ! grep -q "NOTION_SPACE_ID=" .env || ! grep -q "NOTION_USER_ID=" .env; then
        print_error ".env 文件缺少必需的配置项"
        exit 1
    fi
    
    print_success ".env 配置文件检查通过"
}

# Docker 方式启动
start_with_docker() {
    print_info "使用 Docker Compose 启动服务..."
    
    if docker-compose ps 2>/dev/null | grep -q "Up"; then
        print_warning "服务已经在运行，正在重启..."
        docker-compose restart
    else
        print_info "构建并启动容器..."
        docker-compose up -d --build
    fi
    
    print_success "服务已启动"
    
    # 等待服务就绪
    print_info "等待服务就绪..."
    sleep 3
    
    # 检查服务状态
    if docker-compose ps 2>/dev/null | grep -q "Up"; then
        print_success "服务运行正常"
        
        # 获取端口
        PORT=$(grep NGINX_PORT .env 2>/dev/null | cut -d'=' -f2 | tr -d ' "' || echo "8004")
        
        echo ""
        print_success "🎉 启动成功！"
        echo ""
        echo -e "  服务地址: ${GREEN}http://localhost:${PORT}${NC}"
        echo -e "  文档地址: ${GREEN}http://localhost:${PORT}/docs${NC}"
        echo -e "  健康检查: ${GREEN}http://localhost:${PORT}/${NC}"
        echo -e "  查看日志: ${YELLOW}docker-compose logs -f${NC}"
        echo -e "  停止服务: ${YELLOW}docker-compose down${NC}"
        echo ""

        # 显示支持的模型
        echo ""
        echo -e "${BLUE}📋 支持的模型：${NC}"
        echo -e "  ${GREEN}✓${NC} claude-sonnet-4.5  ${YELLOW}(推荐)${NC}"
        echo -e "  ${GREEN}✓${NC} gpt-5"
        echo -e "  ${GREEN}✓${NC} claude-opus-4.1"
        echo -e "  ${GREEN}✓${NC} gpt-4.1"
        echo -e "  ${GREEN}✓${NC} gemini-2.5-flash"
        echo -e "  ${GREEN}✓${NC} gemini-2.5-pro"
    else
        print_error "服务启动失败"
        print_info "查看日志: docker-compose logs"
        exit 1
    fi
}

# Go 本地方式启动
start_with_go() {
    print_info "使用 Go 本地启动服务..."
    
    # 检查是否已经编译
    if [ ! -f "./notion-2api" ]; then
        print_info "首次运行，正在编译..."
        go build -o notion-2api .
        print_success "编译完成"
    fi
    
    # 检查是否已有进程在运行
    if pgrep -f "notion-2api" > /dev/null 2>&1; then
        print_warning "检测到已有进程在运行，正在重启..."
        pkill -f "notion-2api" 2>/dev/null || true
        sleep 1
    fi
    
    # 启动服务
    print_info "启动服务..."
    nohup ./notion-2api > notion-2api.log 2>&1 &
    
    # 等待服务就绪
    sleep 2
    
    # 获取端口
    PORT=$(grep NGINX_PORT .env 2>/dev/null | cut -d'=' -f2 | tr -d ' "' || echo "8004")
    
    if pgrep -f "notion-2api" > /dev/null 2>&1; then
        print_success "🎉 启动成功！"
        echo ""
        echo -e "  服务地址: ${GREEN}http://localhost:${PORT}${NC}"
        echo -e "  文档地址: ${GREEN}http://localhost:${PORT}/docs${NC}"
        echo -e "  进程 PID: ${YELLOW}$(pgrep -f notion-2api)${NC}"
        echo -e "  日志文件: ${YELLOW}tail -f notion-2api.log${NC}"
        echo -e "  停止服务: ${YELLOW}pkill -f notion-2api${NC}"
        echo ""

        # 显示支持的模型
        echo ""
        echo -e "${BLUE}📋 支持的模型：${NC}"
        echo -e "  ${GREEN}✓${NC} claude-sonnet-4.5  ${YELLOW}(推荐)${NC}"
        echo -e "  ${GREEN}✓${NC} gpt-5"
        echo -e "  ${GREEN}✓${NC} claude-opus-4.1"
        echo -e "  ${GREEN}✓${NC} gpt-4.1"
        echo -e "  ${GREEN}✓${NC} gemini-2.5-flash"
        echo -e "  ${GREEN}✓${NC} gemini-2.5-pro"
    else
        print_error "服务启动失败，查看日志: tail -f notion-2api.log"
        exit 1
    fi
}

# 主函数
main() {
    print_header
    
    # 检查 .env 文件
    check_env_file
    
    # 检测启动方式
    if command_exists docker-compose || command_exists docker; then
        if command_exists docker-compose; then
            print_info "检测到 Docker Compose，使用容器方式启动"
            start_with_docker
        else
            print_warning "检测到 Docker 但没有 Docker Compose"
            if command_exists go; then
                print_info "切换到 Go 本地启动方式"
                start_with_go
            else
                print_error "请安装 Docker Compose 或 Go 1.21+"
                exit 1
            fi
        fi
    elif command_exists go; then
        print_info "检测到 Go 环境，使用本地方式启动"
        start_with_go
    else
        print_error "未检测到 Docker 或 Go 环境"
        print_info "请安装以下工具之一："
        echo "  - Docker + Docker Compose"
        echo "  - Go 1.21+"
        exit 1
    fi
    
    echo ""
    echo -e "${BLUE}ℹ${NC} 提示: 使用 ${YELLOW}./stop.sh${NC} 可以停止服务"
}

# 运行主函数
main