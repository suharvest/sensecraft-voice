FROM --platform=linux/arm64 golang:1.22-alpine as builder
WORKDIR /app
ARG VERSION
ARG CLIENT_VERSION
ENV GOPROXY=https://goproxy.cn
ENV ClientVersion=${CLIENT_VERSION:-v0.0.9}

# 安装必要的构建依赖
RUN apk add --no-cache gcc musl-dev

COPY ./go.mod ./
COPY ./go.sum ./
COPY . .

# 构建ARM64可执行文件
RUN CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -ldflags "-s -w -X 'main.version=${VERSION}'" -o /app/sensecraftVoice ./cmd/
# 验证可执行文件是否存在
RUN ls -la /app/sensecraftVoice

FROM --platform=linux/arm64 alpine:latest as runner
# 设置客户端版本环境变量
ENV ClientVersion=v0.0.9
ARG CLIENT_VERSION
ENV ClientVersion=${CLIENT_VERSION:-v0.0.9}

# 安装音频工具包
RUN apk add --no-cache \
    alsa-utils \
    alsa-lib \
    pulseaudio-utils \
    curl \
    && rm -rf /var/cache/apk/*

# 音频设备环境变量将在运行时通过docker run或docker-compose设置
# 移除硬编码的设备ID，让应用使用系统默认设备或运行时指定的设备

# 复制静态文件到正确位置
COPY --from=builder /app/web /app/web
COPY --from=builder /app/config.yaml /etc/sensecraft-voice/config.yaml
COPY --from=builder /app/sensecraftVoice /app/sensecraftVoice
# 添加执行权限
RUN chmod +x /app/sensecraftVoice
# 验证文件
RUN ls -la /app/web

WORKDIR /app
CMD ["./sensecraftVoice"]
