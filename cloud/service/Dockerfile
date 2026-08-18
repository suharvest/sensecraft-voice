# 多阶段构建，支持 x86_64 Linux
FROM --platform=linux/amd64 golang:1.23-alpine AS builder

# 设置工作目录
WORKDIR /app

# 构建参数
ARG VERSION
ARG TARGETOS=linux
ARG TARGETARCH=amd64

# 设置环境变量
ENV GOPROXY=https://goproxy.cn
ENV CGO_ENABLED=0
ENV GOOS=${TARGETOS}
ENV GOARCH=${TARGETARCH}

# 安装必要的构建工具
RUN apk add --no-cache git ca-certificates tzdata

# 复制依赖文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN go build \
    -ldflags "-s -w -X 'main.version=${VERSION}'" \
    -o server \
    ./cmd/server.go

# 运行阶段
FROM --platform=linux/amd64 alpine:latest AS runner

# 安装必要的运行时工具
RUN apk add --no-cache ca-certificates tzdata

# 创建非 root 用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder --chown=appuser:appgroup /app/server .
COPY --from=builder --chown=appuser:appgroup /app/config.yaml /etc/sensecraft-voice/config.yaml

# 切换到非 root 用户
USER appuser

# 暴露端口（如果需要）
EXPOSE 8081

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8081/healthz || exit 1

# 启动命令
ENTRYPOINT ["./server"]