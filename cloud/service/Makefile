.PHONY: run build build-linux image push clean help

# 构建参数
tag = latest
releaseName = sensecraft-voice-service
dockerhubUser ?= registry.example.com/sensecraft-ai
targetDir ?= dist
commitHash = $(shell git rev-parse --short HEAD)
version = $(commitHash)-$(shell date +%Y%m%d%H%M%S)

# 平台配置
PLATFORMS = linux/amd64
DOCKER_BUILDX_NAME = sensecraft-builder

# 默认目标
.DEFAULT_GOAL := help

ALL: run

# 本地运行
run: build
	./$(targetDir)/$(releaseName) --configfile ./config.yaml

# 本地构建
build:
	@echo "Building for local platform..."
	@mkdir -p $(targetDir)
	go build -o $(targetDir)/$(releaseName) -ldflags "-X 'main.version=$(version)'" ./cmd/server.go

# Linux 平台构建
build-linux:
	@echo "Building for Linux AMD64..."
	@mkdir -p $(targetDir)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build \
		-o $(targetDir)/$(releaseName)-linux-amd64 \
		-ldflags "-s -w -X 'main.version=$(version)'" \
		./cmd/server.go

# 创建 Docker 构建器（如果不存在）
create-builder:
	@echo "Creating Docker buildx builder..."
	@docker buildx create --name $(DOCKER_BUILDX_NAME) --use --driver docker-container 2>/dev/null || \
	docker buildx use $(DOCKER_BUILDX_NAME) || true

# 构建 Docker 镜像
image: create-builder
	@echo "Building Docker image: $(dockerhubUser)/$(releaseName):$(tag)"
	@echo "Version: $(version)"
	docker buildx build \
		--platform $(PLATFORMS) \
		--tag $(dockerhubUser)/$(releaseName):$(tag) \
		--build-arg VERSION=$(version) \
		--build-arg TARGETOS=linux \
		--build-arg TARGETARCH=amd64 \
		--load \
		.

# 构建并推送镜像
push: create-builder
	@echo "Building and pushing Docker image: $(dockerhubUser)/$(releaseName):$(tag)"
	@echo "Version: $(version)"
	docker buildx build \
		--platform $(PLATFORMS) \
		--tag $(dockerhubUser)/$(releaseName):$(tag) \
		--tag $(dockerhubUser)/$(releaseName):latest \
		--build-arg VERSION=$(version) \
		--build-arg TARGETOS=linux \
		--build-arg TARGETARCH=amd64 \
		--push \
		.

# 推送现有镜像
push-only:
	@echo "Pushing Docker image: $(dockerhubUser)/$(releaseName):$(tag)"
	docker push $(dockerhubUser)/$(releaseName):$(tag)
	docker push $(dockerhubUser)/$(releaseName):latest

# 清理构建产物
clean:
	@echo "Cleaning build artifacts..."
	-rm -rf $(targetDir)
	-rm -f ./$(releaseName)
	-rm -f ./$(releaseName)-linux-amd64

# 清理 Docker 资源
clean-docker:
	@echo "Cleaning Docker resources..."
	-docker rmi $(dockerhubUser)/$(releaseName):$(tag) 2>/dev/null || true
	-docker rmi $(dockerhubUser)/$(releaseName):latest 2>/dev/null || true
	-docker buildx rm $(DOCKER_BUILDX_NAME) 2>/dev/null || true

# 显示帮助信息
help:
	@echo "Available targets:"
	@echo "  run          - Build and run locally"
	@echo "  build        - Build for local platform"
	@echo "  build-linux  - Build for Linux AMD64"
	@echo "  image        - Build Docker image"
	@echo "  push         - Build and push Docker image"
	@echo "  push-only    - Push existing Docker image"
	@echo "  clean        - Clean build artifacts"
	@echo "  clean-docker - Clean Docker resources"
	@echo "  help         - Show this help message"
	@echo ""
	@echo "Variables:"
	@echo "  tag=$(tag)"
	@echo "  version=$(version)"
	@echo "  dockerhubUser=$(dockerhubUser)"

# 生成 API 文档
.PHONY: api-docs
api-docs:
	@echo "Generating API documentation..."
	swag init --generalInfo ./cmd/appserver.go --output ./api/docs

# 测试构建
test-build: build-linux
	@echo "Testing Linux build..."
	@file $(targetDir)/$(releaseName)-linux-amd64
	@echo "Build successful for Linux AMD64"
