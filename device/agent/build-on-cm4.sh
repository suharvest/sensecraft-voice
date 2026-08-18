#!/bin/bash

# 在CM4设备上本地构建脚本
# 使用方法: ./build-on-cm4.sh [版本标签]

set -e

# 默认版本标签
TAG=${1:-$(git rev-parse --short HEAD)-$(date +%Y%m%d%H%M%S)}
DOCKERHUB_USER="registry.example.com/respeaker"
IMAGE_NAME="sensecraft-voice-client"

echo "🍓 在CM4设备上本地构建镜像..."
echo "📦 镜像名称: ${DOCKERHUB_USER}/${IMAGE_NAME}:${TAG}-cm4-local"
echo "🏗️  目标架构: linux/arm64 (本地构建)"
echo "📅 构建时间: $(date)"

# 检查是否在ARM64架构上
if [ "$(uname -m)" != "aarch64" ]; then
    echo "❌ 警告: 当前不在ARM64架构上，构建的镜像可能无法在CM4上运行"
    echo "💡 建议在CM4设备上运行此脚本"
fi

# 构建ARM64镜像（在本地架构上）
echo "🏗️  开始构建ARM64镜像..."
docker build \
    --platform linux/arm64 \
    --tag ${DOCKERHUB_USER}/${IMAGE_NAME}:${TAG}-cm4-local \
    --tag ${DOCKERHUB_USER}/${IMAGE_NAME}:latest-cm4-local \
    --build-arg VERSION=${TAG} \
    .

echo "✅ CM4本地构建完成！"
echo "📋 镜像信息:"
echo "   - ARM64: ${DOCKERHUB_USER}/${IMAGE_NAME}:${TAG}-cm4-local"
echo "   - 最新版ARM64: ${DOCKERHUB_USER}/${IMAGE_NAME}:latest-cm4-local"

# 显示镜像信息
echo "🔍 镜像详情:"
docker images | grep ${IMAGE_NAME}

echo "🎉 CM4本地构建完成！"
echo "💡 提示: 使用 'docker push ${DOCKERHUB_USER}/${IMAGE_NAME}:${TAG}-cm4-local' 推送到仓库"
echo "🚀 现在可以在CM4上运行: docker run -d --name voice-client -p 8090:8090 ${DOCKERHUB_USER}/${IMAGE_NAME}:${TAG}-cm4-local"
