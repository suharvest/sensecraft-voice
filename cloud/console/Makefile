# Makefile for SenseCraft Voice Web
# 使用方法: make [target]

# 变量定义
IMAGE_NAME := sensecraft-voice-web
TAG := latest
FULL_IMAGE := registry.example.com/sensecraft-ai/$(IMAGE_NAME):$(TAG)
PLATFORM := linux/amd64
PORT := 3000
CONTAINER_NAME := sensecraft-voice-web

# 默认目标
.DEFAULT_GOAL := help

# 显示帮助信息
.PHONY: help
help:
	@echo "🚀 SenseCraft Voice Web Docker 构建工具"
	@echo ""
	@echo "可用命令:"
	@echo "  build       构建 Docker 镜像"
	@echo "  run         运行容器"
	@echo "  stop        停止容器"
	@echo "  clean       清理容器和镜像"
	@echo "  logs        查看容器日志"
	@echo "  shell       进入容器 shell"
	@echo "  push        构建并推送镜像到仓库"
	@echo "  push-only   仅推送镜像（不构建）"
	@echo "  dev         启动开发环境"
	@echo "  install     安装依赖"
	@echo "  lint        代码检查"
	@echo "  test        运行测试"
	@echo "  up          使用 docker-compose 启动"
	@echo "  down        使用 docker-compose 停止"
	@echo "  release     完整发布流程（构建+推送+清理）"
	@echo ""
	@echo "变量:"
	@echo "  IMAGE_NAME=$(IMAGE_NAME)"
	@echo "  TAG=$(TAG)"
	@echo "  PLATFORM=$(PLATFORM)"
	@echo "  PORT=$(PORT)"

# 检查 .env.production 文件
.PHONY: check-env
check-env:
	@if [ ! -f ".env.production" ]; then \
		echo "⚠️  警告: .env.production 文件不存在"; \
		echo "📝 请创建 .env.production 文件"; \
		exit 1; \
	fi

# 构建 Docker 镜像
.PHONY: build
build: check-env
	@echo "🔨 构建 Docker 镜像..."
	@echo "📦 镜像: $(FULL_IMAGE)"
	@echo "🏗️  平台: $(PLATFORM)"
	docker build \
		--platform $(PLATFORM) \
		--tag $(FULL_IMAGE) \
		--file Dockerfile \
		.
	@echo "✅ 构建完成!"
	@docker images $(FULL_IMAGE) --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}\t{{.CreatedAt}}"

# 快速构建（跳过环境检查）
.PHONY: build-fast
build-fast:
	@echo "⚡ 快速构建 Docker 镜像..."
	docker build --platform $(PLATFORM) --tag $(FULL_IMAGE) .

# 运行容器
.PHONY: run
run:
	@echo "🏃 启动容器..."
	docker run -d \
		--name $(CONTAINER_NAME) \
		--platform $(PLATFORM) \
		-p $(PORT):80 \
		--restart unless-stopped \
		$(FULL_IMAGE)
	@echo "✅ 容器已启动!"
	@echo "🌐 访问地址: http://localhost:$(PORT)"

# 停止容器
.PHONY: stop
stop:
	@echo "🛑 停止容器..."
	-docker stop $(CONTAINER_NAME)
	-docker rm $(CONTAINER_NAME)
	@echo "✅ 容器已停止!"

# 查看日志
.PHONY: logs
logs:
	@echo "📋 查看容器日志..."
	docker logs -f $(CONTAINER_NAME)

# 进入容器 shell
.PHONY: shell
shell:
	@echo "🐚 进入容器 shell..."
	docker exec -it $(CONTAINER_NAME) /bin/sh

# 清理容器和镜像
.PHONY: clean
clean: stop
	@echo "🧹 清理镜像..."
	-docker rmi $(FULL_IMAGE)
	@echo "✅ 清理完成!"

# 推送镜像（自动构建+推送）
.PHONY: push
push: build
	@echo "📤 推送镜像到仓库..."
	docker push $(FULL_IMAGE)
	@echo "✅ 推送完成!"
	@echo "🏷️  镜像: $(FULL_IMAGE)"

# 仅推送（不构建）
.PHONY: push-only
push-only:
	@echo "📤 推送镜像到仓库..."
	docker push $(FULL_IMAGE)
	@echo "✅ 推送完成!"

# 使用 docker-compose 启动
.PHONY: up
up:
	@echo "🚀 使用 docker-compose 启动服务..."
	docker-compose up -d
	@echo "✅ 服务已启动!"

# 使用 docker-compose 停止
.PHONY: down
down:
	@echo "🛑 使用 docker-compose 停止服务..."
	docker-compose down
	@echo "✅ 服务已停止!"

# 开发环境相关命令
.PHONY: install
install:
	@echo "📦 安装依赖..."
	npm install

.PHONY: dev
dev:
	@echo "🔧 启动开发服务器..."
	npm run dev

.PHONY: build-local
build-local:
	@echo "🏗️  本地构建..."
	npm run build:prod

.PHONY: lint
lint:
	@echo "🔍 代码检查..."
	npm run lint

.PHONY: preview
preview:
	@echo "👀 预览构建结果..."
	npm run preview

# 完整的工作流
.PHONY: deploy
deploy: build run
	@echo "🚀 部署完成!"
	@echo "🌐 访问地址: http://localhost:$(PORT)"

# 重新部署
.PHONY: redeploy
redeploy: stop build run
	@echo "🔄 重新部署完成!"

# 发布流程（构建、推送、清理）
.PHONY: release
release: push clean-all
	@echo "🎉 发布完成!"
	@echo "🏷️  镜像: $(FULL_IMAGE)"
	@echo "📝 可以使用以下命令部署:"
	@echo "   docker run -d -p 3009:3000 --name $(CONTAINER_NAME) $(FULL_IMAGE)"

# 健康检查
.PHONY: health
health:
	@echo "🏥 检查容器健康状态..."
	@curl -f http://localhost:3009/ || echo "❌ 健康检查失败"

# 显示容器状态
.PHONY: status
status:
	@echo "📊 容器状态:"
	@docker ps -a --filter name=$(CONTAINER_NAME) --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

# 清理所有相关资源
.PHONY: clean-all
clean-all: down clean
	@echo "🧹 清理所有相关资源..."
	-docker volume prune -f
	-docker network prune -f
	@echo "✅ 全部清理完成!"
