# OpenAI API v2 使用指南

## 🎯 快速开始

### 1. 配置API密钥

在 `config.yaml` 中设置你的OpenAI API密钥：

```yaml
openai:
  api_key: "sk-your-openai-api-key-here"  # 替换为你的API密钥
  base_url: "https://api.openai.com/v1"
  timeout: 120
  max_tokens: 4096
  temperature: 0.7
  model: "gpt-3.5-turbo"
```

### 2. 启动服务

```bash
# 编译并运行
go build -o server cmd/server.go
./server

# 或者直接运行
go run cmd/server.go
```

### 3. 测试API

```bash
# 运行测试脚本
./test_openai_api.sh
```

## 📚 接口概览

| 接口 | 方法 | 路径 | 说明 |
|------|------|------|------|
| 发送消息 | POST | `/api/v2/openai/chat/send` | 发送消息并获取回复 |
| 流式聊天 | POST | `/api/v2/openai/chat/stream` | 流式获取AI回复 |
| 获取历史 | GET | `/api/v2/openai/chat/history/{session_id}` | 获取聊天历史 |
| 创建会话 | POST | `/api/v2/openai/chat/session` | 创建新会话 |
| 关闭会话 | DELETE | `/api/v2/openai/chat/session/{session_id}` | 关闭会话 |

## 🔧 使用示例

### JavaScript/前端

```javascript
// 发送聊天消息
async function chatWithAI(message, userId, sessionId = null) {
  const response = await fetch('/api/v2/openai/chat/send', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      message: message,
      user_id: userId,
      session_id: sessionId
    })
  });
  
  const data = await response.json();
  if (data.code === 200) {
    return data.result;
  } else {
    throw new Error(data.message);
  }
}

// 流式聊天
function streamChat(message, userId, sessionId, onMessage, onComplete) {
  const eventSource = new EventSource('/api/v2/openai/chat/stream');
  
  eventSource.onmessage = function(event) {
    const data = JSON.parse(event.data);
    if (data.event === 'message') {
      onMessage(data.data.content);
    } else if (data.event === 'completed') {
      onComplete(data.data.session_id);
      eventSource.close();
    } else if (data.event === 'error') {
      console.error('错误:', data.data);
      eventSource.close();
    }
  };
  
  // 发送消息
  fetch('/api/v2/openai/chat/stream', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      message: message,
      user_id: userId,
      session_id: sessionId
    })
  });
}

// 使用示例
chatWithAI("你好，请介绍一下Go语言", "user_123")
  .then(result => {
    console.log('AI回复:', result.message);
    console.log('会话ID:', result.session_id);
  })
  .catch(error => {
    console.error('错误:', error);
  });
```

### Python

```python
import requests
import json

class OpenAIClient:
    def __init__(self, base_url="http://localhost:3008"):
        self.base_url = f"{base_url}/api/v2/openai"
    
    def send_message(self, message, user_id, session_id=None):
        """发送聊天消息"""
        url = f"{self.base_url}/chat/send"
        data = {
            "message": message,
            "user_id": user_id,
            "session_id": session_id
        }
        
        response = requests.post(url, json=data)
        return response.json()
    
    def get_chat_history(self, session_id, limit=20):
        """获取聊天历史"""
        url = f"{self.base_url}/chat/history/{session_id}"
        params = {"limit": limit}
        
        response = requests.get(url, params=params)
        return response.json()
    
    def create_session(self, user_id):
        """创建新会话"""
        url = f"{self.base_url}/chat/session"
        data = {"user_id": user_id}
        
        response = requests.post(url, json=data)
        return response.json()
    
    def close_session(self, session_id):
        """关闭会话"""
        url = f"{self.base_url}/chat/session/{session_id}"
        
        response = requests.delete(url)
        return response.json()

# 使用示例
client = OpenAIClient()

# 创建会话
session = client.create_session("user_123")
session_id = session["result"]["session_id"]

# 发送消息
response = client.send_message("你好，请介绍一下Python", "user_123", session_id)
print(f"AI回复: {response['result']['message']}")

# 获取历史
history = client.get_chat_history(session_id)
print(f"历史记录: {len(history['result'])} 条消息")

# 关闭会话
client.close_session(session_id)
```

### cURL

```bash
# 1. 创建会话
curl -X POST http://localhost:3008/api/v2/openai/chat/session \
  -H 'Content-Type: application/json' \
  -d '{"user_id": "user_123"}'

# 2. 发送消息
curl -X POST http://localhost:3008/api/v2/openai/chat/send \
  -H 'Content-Type: application/json' \
  -d '{
    "message": "你好，请介绍一下Go语言",
    "user_id": "user_123",
    "session_id": "openai_1703123456789"
  }'

# 3. 获取历史
curl -X GET 'http://localhost:3008/api/v2/openai/chat/history/openai_1703123456789?limit=10' \
  -H 'Content-Type: application/json'

# 4. 流式聊天
curl -X POST http://localhost:3008/api/v2/openai/chat/stream \
  -H 'Content-Type: application/json' \
  -H 'Accept: text/event-stream' \
  -d '{
    "message": "请详细解释微服务架构",
    "user_id": "user_123",
    "session_id": "openai_1703123456789"
  }'
```

## 🔍 上下文管理

系统会自动管理对话上下文：

- **自动保存**: 所有用户消息和AI回复都会保存到数据库
- **上下文维护**: 自动保持最近20条消息作为上下文
- **会话隔离**: 不同会话的上下文完全隔离
- **持久化存储**: 使用现有的 `chat_sessions` 和 `chat_messages` 表

## ⚙️ 配置参数

### OpenAI配置

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `api_key` | string | - | OpenAI API密钥 |
| `base_url` | string | `https://api.openai.com/v1` | API基础URL |
| `timeout` | int | 120 | 请求超时时间（秒） |
| `max_tokens` | int | 4096 | 最大Token数量 |
| `temperature` | float | 0.7 | 温度参数（0-2） |
| `model` | string | `gpt-3.5-turbo` | 使用的模型 |

### 支持的模型

- `gpt-3.5-turbo`: 默认模型，性价比高
- `gpt-4`: 更强大的模型
- `gpt-4-turbo`: 最新版本
- 其他OpenAI支持的模型

## 🚨 错误处理

### 常见错误码

| 错误码 | 说明 | 解决方案 |
|--------|------|----------|
| 400 | 请求参数错误 | 检查必填字段 |
| 500 | 服务器内部错误 | 检查服务状态 |
| 401 | 认证失败 | 检查API密钥 |
| 429 | 请求频率限制 | 降低请求频率 |

### 错误响应格式

```json
{
  "code": 400,
  "message": "消息内容不能为空"
}
```

## 📊 监控和调试

### 日志记录

系统会记录以下信息：
- API调用次数和响应时间
- Token使用情况
- 错误信息和堆栈跟踪

### 性能监控

- 响应时间监控
- 并发请求处理能力
- 数据库查询性能

## 🔒 安全注意事项

1. **API密钥保护**: 确保OpenAI API密钥安全存储
2. **输入验证**: 对用户输入进行适当验证和过滤
3. **访问控制**: 根据需要实现用户认证和授权
4. **数据隐私**: 注意聊天内容的隐私保护

## 🚀 最佳实践

### 1. 会话管理
- 为每个用户或对话主题创建独立的会话
- 定期清理不需要的会话以节省存储空间
- 使用有意义的会话ID便于管理

### 2. 上下文优化
- 系统会自动管理上下文，保持最近20条消息
- 对于长对话，考虑定期创建新会话
- 使用系统消息设置AI的行为模式

### 3. 错误处理
- 实现重试机制处理临时错误
- 监控API使用量和成本
- 设置合理的超时时间

### 4. 性能优化
- 使用流式接口提升用户体验
- 合理设置max_tokens避免过长回复
- 根据需求选择合适的模型

## 📞 技术支持

如有问题或建议，请联系开发团队或查看项目文档。

---

**版本**: v2.0  
**更新时间**: 2024年12月  
**维护者**: SenseCraft Voice Service Team
