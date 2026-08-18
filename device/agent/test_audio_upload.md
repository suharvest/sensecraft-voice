# 音频上传功能测试指南

## 测试准备

### 1. 配置文件设置
确保 `config.yaml` 中的配置正确：

```yaml
# 远程服务配置
remote:
  base_url: "http://voice-service.example.com:3008"

# 音频文件上传配置
audio_upload:
  enabled: true
  scan_dir: "./recordings/data"
  mac_address: ""  # 留空则自动获取
  timeout: "30s"
  max_file_size: 50
  max_concurrent: 3
```

### 2. 创建测试音频文件
在 `./recordings/data/` 目录下创建测试文件：

```bash
# 创建测试目录
mkdir -p ./recordings/data

# 格式1: 文件名格式 sessionid-audioid.wav
touch ./recordings/data/ef7aea10397ecc52dc4fd88a9470c752-vv0hxr.wav
touch ./recordings/data/abc123def456-xyz789.wav

# 格式2: 目录结构格式 sessionid/audioid.wav
mkdir -p ./recordings/data/ef7aea10397ecc52dc4fd88a9470c752
mkdir -p ./recordings/data/abc123def456
touch ./recordings/data/ef7aea10397ecc52dc4fd88a9470c752/zzlk42.wav
touch ./recordings/data/abc123def456/xyz789.wav
```

## 测试步骤

### 1. 启动服务
```bash
go run cmd/client.go
```

### 2. 观察日志
查看控制台输出，应该看到类似以下日志：

```
INFO: 找到 4 个音频文件待上传
INFO: 文件 ef7aea10397ecc52dc4fd88a9470c752-vv0hxr.wav 上传成功并已删除
INFO: 文件 abc123def456-xyz789.wav 上传成功并已删除
INFO: 文件 ef7aea10397ecc52dc4fd88a9470c752/zzlk42.wav 上传成功并已删除
INFO: 文件 abc123def456/xyz789.wav 上传成功并已删除
```

### 3. 验证文件删除
检查 `./recordings/data/` 目录，成功上传的文件应该被自动删除。

### 4. 测试失败场景
创建一个格式错误的文件：

```bash
# 创建格式错误的文件
touch ./recordings/data/invalid-name.wav
```

观察日志，应该看到警告信息：

```
WARN: 文件 invalid-name.wav 文件名格式错误，跳过处理
```

## API接口测试

### 使用cURL测试上传接口

```bash
curl -X POST "http://voice-service.example.com:3008/api/v1/recordings/upload" \
  -F "file=@ef7aea10397ecc52dc4fd88a9470c752-vv0hxr.wav" \
  -F "session_id=ef7aea10397ecc52dc4fd88a9470c752" \
  -F "audio_id=vv0hxr" \
  -F "mac_address=6e:8e:84:f9:73:d6"
```

### 预期响应

**成功响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "uploaded_file_id",
    "session_id": "ef7aea10397ecc52dc4fd88a9470c752",
    "audio_id": "vv0hxr",
    "file_size": 1024000,
    "upload_time": 1705312245000
  }
}
```

## 功能验证清单

- [ ] 服务启动正常
- [ ] 自动扫描音频文件
- [ ] 正确解析文件名格式
- [ ] 成功上传文件到云端API
- [ ] 上传成功后删除本地文件
- [ ] 上传失败时保留本地文件
- [ ] 自动获取MAC地址
- [ ] 错误文件被跳过处理
- [ ] 日志记录完整

## 故障排查

### 常见问题

1. **文件不上传**
   - 检查 `remote.base_url` 配置
   - 检查网络连接
   - 查看错误日志

2. **文件名格式错误**
   - 确保文件名为 `sessionid-audioid.wav` 格式
   - 检查文件名中是否包含特殊字符

3. **MAC地址获取失败**
   - 检查网络接口配置
   - 查看日志中的MAC地址信息

4. **上传失败**
   - 检查云端API服务状态
   - 验证API接口参数格式
   - 查看HTTP响应状态码
