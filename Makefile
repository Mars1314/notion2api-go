.PHONY: help build run test clean docker-build docker-run docker-stop install deps

# 默认目标
help:
	@echo "notion-2api-go - Makefile 命令列表"
	@echo ""
	@echo "使用方法: make [target]"
	@echo ""
	@echo "可用目标:"
	@echo "  help         - 显示此帮助信息"
	@echo "  deps         - 下载 Go 依赖"
	@echo "  build        - 编译项目"
	@echo "  run          - 运行项目"
	@echo "  test         - 运行测试"
	@echo "  clean        - 清理编译文件"
	@echo "  docker-build - 构建 Docker 镜像"
	@echo "  docker-run   - 运行 Docker 容器"
	@echo "  docker-stop  - 停止 Docker 容器"
	@echo "  docker-logs  - 查看 Docker 日志"
	@echo "  install      - 安装到系统"

# 下载依赖
deps:
	@echo "下载依赖..."
	go mod download
	go mod verify

# 编译项目
build: deps
	@echo "编译项目..."
	go build -ldflags="-w -s" -o notion-2api-go .

# 运行项目
run: deps
	@echo "运行项目..."
	go run main.go

# 运行测试
test:
	@echo "运行测试..."
	go test -v ./...

# 清理编译文件
clean:
	@echo "清理编译文件..."
	rm -f notion-2api-go
	go clean

# 构建 Docker 镜像
docker-build:
	@echo "构建 Docker 镜像..."
	docker-compose build

# 运行 Docker 容器
docker-run:
	@echo "运行 Docker 容器..."
	docker-compose up -d

# 停止 Docker 容器
docker-stop:
	@echo "停止 Docker 容器..."
	docker-compose down

# 查看 Docker 日志
docker-logs:
	@echo "查看 Docker 日志..."
	docker-compose logs -f

# 安装到系统
install: build
	@echo "安装到系统..."
	sudo cp notion-2api-go /usr/local/bin/

# 格式化代码
fmt:
	@echo "格式化代码..."
	go fmt ./...

# 代码检查
lint:
	@echo "代码检查..."
	golangci-lint run

# 生成 go.sum
tidy:
	@echo "整理依赖..."
	go mod tidy