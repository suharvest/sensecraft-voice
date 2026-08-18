# OSS文件上传API文档

## 概述

本文档描述了基于MinIO的对象存储服务(OSS)API接口，支持文件上传、下载、管理和访问功能。

## 基础信息

- **基础URL**: `/api/v1/oss`
- **认证方式**: 待定
- **数据格式**: JSON (除文件上传外)
- **字符编码**: UTF-8

## 接口列表

### 1. 上传文件

**接口描述**: 上传文件到MinIO存储

**请求信息**:
- **URL**: `POST /api/v1/oss/upload`
- **Content-Type**: `multipart/form-data`

**请求参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| file | file | 是 | 要上传的文件 |
| uploader | string | 否 | 上传者标识 |

**响应信息**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "file_id": 1,
    "file_name": "uuid-filename.ext",
    "original_name": "original-filename.ext",
    "file_size": 1024000,
    "content_type": "image/jpeg",
    "minio_path": "files/2024/01/15/uuid-filename.ext",
    "download_url": "/api/v1/oss/download/1",
    "uploaded_at": 1640995200000
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

### 2. 获取文件信息

**接口描述**: 根据文件ID获取文件详细信息

**请求信息**:
- **URL**: `GET /api/v1/oss/file/{file_id}`
- **参数**: file_id (路径参数)

**响应信息**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "file_name": "uuid-filename.ext",
    "original_name": "original-filename.ext",
    "file_size": 1024000,
    "content_type": "image/jpeg",
    "minio_path": "files/2024/01/15/uuid-filename.ext",
    "checksum": "md5_hash",
    "uploader": "user123",
    "status": 1,
    "created_at": 1640995200000,
    "updated_at": 1640995200000
  }
}
```

### 3. 获取文件列表

**接口描述**: 分页获取文件列表

**请求信息**:
- **URL**: `GET /api/v1/oss/files`
- **参数**: 查询参数

**查询参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| uploader | string | 否 | 上传者标识 |
| status | int | 否 | 文件状态(1=正常,0=已删除) |
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
        "file_name": "uuid-filename.ext",
        "original_name": "original-filename.ext",
        "file_size": 1024000,
        "content_type": "image/jpeg",
        "minio_path": "files/2024/01/15/uuid-filename.ext",
        "checksum": "md5_hash",
        "uploader": "user123",
        "status": 1,
        "created_at": 1640995200000,
        "updated_at": 1640995200000
      }
    ]
  }
}
```

### 4. 下载文件

**接口描述**: 根据文件ID下载文件

**请求信息**:
- **URL**: `GET /api/v1/oss/download/{file_id}`
- **参数**: file_id (路径参数)

**响应信息**:
- **Content-Type**: 根据文件类型自动设置
- **Content-Disposition**: attachment; filename=original-filename.ext
- **Content-Length**: 文件大小
- **Body**: 文件二进制内容

### 5. 删除文件

**接口描述**: 根据文件ID删除文件

**请求信息**:
- **URL**: `DELETE /api/v1/oss/file/{file_id}`
- **参数**: file_id (路径参数)

**响应信息**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "file deleted successfully"
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
| 10005 | 文件格式不支持 |
| 10006 | 文件大小超限 |
| 10007 | 校验和验证失败 |
| 10008 | MinIO上传失败 |

## 使用示例

### 1. 上传文件

```bash
# 使用curl上传文件
curl -X POST "http://localhost:3008/api/v1/oss/upload" \
  -F "file=@/path/to/your/file.jpg" \
  -F "uploader=user123"
```

```javascript
// 使用JavaScript上传文件
const formData = new FormData();
formData.append('file', fileInput.files[0]);
formData.append('uploader', 'user123');

fetch('/api/v1/oss/upload', {
  method: 'POST',
  body: formData
})
.then(response => response.json())
.then(data => console.log(data));
```

### 2. 获取文件列表

```bash
# 获取所有文件
curl "http://localhost:3008/api/v1/oss/files?limit=10"

# 获取特定上传者的文件
curl "http://localhost:3008/api/v1/oss/files?uploader=user123&limit=10"
```

### 3. 下载文件

```bash
# 下载文件
curl "http://localhost:3008/api/v1/oss/download/1" -o downloaded_file.jpg
```

```javascript
// 使用JavaScript下载文件
window.open('/api/v1/oss/download/1', '_blank');
```

### 4. 删除文件

```bash
# 删除文件
curl -X DELETE "http://localhost:3008/api/v1/oss/file/1"
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

### 文件处理配置
```yaml
audio:
  processing:
    max_chunk_size: 10485760  # 10MB
    allowed_formats: ["wav", "mp3", "pcm"]
    max_duration: 300000      # 5分钟
    min_duration: 1000        # 1秒
```

## 存储结构

文件在MinIO中的存储结构：
```
audio-bucket/
├── files/
│   ├── 2024/
│   │   ├── 01/
│   │   │   ├── 15/
│   │   │   │   ├── uuid-filename1.jpg
│   │   │   │   ├── uuid-filename2.pdf
│   │   │   │   └── uuid-filename3.mp4
│   │   │   └── 16/
│   │   └── 02/
│   └── 2023/
```

## 数据库表结构

### file_uploads 表
```sql
CREATE TABLE `file_uploads` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `file_name` varchar(255) NOT NULL COMMENT '存储文件名',
  `original_name` varchar(255) NOT NULL COMMENT '原始文件名',
  `file_size` bigint NOT NULL COMMENT '文件大小(字节)',
  `content_type` varchar(100) NOT NULL COMMENT '文件类型',
  `minio_path` varchar(500) NOT NULL COMMENT 'MinIO存储路径',
  `checksum` varchar(32) NOT NULL COMMENT '文件MD5校验和',
  `uploader` varchar(64) DEFAULT '' COMMENT '上传者',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '1=正常,0=已删除',
  `created_at` bigint NOT NULL COMMENT '创建时间(毫秒)',
  `updated_at` bigint NOT NULL COMMENT '更新时间(毫秒)',
  PRIMARY KEY (`id`),
  KEY `idx_uploader` (`uploader`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

## 注意事项

1. **文件大小限制**: 单个文件最大10MB
2. **支持格式**: 支持所有文件格式
3. **并发上传**: 支持多个用户同时上传文件
4. **错误重试**: 上传失败时建议重试机制
5. **文件安全**: 文件通过UUID重命名，避免文件名冲突
6. **软删除**: 删除文件时只标记状态，不立即删除MinIO中的文件
7. **校验和**: 自动计算并验证文件MD5校验和
8. **路径生成**: 按年/月/日自动组织文件存储路径
