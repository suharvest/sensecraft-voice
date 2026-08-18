# 远程配置动态修改 API 文档

## 概述

本文档描述了如何通过 REST API 动态修改远程服务配置，无需重启服务即可生效。

## API 接口

### 1. 获取远程配置

**接口**: `GET /v1/voice/config/remote`

**描述**: 获取当前的远程服务配置信息

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "result": {
    "base_url": "http://voice-service.example.com:3008",
    "enabled": true
  }
}
```

### 2. 更新远程配置

**接口**: `PUT /v1/voice/config/remote`

**描述**: 动态更新远程服务的基础URL，修改后立即生效

**请求体**:
```json
{
  "base_url": "http://192.168.1.100:3008"
}
```

**参数说明**:
- `base_url` (string, 必需): 新的远程服务基础URL，支持 http 和 https 协议

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "result": {
    "message": "Remote config updated successfully",
    "base_url": "http://192.168.1.100:3008",
    "timestamp": 1703123456789
  }
}
```

**错误响应示例**:
```json
{
  "code": 400,
  "message": "failed",
  "result": {
    "error": "invalid remote URL: unsupported protocol: ftp, only http and https are supported"
  }
}
```

### 3. 测试远程连接

**接口**: `POST /v1/voice/config/remote/test`

**描述**: 测试指定远程服务的连接可达性，由后端发起请求避免CORS问题

**请求体**:
```json
{
  "base_url": "http://192.168.1.100:3008"
}
```

**参数说明**:
- `base_url` (string, 必需): 要测试的远程服务基础URL

**成功响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "result": {
    "reachable": true,
    "status_code": 200,
    "response_time_ms": 150,
    "test_url": "http://192.168.1.100:3008/v1/voice/test-connection",
    "message": "连接成功"
  }
}
```

**失败响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "result": {
    "reachable": false,
    "status_code": 0,
    "response_time_ms": 0,
    "test_url": "",
    "message": "连接失败: dial tcp 192.168.1.100:3008: connect: connection refused"
  }
}
```

**错误响应示例**:
```json
{
  "code": 400,
  "message": "failed",
  "result": {
    "error": "invalid URL format"
  }
}
```

## 使用示例

### 使用 curl 命令

```bash
# 获取当前配置
curl -X GET "http://localhost:8080/v1/voice/config/remote"

# 更新配置
curl -X PUT "http://localhost:8080/v1/voice/config/remote" \
  -H "Content-Type: application/json" \
  -d '{"base_url": "http://192.168.1.100:3008"}'

# 测试远程连接
curl -X POST "http://localhost:8080/v1/voice/config/remote/test" \
  -H "Content-Type: application/json" \
  -d '{"base_url": "http://192.168.1.100:3008"}'
```

### 使用测试脚本

```bash
# 运行测试脚本
./test_remote_config.sh

# 指定服务器地址
./test_remote_config.sh http://192.168.1.100:8080
```

## 功能特性

### 1. 动态生效
- 配置修改后立即生效，无需重启服务
- 自动重建远程连接
- 保持ASR结果订阅

### 2. 配置验证
- URL格式验证
- 协议支持检查（仅支持 http/https）
- 主机名和端口验证

### 3. 连接测试
- 后端发起连接测试，避免CORS问题
- 支持多个测试端点（/v1/voice/test-connection, /, /health）
- 200或404状态码表示服务可达
- 记录响应时间，便于性能诊断
- 详细的连接状态和错误信息

### 4. 错误处理
- 配置更新失败时自动回滚
- 详细的错误信息提示
- 连接超时保护

### 5. 线程安全
- 使用读写锁保护配置访问
- 原子性配置更新
- 并发安全

## 实现架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   API Router    │───▶│   Options       │───▶│  Voice Manager  │
│                 │    │                 │    │                 │
│ GET/PUT config  │    │ UpdateRemote    │    │ UpdateRemote    │
│                 │    │ Config()        │    │ Sink()          │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       ▼
         │                       │              ┌─────────────────┐
         │                       │              │  Remote Sink    │
         │                       │              │                 │
         │                       │              │ Close() +       │
         │                       │              │ Recreate()      │
         │                       │              └─────────────────┘
         │                       │
         ▼                       ▼
┌─────────────────┐    ┌─────────────────┐
│  HTTP Response  │    │  Config Update  │
│                 │    │  Logging        │
└─────────────────┘    └─────────────────┘
```

## 注意事项

1. **服务依赖**: 确保远程服务在配置更新时可用
2. **网络连接**: 配置更新会重建网络连接，可能有短暂中断
3. **ASR结果**: 配置更新期间ASR结果可能短暂丢失
4. **配置持久化**: 当前实现仅更新内存配置，重启后恢复原配置

## 故障排除

### 常见错误

1. **"invalid remote URL"**: URL格式不正确
2. **"remote stream is not enabled"**: 远程流功能未启用
3. **"failed to start new remote sink"**: 无法连接到新的远程服务

### 调试方法

1. 检查日志输出
2. 验证远程服务可用性
3. 确认网络连接
4. 检查配置格式

## 未来改进

1. **配置持久化**: 将配置修改保存到文件
2. **配置历史**: 记录配置修改历史
3. **健康检查**: 添加远程服务健康检查
4. **批量配置**: 支持批量配置更新
