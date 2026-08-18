# 音频上传API文档

## 概述

本文档描述了音频分块上传相关的API接口，支持设备端每30秒上报一个音频块，并携带开始时间和结束时间信息。

## 基础信息

- **基础URL**: `/api/v1/audio`
- **认证方式**: 待定
- **数据格式**: JSON
- **字符编码**: UTF-8

## 接口列表

### 1. 上传音频块

**接口描述**: 上传音频数据块到MinIO存储

**请求信息**:
- **URL**: `POST /api/v1/audio/upload-chunk`
- **Content-Type**: `application/json`

**请求参数**:
```json
{
  "device_id": "AA:BB:CC:DD:EE:FF",
  "session_id": "uuid-optional",
  "chunk_info": {
    "chunk_index": 1,
    "start_time": 1640995200000,
    "end_time": 1640995230000,
    "duration": 30000,
    "format": "wav",
    "sample_rate": 16000,
    "channels": 1,
    "bit_rate": 128000
  },
  "audio_data": "base64_encoded_audio",
  "checksum": "md5_hash"
}
```

**参数说明**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| device_id | string | 是 | 设备MAC地址 |
| session_id | string | 否 | 会话ID，为空时自动生成 |
| chunk_info | object | 是 | 音频块信息 |
| chunk_info.chunk_index | int | 是 | 块索引，从1开始 |
| chunk_info.start_time | int64 | 是 | 音频开始时间(毫秒) |
| chunk_info.end_time | int64 | 是 | 音频结束时间(毫秒) |
| chunk_info.duration | int | 是 | 音频时长(毫秒) |
| chunk_info.format | string | 是 | 音频格式(wav/mp3/pcm) |
| chunk_info.sample_rate | int | 是 | 采样率 |
| chunk_info.channels | int | 是 | 声道数 |
| chunk_info.bit_rate | int | 否 | 比特率 |
| audio_data | string | 是 | base64编码的音频数据 |
| checksum | string | 是 | MD5校验和 |

**响应信息**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "chunk_id": "chunk_uuid",
    "session_id": "session_uuid",
    "minio_path": "audio/2024/01/15/AA-BB-CC-DD-EE-FF/session_001/chunk_001.wav",
    "uploaded_at": 1640995235000
  }
}
```

**错误响应**:
```json
{
  "code": 400,
  "message": "invalid request",
  "data": null
}
```

### 2. 创建音频会话

**接口描述**: 创建新的音频会话

**请求信息**:
- **URL**: `POST /api/v1/audio/session`
- **Content-Type**: `application/json`

**请求参数**:
```json
{
  "device_id": "AA:BB:CC:DD:EE:FF",
  "session_id": "uuid-optional",
  "start_time": 1640995200000
}
```

**响应信息**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "session_id": "session_uuid",
    "device_id": "AA:BB:CC:DD:EE:FF",
    "start_time": 1640995200000,
    "status": 0,
    "created_at": 1640995200000
  }
}
```

### 3. 获取音频会话

**接口描述**: 根据会话ID获取音频会话信息

**请求信息**:
- **URL**: `GET /api/v1/audio/session/{session_id}`
- **参数**: session_id (路径参数)

**响应信息**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "session_id": "session_uuid",
    "device_id": "AA:BB:CC:DD:EE:FF",
    "start_time": 1640995200000,
    "end_time": 1640995800000,
    "total_chunks": 20,
    "status": 1,
    "created_at": 1640995200000,
    "updated_at": 1640995800000
  }
}
```

### 4. 获取音频会话列表

**接口描述**: 分页获取音频会话列表

**请求信息**:
- **URL**: `GET /api/v1/audio/sessions`
- **参数**: 查询参数

**查询参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| device_id | string | 否 | 设备MAC地址 |
| start_time | int64 | 否 | 开始时间(毫秒) |
| end_time | int64 | 否 | 结束时间(毫秒) |
| status | int | 否 | 会话状态(0=进行中,1=已完成,2=已失败) |
| offset | int | 否 | 偏移量，默认0 |
| limit | int | 否 | 每页数量，默认50，最大1000 |

**响应信息**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total": 100,
    "items": [
      {
        "id": 1,
        "session_id": "session_uuid",
        "device_id": "AA:BB:CC:DD:EE:FF",
        "start_time": 1640995200000,
        "end_time": 1640995800000,
        "total_chunks": 20,
        "status": 1,
        "created_at": 1640995200000,
        "updated_at": 1640995800000
      }
    ]
  }
}
```

### 5. 获取音频块列表

**接口描述**: 根据会话ID获取音频块列表

**请求信息**:
- **URL**: `GET /api/v1/audio/chunks`
- **参数**: 查询参数

**查询参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| session_id | string | 是 | 会话ID |
| start_time | int64 | 否 | 开始时间(毫秒) |
| end_time | int64 | 否 | 结束时间(毫秒) |
| offset | int | 否 | 偏移量，默认0 |
| limit | int | 否 | 每页数量，默认50，最大1000 |

**响应信息**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total": 20,
    "items": [
      {
        "id": 1,
        "session_id": "session_uuid",
        "chunk_index": 1,
        "device_id": "AA:BB:CC:DD:EE:FF",
        "start_time": 1640995200000,
        "end_time": 1640995230000,
        "duration": 30000,
        "file_size": 1024000,
        "format": "wav",
        "sample_rate": 16000,
        "channels": 1,
        "minio_path": "audio/2024/01/15/AA-BB-CC-DD-EE-FF/session_001/chunk_001.wav",
        "checksum": "md5_hash",
        "status": 1,
        "created_at": 1640995200000
      }
    ]
  }
}
```

### 6. 时间同步

**接口描述**: 设备与服务端时间同步

**请求信息**:
- **URL**: `POST /api/v1/audio/sync-time`
- **Content-Type**: `application/json`

**请求参数**:
```json
{
  "device_id": "AA:BB:CC:DD:EE:FF",
  "device_time": 1640995200000,
  "server_time": 1640995201000
}
```

**响应信息**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "device_id": "AA:BB:CC:DD:EE:FF",
    "device_time": 1640995200000,
    "server_time": 1640995201000,
    "offset": -1000,
    "last_sync": 1640995201000
  }
}
```

## 状态码说明

| 状态码 | 说明 |
|--------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

## 错误码说明

| 错误码 | 说明 |
|--------|------|
| 10001 | 请求参数无效 |
| 10002 | 资源不存在 |
| 10003 | 服务器内部错误 |
| 10004 | 时间戳验证失败 |
| 10005 | 文件格式不支持 |
| 10006 | 文件大小超限 |
| 10007 | 校验和验证失败 |

## 使用示例

### 设备端上传音频块流程

```javascript
// 1. 创建音频会话
const createSession = async (deviceId, startTime) => {
  const response = await fetch('/api/v1/audio/session', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      device_id: deviceId,
      start_time: startTime
    })
  });
  return response.json();
};

// 2. 上传音频块
const uploadChunk = async (deviceId, sessionId, chunkData) => {
  const response = await fetch('/api/v1/audio/upload-chunk', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      device_id: deviceId,
      session_id: sessionId,
      chunk_info: {
        chunk_index: chunkData.index,
        start_time: chunkData.startTime,
        end_time: chunkData.endTime,
        duration: chunkData.duration,
        format: 'wav',
        sample_rate: 16000,
        channels: 1
      },
      audio_data: chunkData.base64Data,
      checksum: chunkData.md5Hash
    })
  });
  return response.json();
};

// 3. 时间同步
const syncTime = async (deviceId) => {
  const deviceTime = Date.now();
  const response = await fetch('/api/v1/audio/sync-time', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      device_id: deviceId,
      device_time: deviceTime,
      server_time: Date.now()
    })
  });
  return response.json();
};
```

## 配置说明

### MinIO配置
```yaml
audio:
  minio:
    endpoint: "localhost:9000"
    access_key: "minioadmin"
    secret_key: "minioadmin"
    use_ssl: false
    bucket_name: "audio-bucket"
    region: "us-east-1"
    timeout: 30
```

### 音频处理配置
```yaml
audio:
  processing:
    max_chunk_size: 10485760  # 10MB
    allowed_formats: ["wav", "mp3", "pcm"]
    max_duration: 300000      # 5分钟
    min_duration: 1000        # 1秒
```

### 时间同步配置
```yaml
audio:
  time_sync:
    max_offset: 5000          # 最大时间偏移5秒
    sync_interval: 3600000    # 1小时同步一次
    tolerance: 1000           # 时间容差1秒
```

## 存储结构

音频文件在MinIO中的存储结构：
```
audio-bucket/
├── 2024/
│   ├── 01/
│   │   ├── 15/
│   │   │   ├── AA-BB-CC-DD-EE-FF/          # 设备目录
│   │   │   │   ├── session_001/             # 会话目录
│   │   │   │   │   ├── chunk_001.wav        # 00:00-00:30
│   │   │   │   │   ├── chunk_002.wav        # 00:30-01:00
│   │   │   │   │   ├── chunk_003.wav        # 01:00-01:30
│   │   │   │   │   └── session_metadata.json
│   │   │   │   └── session_002/
│   │   │   └── CC-DD-EE-FF-AA-BB/
│   │   └── 16/
│   └── 02/
```

## 注意事项

1. **时间戳精度**: 使用毫秒级时间戳，确保时间同步准确
2. **文件大小限制**: 单个音频块最大10MB
3. **格式支持**: 目前支持WAV、MP3、PCM格式
4. **并发上传**: 支持多个设备同时上传音频块
5. **错误重试**: 上传失败时建议重试机制
6. **时间同步**: 建议定期进行时间同步，确保时间戳准确性
