#!/bin/bash

# CM4 (ARM64) 专用构建脚本
# 使用方法: ./build-cm4.sh [版本标签]

set -e

# 默认版本标签
TAG=${1:-$(git rev-parse --short HEAD)-$(date +%Y%m%d%H%M%S)}
DOCKERHUB_USER="registry.example.com/respeaker"
IMAGE_NAME="sensecraft-voice-client"

echo "🍓 开始构建CM4 (ARM64) 专用镜像..."
echo "📦 镜像名称: ${DOCKERHUB_USER}/${IMAGE_NAME}:v0.0.1"
echo "🏗️  目标架构: linux/arm64"
echo "📅 构建时间: $(date)"

# 构建ARM64镜像
echo "🏗️  开始构建ARM64镜像..."
docker build \
    --platform linux/arm64 \
    --tag ${DOCKERHUB_USER}/${IMAGE_NAME}:v0.0.1 \
    --build-arg VERSION=${TAG} \
    .

echo "✅ CM4 (ARM64) 镜像构建完成！"
echo "📋 镜像信息:"
echo "   - ARM64: ${DOCKERHUB_USER}/${IMAGE_NAME}:v0.0.1"
echo "   - 最新版ARM64: ${DOCKERHUB_USER}/${IMAGE_NAME}:v0.0.1"

# 显示镜像信息
echo "🔍 镜像详情:"
docker images | grep ${IMAGE_NAME}

echo "🎉 CM4构建完成！"
echo "💡 提示: 使用 'docker push ${DOCKERHUB_USER}/${IMAGE_NAME}:v0.0.1' 推送到仓库"
