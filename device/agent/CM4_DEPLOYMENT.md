# 🍓 CM4 部署指南

## 问题分析

你遇到的 `exec format error` 错误是因为：
1. 镜像在x86_64主机上构建
2. 但要在ARM64的树莓派CM4上运行
3. 架构不匹配导致无法执行

## 解决方案

### 方案1: 在CM4设备上本地构建（推荐）

```bash
# 1. 在CM4设备上克隆代码
git clone <your-repo>
cd sensecraft-voice-client

# 2. 给构建脚本添加执行权限
chmod +x build-on-cm4.sh

# 3. 在CM4上本地构建
./build-on-cm4.sh

# 4. 运行容器
docker run -d \
  --name sensecraft-respeaker \
  -p 8090:8090 \
  -v $(pwd)/config.yaml:/etc/sensecraft-voice/config.yaml \
  registry.example.com/respeaker/sensecraft-voice-client:latest-cm4-local
```

### 方案2: 使用docker-compose

```bash
# 使用专门的CM4配置
docker-compose -f docker-compose.cm4.yml up -d
```

### 方案3: 交叉编译构建

如果你必须在x86_64主机上构建ARM64镜像：

```bash
# 1. 启用docker buildx
docker buildx create --name multiarch --use

# 2. 构建多架构镜像
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --tag your-image:latest \
  --push .
```

## 验证部署

```bash
# 检查容器状态
docker ps -a

# 查看日志
docker logs -f sensecraft-respeaker

# 测试API
curl http://localhost:8090/v1/voice/device/status
```

## 常见问题

### Q: 为什么会出现exec format error？
A: 这是因为镜像架构与运行环境架构不匹配。

### Q: 如何在CM4上构建？
A: 直接在CM4设备上运行 `./build-on-cm4.sh` 脚本。

### Q: 可以推送到Docker Hub吗？
A: 可以，构建完成后使用 `docker push` 命令。

## 性能优化

CM4设备上的优化建议：
- 限制内存使用：`--memory=2g`
- 使用SSD存储提高I/O性能
- 定期清理日志和录音文件
- 监控系统资源使用情况
