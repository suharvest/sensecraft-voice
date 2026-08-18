# 音频文件上传API文档

## 概述

音频文件上传功能通过JobManager每5秒自动扫描指定目录下的音频文件，并上传到远程服务器。上传成功后立即删除本地文件，失败则保留文件等待下次重试。

## 功能特性

- **自动扫描**: 每5秒扫描一次指定目录
- **文件格式**: 支持 `sessionid-audioid.wav` 格式的文件
- **成功即删**: 上传成功后立即删除本地文件
- **失败保留**: 上传失败保留文件，自动重试
- **并发控制**: 支持配置最大并发上传数
- **文件大小限制**: 支持配置最大文件大小限制
- **超时控制**: 支持配置上传超时时间

## 配置说明

### 配置文件 (config.yaml)

```yaml
# 音频文件上传配置
audio_upload:
  # 是否启用音频文件上传
  enabled: true
  # 扫描目录
  scan_dir: "./recordings/data"
  # MAC地址（留空则自动获取）
  mac_address: ""
  # 上传超时
  timeout: "30s"
  # 最大文件大小 (MB)
  max_file_size: 50
  # 最大并发上传数
  max_concurrent: 3
```

### 配置参数说明

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| enabled | bool | 是 | false | 是否启用音频文件上传功能 |
| scan_dir | string | 是 | - | 扫描音频文件的目录路径 |
| mac_address | string | 否 | - | 设备MAC地址（留空则自动获取） |
| timeout | string | 否 | "30s" | HTTP请求超时时间 |
| max_file_size | int64 | 否 | 50 | 最大文件大小限制(MB) |
| max_concurrent | int | 否 | 3 | 最大并发上传数 |

**注意**: 上传URL通过 `remote.base_url` 统一管理，无需单独配置 `upload_url`。实际的上传地址为 `{remote.base_url}/api/v1/recordings/upload`。

## 文件格式要求

### 文件命名规则

音频文件支持以下两种格式：

**格式1**: 文件名格式
```
{session_id}-{audio_id}.wav
```

**格式2**: 目录结构格式
```
{session_id}/{audio_id}.wav
```

**示例**:
- `ef7aea10397ecc52dc4fd88a9470c752-vv0hxr.wav`
- `abc123def456-xyz789.wav`
- `ef7aea10397ecc52dc4fd88a9470c752/zzlk42.wav`
- `abc123def456/xyz789.wav`

### 文件结构

```
recordings/data/
├── session1-audio1.wav          # 格式1: 文件名格式
├── session1-audio2.wav
├── session2-audio1.wav
├── session3/                    # 格式2: 目录结构格式
│   ├── audio1.wav
│   └── audio2.wav
└── session4/
    └── audio1.wav
```

#### 响应格式

**成功响应** (HTTP 200/201):
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

**失败响应** (HTTP 4xx/5xx):
```json
{
  "code": 400,
  "message": "Invalid file format",
  "data": null
}
```

**常见错误码**:
- `400 Bad Request`: 请求参数错误
- `413 Payload Too Large`: 文件过大
- `415 Unsupported Media Type`: 不支持的文件格式
- `500 Internal Server Error`: 服务器内部错误
- `503 Service Unavailable`: 服务不可用

### 云端API实现建议

#### 1. 接口验证
- 验证文件格式是否为WAV
- 验证文件大小是否在限制范围内
- 验证session_id和audio_id格式
- 验证MAC地址格式

#### 2. 文件处理
- 支持大文件上传（建议支持断点续传）
- 文件去重处理
- 文件存储到云存储服务
- 生成唯一文件ID

#### 3. 数据库设计
```sql
CREATE TABLE audio_recordings (
    id VARCHAR(64) PRIMARY KEY,
    session_id VARCHAR(64) NOT NULL,
    audio_id VARCHAR(64) NOT NULL,
    mac_address VARCHAR(17) NOT NULL,
    file_path VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL,
    upload_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status TINYINT DEFAULT 1,
    INDEX idx_session_id (session_id),
    INDEX idx_mac_address (mac_address),
    INDEX idx_upload_time (upload_time)
);
```

#### 4. MAC地址自动获取
- 当配置中 `mac_address` 为空时，系统会自动获取设备MAC地址
- 自动获取逻辑会跳过回环接口、虚拟接口（docker、veth、br-等）
- 优先选择第一个有效的物理网络接口的MAC地址
- 如果无法获取MAC地址，将使用 "unknown" 作为默认值

#### 5. 安全考虑
- 实现请求频率限制
- 添加API密钥验证
- 文件类型白名单验证
- 防止恶意文件上传

#### 6. 性能优化
- 使用CDN加速文件上传
- 实现文件分片上传
- 添加负载均衡
- 数据库读写分离

#### 7. 监控告警
- 上传成功率监控
- 响应时间监控
- 错误率告警
- 存储空间监控

## 云端API接口设计

### 音频文件上传接口

#### 接口地址
```
POST {remote.base_url}/api/v1/recordings/upload
```

#### 请求头
```
Content-Type: multipart/form-data
User-Agent: SenseCraft-Voice-Client/1.0
```

#### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| file | file | 是 | 音频文件(.wav格式) |
| session_id | string | 是 | 会话ID，从文件名解析 |
| audio_id | string | 是 | 音频ID，从文件名解析 |
| mac_address | string | 是 | 设备MAC地址（自动获取），格式为AA:BB:CC:DD:EE:FF |

#### 请求示例

**cURL示例**:
```bash
curl -X POST "http://voice-service.example.com:3008/api/v1/recordings/upload" \
  -F "file=@ef7aea10397ecc52dc4fd88a9470c752-vv0hxr.wav" \
  -F "session_id=ef7aea10397ecc52dc4fd88a9470c752" \
  -F "audio_id=vv0hxr" \
  -F "mac_address=6e:8e:84:f9:73:d6"
```

#### 响应格式

**成功响应** (HTTP 200/201):
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

**失败响应** (HTTP 4xx/5xx):
```json
{
  "code": 400,
  "message": "Invalid file format",
  "data": null
}
```

**常见错误码**:
- `400 Bad Request`: 请求参数错误
- `413 Payload Too Large`: 文件过大
- `415 Unsupported Media Type`: 不支持的文件格式
- `500 Internal Server Error`: 服务器内部错误
- `503 Service Unavailable`: 服务不可用

## 日志监控

### 日志级别

- **INFO**: 正常扫描和上传操作
- **WARN**: 文件格式错误、上传失败等警告
- **ERROR**: 系统错误、配置错误等

### 关键日志示例

**文件扫描日志**:
```
INFO: 找到 3 个音频文件待上传
INFO: 文件 ef7aea10397ecc52dc4fd88a9470c752-vv0hxr.wav 上传成功并已删除
WARN: 文件 invalid-name.wav 文件名格式错误，跳过处理
```

**上传结果日志**:
```
INFO: 音频上传任务完成: 成功 4, 失败 1
WARN: 文件 large-file.wav 超过大小限制(50MB)，跳过上传
ERROR: 上传文件 session1-audio1.wav 失败: 连接超时
```

## 错误处理

### 常见错误及处理

| 错误类型 | 错误信息 | 处理方式 |
|----------|----------|----------|
| 文件格式错误 | "invalid filename format" | 跳过处理，记录警告日志 |
| 文件过大 | "file too large" | 跳过处理，记录警告日志 |
| 网络错误 | "failed to send HTTP request" | 保留文件，等待下次重试 |
| 服务器错误 | "HTTP请求失败，状态码: 500" | 保留文件，等待下次重试 |
| 文件不存在 | "failed to open file" | 跳过处理，记录错误日志 |

### 重试机制

- 上传失败的文件会在下次扫描时自动重试
- 无重试次数限制，直到上传成功或文件被手动删除
- 每次扫描都会尝试上传所有失败的文件

## 性能优化

### 并发控制

- 默认最大并发数为3，可根据服务器性能调整
- 并发上传不会影响文件扫描的准确性
- 建议根据网络带宽和服务器性能调整并发数

### 文件大小限制

- 默认最大文件大小为50MB
- 超过限制的文件会被跳过，不会尝试上传
- 可根据实际需求调整文件大小限制

### 扫描频率

- 默认每5秒扫描一次
- 扫描频率过高可能影响系统性能
- 扫描频率过低可能导致文件积压

## 部署说明

### 目录权限

确保应用对以下目录有读写权限：
- 扫描目录: `./recordings/data`
- 日志目录: `./logs`

### 网络要求

- 确保能够访问配置的上传URL
- 建议配置防火墙规则允许出站HTTP/HTTPS请求
- 如使用代理，需要配置相应的代理设置

### 监控建议

- 监控上传成功率
- 监控失败文件数量
- 监控磁盘空间使用情况
- 监控网络连接状态

## 故障排查

### 常见问题

1. **文件不上传**
   - 检查配置是否正确
   - 检查文件命名格式
   - 检查网络连接
   - 查看日志错误信息

2. **上传失败率高**
   - 检查服务器状态
   - 检查网络稳定性
   - 调整超时时间
   - 检查文件大小限制

3. **文件积压**
   - 检查上传URL是否可访问
   - 检查服务器响应
   - 考虑降低扫描频率
   - 手动清理无效文件

### 调试模式

启用调试模式可以获取更详细的日志信息：

```yaml
default:
  mode: debug
  log_format: text
```

## 更新日志

### v1.0.0 (2024-01-XX)
- 初始版本发布
- 支持基本的音频文件上传功能
- 支持文件格式验证和错误处理
- 支持配置化的扫描和上传参数
