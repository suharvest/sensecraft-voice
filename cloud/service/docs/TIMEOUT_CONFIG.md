# OpenAI API 超时配置说明

## ⏰ 超时时间设置

### 当前配置

OpenAI API的超时时间已设置为 **2分钟（120秒）**，这个设置适用于大多数AI模型调用场景。

### 配置位置

1. **主配置文件**: `config.yaml`
```yaml
openai:
  timeout: 120  # 2分钟超时
```

2. **默认配置**: `pkg/util/openai/types.go`
```go
func DefaultConfig() *Config {
    return &Config{
        Timeout: 120,  // 2分钟超时
        // ... 其他配置
    }
}
```

## 🔧 超时时间说明

### 为什么选择2分钟？

1. **AI模型响应时间**: 大多数AI模型在2分钟内能够完成响应
2. **用户体验**: 2分钟是一个合理的等待时间，不会让用户感到过长
3. **资源平衡**: 避免过长的连接占用服务器资源
4. **错误处理**: 如果模型真的需要更长时间，可能是出现了问题

### 不同场景的超时需求

| 场景 | 建议超时时间 | 说明 |
|------|-------------|------|
| 简单问答 | 30-60秒 | 快速响应 |
| 复杂分析 | 120秒 | 当前设置 |
| 长文本生成 | 180-300秒 | 需要更长时间 |
| 流式响应 | 无限制 | 实时流式输出 |

## ⚙️ 自定义超时时间

### 方法1: 修改配置文件

编辑 `config.yaml`:
```yaml
openai:
  timeout: 180  # 改为3分钟
```

### 方法2: 环境变量

```bash
export OPENAI_TIMEOUT=180
```

### 方法3: 代码中动态设置

```go
config := &openai.Config{
    Timeout: 180,  // 3分钟
    // ... 其他配置
}
```

## 🚨 超时处理

### 客户端超时

当API调用超时时，会返回以下错误：

```json
{
  "code": 500,
  "message": "AI服务调用失败: 发送请求失败: context deadline exceeded"
}
```

### 流式响应超时

流式响应有特殊的超时处理：
- 连接建立后不会立即超时
- 如果长时间没有数据，会主动关闭连接
- 客户端会收到 `error` 事件

## 📊 性能监控

### 监控指标

1. **平均响应时间**: 监控API调用的平均耗时
2. **超时率**: 统计超时请求的比例
3. **成功率**: 监控API调用的成功率

### 日志记录

系统会记录以下超时相关信息：
```
INFO: OpenAI API调用成功，使用tokens: 150, 耗时: 2.3秒
WARN: OpenAI API调用超时，耗时: 120秒
ERROR: OpenAI API调用失败: context deadline exceeded
```

## 🔍 故障排除

### 常见超时问题

1. **网络问题**
   - 检查网络连接
   - 确认API服务可访问

2. **模型负载过高**
   - 尝试使用不同的模型
   - 降低请求频率

3. **请求内容过长**
   - 减少输入内容长度
   - 调整max_tokens参数

### 调试方法

1. **启用调试日志**
```yaml
openai:
  enable_debug: true
```

2. **监控网络请求**
```bash
# 使用curl测试
curl -w "@curl-format.txt" -X POST http://localhost:3008/api/v2/openai/chat/send
```

3. **检查API状态**
```bash
# 检查服务健康状态
curl http://localhost:3008/healthz
```

## 🚀 最佳实践

### 1. 合理设置超时时间

- 根据业务需求调整
- 考虑网络环境
- 平衡用户体验和系统资源

### 2. 实现重试机制

```go
// 示例重试逻辑
for i := 0; i < 3; i++ {
    resp, err := client.ChatWithContext(sessionID, message, userID)
    if err == nil {
        return resp, nil
    }
    if !isTimeoutError(err) {
        return nil, err
    }
    time.Sleep(time.Duration(i+1) * time.Second)
}
```

### 3. 提供用户反馈

- 显示加载状态
- 提供取消选项
- 设置合理的等待提示

## 📝 相关文件

- `config.yaml` - 主配置文件
- `pkg/util/openai/types.go` - 默认配置
- `pkg/util/openai/client.go` - HTTP客户端实现
- `pkg/controller/openai/openai.go` - 控制器实现

---

**更新时间**: 2024年12月  
**当前超时**: 120秒（2分钟）  
**配置状态**: ✅ 已生效
