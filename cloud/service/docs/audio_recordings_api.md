# 音频录音API文档

## 概述

音频录音API提供了音频文件的上传、存储、查询和下载功能。支持按照 `{session_id}-{audio_id}.wav` 的格式存储音频文件。

## 文件命名规则

音频文件必须按照以下格式命名：
```
{session_id}-{audio_id}.wav
```

**示例**:
- `ef7aea10397ecc52dc4fd88a9470c752-vv0hxr.wav`
- `abc123def456-xyz789.wav`

## 文件结构

```
recordings/data/
├── session1-audio1.wav
├── session1-audio2.wav
├── session2-audio1.wav
└── subdirectory/
    └── session3-audio1.wav
```

## API接口

### 1. 上传音频录音

**接口**: `POST /api/v1/recordings/upload`

**请求参数**:
- `file` (multipart/form-data): 音频文件，必须是WAV格式
- `session_id` (form): 会话ID，必填
- `audio_id` (form): 音频ID，必填
- `mac_address` (form): MAC地址，必填，格式为 `AA:BB:CC:DD:EE:FF`

**响应格式**:
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

### 2. 获取音频录音信息

**接口**: `GET /api/v1/recordings/{id}`

**路径参数**:
- `id`: 音频录音ID

**响应格式**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "uploaded_file_id",
    "session_id": "ef7aea10397ecc52dc4fd88a9470c752",
    "audio_id": "vv0hxr",
    "mac_address": "AA:BB:CC:DD:EE:FF",
    "file_path": "recordings/data/ef7aea10397ecc52dc4fd88a9470c752-vv0hxr.wav",
    "file_size": 1024000,
    "upload_time": 1705312245000,
    "status": 1,
    "created_at": 1705312245000,
    "updated_at": 1705312245000
  }
}
```

### 3. 根据session_id和audio_id获取音频录音信息

**接口**: `GET /api/v1/recordings/session/{session_id}/audio/{audio_id}`

**路径参数**:
- `session_id`: 会话ID
- `audio_id`: 音频ID

**响应格式**: 同接口2

### 4. 获取音频录音列表

**接口**: `GET /api/v1/recordings/`

**查询参数**:
- `session_id` (可选): 会话ID
- `mac_address` (可选): MAC地址
- `start_time` (可选): 开始时间戳
- `end_time` (可选): 结束时间戳
- `status` (可选): 状态，1=正常，0=已删除
- `offset` (可选): 偏移量，默认0
- `limit` (可选): 限制数量，默认50，最大1000

**响应格式**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total": 100,
    "items": [
      {
        "id": "uploaded_file_id",
        "session_id": "ef7aea10397ecc52dc4fd88a9470c752",
        "audio_id": "vv0hxr",
        "mac_address": "AA:BB:CC:DD:EE:FF",
        "file_path": "recordings/data/ef7aea10397ecc52dc4fd88a9470c752-vv0hxr.wav",
        "file_size": 1024000,
        "upload_time": 1705312245000,
        "status": 1,
        "created_at": 1705312245000,
        "updated_at": 1705312245000
      }
    ]
  }
}
```

### 5. 下载音频录音

**接口**: `GET /api/v1/recordings/{id}/download`

**路径参数**:
- `id`: 音频录音ID

**响应**: 直接返回音频文件流，Content-Type为 `audio/wav`

### 6. 删除音频录音

**接口**: `DELETE /api/v1/recordings/{id}`

**路径参数**:
- `id`: 音频录音ID

**响应格式**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "audio recording deleted successfully"
  }
}
```

## 错误码

**常见错误码**:
- `400 Bad Request`: 请求参数错误
- `404 Not Found`: 资源不存在
- `409 Conflict`: 资源已存在（相同的session_id+audio_id组合）
- `413 Payload Too Large`: 文件过大
- `415 Unsupported Media Type`: 不支持的文件格式
- `500 Internal Server Error`: 服务器内部错误
- `503 Service Unavailable`: 服务不可用

## 数据库设计

```sql
CREATE TABLE audio_recordings (
    id VARCHAR(64) PRIMARY KEY,
    session_id VARCHAR(64) NOT NULL,
    audio_id VARCHAR(64) NOT NULL,
    mac_address VARCHAR(17) NOT NULL,
    file_path VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL,
    upload_time BIGINT NOT NULL,
    status TINYINT DEFAULT 1,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    UNIQUE KEY idx_session_audio (session_id, audio_id),
    INDEX idx_session_id (session_id),
    INDEX idx_mac_address (mac_address),
    INDEX idx_upload_time (upload_time)
);
```

## 使用示例

### 上传音频文件
```bash
curl -X POST "http://localhost:8080/api/v1/recordings/upload" \
  -F "file=@audio.wav" \
  -F "session_id=ef7aea10397ecc52dc4fd88a9470c752" \
  -F "audio_id=vv0hxr" \
  -F "mac_address=AA:BB:CC:DD:EE:FF"
```

### 获取音频录音信息
```bash
curl "http://localhost:8080/api/v1/recordings/uploaded_file_id"
```

### 根据session_id和audio_id获取
```bash
curl "http://localhost:8080/api/v1/recordings/session/ef7aea10397ecc52dc4fd88a9470c752/audio/vv0hxr"
```

### 下载音频文件
```bash
curl -O "http://localhost:8080/api/v1/recordings/uploaded_file_id/download"
```

### 获取列表
```bash
curl "http://localhost:8080/api/v1/recordings/?session_id=ef7aea10397ecc52dc4fd88a9470c752&limit=10"
```

### 删除音频录音
```bash
curl -X DELETE "http://localhost:8080/api/v1/recordings/uploaded_file_id"
```
