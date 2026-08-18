#!/bin/bash

# 多架构构建脚本 - 支持CM4 (ARM64) 和 x86_64 (AMD64)
# 使用方法: ./build-multiarch.sh [版本标签]

set -e

# 默认版本标签
TAG=${1:-$(git rev-parse --short HEAD)-$(date +%Y%m%d%H%M%S)}
DOCKERHUB_USER="registry.example.com/respeaker"
IMAGE_NAME="sensecraft-voice-client"

echo "🚀 开始构建多架构镜像..."
echo "📦 镜像名称: ${DOCKERHUB_USER}/${IMAGE_NAME}:${TAG}"
echo "🏗️  支持架构: linux/amd64, linux/arm64"
echo "📅 构建时间: $(date)"

# 检查是否安装了docker buildx
if ! docker buildx version > /dev/null 2>&1; then
    echo "❌ 错误: 请先安装 docker buildx"
    echo "💡 提示: 运行 'docker buildx install' 来安装"
    exit 1
fi

# 创建新的构建器实例（如果不存在）
BUILDER_NAME="multiarch-builder"
if ! docker buildx inspect $BUILDER_NAME > /dev/null 2>&1; then
    echo "🔧 创建新的构建器实例: $BUILDER_NAME"
    docker buildx create --name $BUILDER_NAME --use
fi

# 使用指定的构建器
docker buildx use $BUILDER_NAME

# 构建多架构镜像
echo "🏗️  开始构建..."
docker buildx build \
    --platform linux/amd64,linux/arm64 \
    --tag ${DOCKERHUB_USER}/${IMAGE_NAME}:${TAG} \
    --tag ${DOCKERHUB_USER}/${IMAGE_NAME}:latest \
    --build-arg VERSION=${TAG} \
    --push \
    .

echo "✅ 多架构镜像构建完成！"
echo "📋 镜像信息:"
echo "   - AMD64: ${DOCKERHUB_USER}/${IMAGE_NAME}:${TAG}"
echo "   - ARM64: ${DOCKERHUB_USER}/${IMAGE_NAME}:${TAG}"
echo "   - 最新版: ${DOCKERHUB_USER}/${IMAGE_NAME}:latest"

# 显示镜像清单
echo "🔍 镜像清单:"
docker buildx imagetools inspect ${DOCKERHUB_USER}/${IMAGE_NAME}:${TAG}

echo "🎉 构建完成！现在可以在CM4设备上运行了。"
