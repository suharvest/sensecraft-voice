# OSS文件上传服务

## 功能概述

这是一个基于MinIO的简单文件上传服务，支持：
- HTTP文件上传到MinIO
- 文件下载和访问
- 文件列表管理
- 文件删除（软删除）

## 快速开始

### 1. 启动服务

```bash
# 编译
go build -o server cmd/server.go

# 启动服务
./server
```

服务将在 `http://localhost:3008` 启动

### 2. 测试API

```bash
# 运行测试脚本
./test_oss_api.sh
```

## API接口

### 上传文件
```bash
curl -X POST "http://localhost:3008/api/v1/oss/upload" \
  -F "file=@/path/to/your/file.jpg" \
  -F "uploader=user123"
```

### 获取文件列表
```bash
curl "http://localhost:3008/api/v1/oss/files?limit=10"
```

### 下载文件
```bash
curl "http://localhost:3008/api/v1/oss/download/1" -o downloaded_file.jpg
```

### 删除文件
```bash
curl -X DELETE "http://localhost:3008/api/v1/oss/file/1"
```

## 配置说明

在 `config.yaml` 中配置MinIO连接信息：

```yaml
oss:
  minio:
    endpoint: "localhost:9000"
    access_key: "minioadmin"
    secret_key: "minioadmin"
    use_ssl: false
    bucket_name: "oss-bucket"
    region: "us-east-1"
    timeout: 30
  
  processing:
    max_file_size: 10485760  # 10MB
```

## 存储结构

文件在MinIO中的存储路径：
```
oss-bucket/
├── files/
│   ├── 2024/
│   │   ├── 01/
│   │   │   ├── 15/
│   │   │   │   ├── uuid-filename1.jpg
│   │   │   │   ├── uuid-filename2.pdf
│   │   │   │   └── uuid-filename3.mp4
```

## 数据库表

系统会自动创建 `file_uploads` 表来记录文件元数据。

## 注意事项

1. 确保MinIO服务正在运行
2. 文件大小限制为10MB
3. 支持所有文件格式
4. 文件通过UUID重命名避免冲突
5. 删除文件为软删除，不会立即删除MinIO中的文件
