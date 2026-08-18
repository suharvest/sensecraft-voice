# Docker 部署指南

本文档说明如何使用 Docker 构建和部署 SenseCraft Voice Client。

## 环境变量配置

### ClientVersion 环境变量

`ClientVersion` 环境变量用于指定客户端版本号，这个值会被 `RegisterDeviceJob` 使用来向远程服务注册设备。

**默认值**: `v0.0.1`

**设置方式**:
1. 在 Dockerfile 中设置
2. 通过 docker run 命令的 `-e` 参数设置
3. 通过 docker-compose 的环境变量设置
4. 通过 Makefile 的变量设置

## 构建 Docker 镜像

### 使用 Makefile（推荐）

```bash
# 使用默认版本构建
make build-docker

# 指定版本构建
make build-docker VERSION=v1.0.0 CLIENT_VERSION=v1.2.3

# 查看帮助
make help
```

### 直接使用 Docker 命令

```bash
# 基本构建
docker build -t sensecraft-voice-client:latest .

# 指定版本构建
docker build \
  --build-arg VERSION=v1.0.0 \
  --build-arg CLIENT_VERSION=v1.2.3 \
  -t sensecraft-voice-client:v1.0.0 .
```

## 运行容器

### 使用 Makefile

```bash
# 本地运行
make run

# Docker 运行
make run-docker

# 使用 Docker Compose 运行（包括模拟服务）
make up
```

### 直接使用 Docker 命令

```bash
# 基本运行
docker run -d \
  --name sensecraft-voice-client \
  -p 8090:8090 \
  -e ClientVersion=v1.2.3 \
  sensecraft-voice-client:latest

# 挂载配置文件和日志目录
docker run -d \
  --name sensecraft-voice-client \
  -p 8090:8090 \
  -e ClientVersion=v1.2.3 \
  -v $(PWD)/config.yaml:/etc/sensecraft-voice/config.yaml:ro \
  -v $(PWD)/logs:/app/logs \
  -v $(PWD)/recordings:/app/recordings \
  sensecraft-voice-client:latest
```

## Docker Compose 部署

### 启动所有服务

```bash
# 启动所有服务（包括模拟设备注册服务）
make up

# 或者直接使用 docker-compose
docker-compose up -d
```

### 查看日志

```bash
# 查看应用日志
make logs

# 或者直接使用 docker-compose
docker-compose logs -f sensecraft-voice-client
```

### 停止服务

```bash
# 停止所有服务
make down

# 或者直接使用 docker-compose
docker-compose down
```

## 环境变量配置示例

### 开发环境

```bash
export CLIENT_VERSION=v0.1.0-dev
make build-docker
make run-docker
```

### 生产环境

```bash
export CLIENT_VERSION=v1.2.3
export VERSION=v1.2.3
make build-docker
make run-docker
```

### Kubernetes 部署

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sensecraft-voice-client
spec:
  replicas: 1
  selector:
    matchLabels:
      app: sensecraft-voice-client
  template:
    metadata:
      labels:
        app: sensecraft-voice-client
    spec:
      containers:
      - name: sensecraft-voice-client
        image: sensecraft-voice-client:v1.2.3
        env:
        - name: ClientVersion
          value: "v1.2.3"
        ports:
        - containerPort: 8090
        volumeMounts:
        - name: config
          mountPath: /etc/sensecraftVoice
        - name: logs
          mountPath: /app/logs
        - name: recordings
          mountPath: /app/recordings
      volumes:
      - name: config
        configMap:
          name: sensecraft-config
      - name: logs
        emptyDir: {}
      - name: recordings
        emptyDir: {}
```

## 配置文件挂载

### 配置文件

将 `config.yaml` 挂载到容器的 `/etc/sensecraft-voice/config.yaml`：

```bash
-v $(PWD)/config.yaml:/etc/sensecraft-voice/config.yaml:ro
```

### 日志目录

将日志目录挂载到容器的 `/app/logs`：

```bash
-v $(PWD)/logs:/app/logs
```

### 录音文件目录

将录音文件目录挂载到容器的 `/app/recordings`：

```bash
-v $(PWD)/recordings:/app/recordings
```

## 版本管理

### 版本标签策略

- `latest`: 最新开发版本
- `v1.2.3`: 语义化版本标签
- `v1.2.3-20240829`: 带日期的版本标签

### 构建不同版本

```bash
# 开发版本
make build-docker VERSION=dev CLIENT_VERSION=v0.1.0-dev

# 测试版本
make build-docker VERSION=test CLIENT_VERSION=v1.2.3-rc1

# 生产版本
make build-docker VERSION=v1.2.3 CLIENT_VERSION=v1.2.3
```

## 故障排除

### 常见问题

1. **环境变量未生效**
   - 检查 Dockerfile 中的 ENV 设置
   - 确认 docker run 命令中的 `-e` 参数
   - 验证容器内的环境变量：`docker exec -it container_name env`

2. **配置文件挂载失败**
   - 检查文件路径是否正确
   - 确认文件权限
   - 使用 `:ro` 标志进行只读挂载

3. **端口绑定失败**
   - 检查端口是否被占用
   - 确认防火墙设置
   - 验证端口映射：`docker port container_name`

### 调试命令

```bash
# 查看容器状态
docker ps -a

# 查看容器日志
docker logs sensecraft-voice-client

# 进入容器调试
docker exec -it sensecraft-voice-client sh

# 查看容器环境变量
docker exec sensecraft-voice-client env

# 查看容器挂载点
docker inspect sensecraft-voice-client | grep -A 10 "Mounts"
```

## 最佳实践

1. **版本管理**
   - 使用语义化版本号
   - 为每个版本打标签
   - 在生产环境使用固定版本号

2. **配置管理**
   - 将配置文件挂载到容器
   - 使用环境变量覆盖敏感配置
   - 避免在镜像中硬编码配置

3. **日志管理**
   - 将日志目录挂载到宿主机
   - 配置日志轮转
   - 监控日志文件大小

4. **安全考虑**
   - 使用非 root 用户运行容器
   - 限制容器权限
   - 定期更新基础镜像

## 相关文件

- `Dockerfile`: Docker 镜像构建文件
- `docker-compose.yml`: Docker Compose 配置文件
- `Makefile`: 构建和部署脚本
- `.dockerignore`: Docker 构建忽略文件
- `config.yaml`: 应用配置文件
