# OpenAI API v2 接口文档

## 📋 概述

OpenAI API v2 提供了完整的AI聊天功能，支持上下文管理、流式响应、会话管理和系统提示词自动注入。该接口基于现有的chat表结构实现上下文管理，确保多轮对话的连贯性，并支持通过系统提示词自定义AI的行为模式。

## 🔗 基础信息

- **基础URL**: `http://localhost:3008/api/v2/openai`
- **认证方式**: 无需认证（可根据需要添加）
- **数据格式**: JSON
- **字符编码**: UTF-8

## 📚 接口列表

### 1. 发送聊天消息

**接口描述**: 发送消息给AI并获取回复，支持系统提示词自动注入

**请求信息**:
- **URL**: `/api/v2/openai/chat/send`
- **方法**: `POST`
- **Content-Type**: `application/json`

**请求参数**:
```json
{
  "session_id": "string",              // 可选，会话ID，不提供则自动创建
  "message": "string",                 // 必填，用户消息内容
  "user_id": "string",                 // 必填，用户ID
  "system_prompt_id": 1,               // 可选，系统提示词ID
  "system_prompt_content": "string"    // 可选，系统提示词内容
}
```

**请求示例**:

使用默认系统提示词：
```bash
curl -X POST http://localhost:3008/api/v2/openai/chat/send \
  -H 'Content-Type: application/json' \
  -d '{
    "message": "你好，请介绍一下Go语言的特点",
    "user_id": "user_123"
  }'
```

使用指定的系统提示词ID：
```bash
curl -X POST http://localhost:3008/api/v2/openai/chat/send \
  -H 'Content-Type: application/json' \
  -d '{
    "message": "请帮我写一个Go程序",
    "user_id": "user_123",
    "system_prompt_id": 2
  }'
```

使用直接提供的系统提示词内容：
```bash
curl -X POST http://localhost:3008/api/v2/openai/chat/send \
  -H 'Content-Type: application/json' \
  -d '{
    "message": "请解释一下微服务架构",
    "user_id": "user_123",
    "system_prompt_content": "你是一个专业的软件架构师，擅长微服务设计和最佳实践。"
  }'
```

**响应格式**:
```json
{
  "code": 200,
  "result": {
    "session_id": "openai_1703123456789",
    "message": "Go语言是一种开源的编程语言，由Google开发...",
    "usage": {
      "prompt_tokens": 15,
      "completion_tokens": 120,
      "total_tokens": 135
    },
    "created_at": 1703123456789
  }
}
```

**响应字段说明**:
- `session_id`: 会话ID，用于后续对话
- `message`: AI回复内容
- `usage`: Token使用情况
  - `prompt_tokens`: 输入Token数量
  - `completion_tokens`: 输出Token数量
  - `total_tokens`: 总Token数量
- `created_at`: 创建时间戳（毫秒）

---

### 2. 流式聊天消息

**接口描述**: 发送消息给AI并获取流式回复，支持系统提示词自动注入

**请求信息**:
- **URL**: `/api/v2/openai/chat/stream`
- **方法**: `POST`
- **Content-Type**: `application/json`
- **响应格式**: `text/event-stream` (Server-Sent Events)

**请求参数**:
```json
{
  "session_id": "string",              // 可选，会话ID
  "message": "string",                 // 必填，用户消息内容
  "user_id": "string",                 // 必填，用户ID
  "system_prompt_id": 1,               // 可选，系统提示词ID
  "system_prompt_content": "string"    // 可选，系统提示词内容
}
```

**请求示例**:
```bash
curl -X POST http://localhost:3008/api/v2/openai/chat/stream \
  -H 'Content-Type: application/json' \
  -d '{
    "message": "请详细解释一下微服务架构",
    "user_id": "user_123"
  }'
```

**响应格式** (SSE):
```
data: {"event": "message", "data": {"session_id": "openai_1703123456789", "content": "微服务架构是一种", "timestamp": 1703123456789}}

data: {"event": "message", "data": {"session_id": "openai_1703123456789", "content": "将应用程序构建为", "timestamp": 1703123456790}}

data: {"event": "completed", "data": {"session_id": "openai_1703123456789", "timestamp": 1703123456791}}
```

**事件类型**:
- `message`: 流式消息内容
- `completed`: 消息完成
- `error`: 错误信息

---

### 3. 获取聊天历史

**接口描述**: 获取指定会话的聊天历史记录

**请求信息**:
- **URL**: `/api/v2/openai/chat/history/{session_id}`
- **方法**: `GET`

**路径参数**:
- `session_id`: 会话ID

**查询参数**:
- `limit`: 可选，返回消息数量限制，默认20，最大100

**请求示例**:
```bash
curl -X GET 'http://localhost:3008/api/v2/openai/chat/history/openai_1703123456789?limit=10' \
  -H 'Content-Type: application/json'
```

**响应格式**:
```json
{
  "code": 200,
  "result": [
    {
      "id": 1,
      "session_id": "openai_1703123456789",
      "message_id": "msg_1703123456789",
      "event": "user",
      "content": "你好，请介绍一下Go语言的特点",
      "data": "{\"user_id\": \"user_123\", \"timestamp\": 1703123456789}",
      "created_at": 1703123456789
    },
    {
      "id": 2,
      "session_id": "openai_1703123456789",
      "message_id": "msg_1703123456790",
      "event": "assistant",
      "content": "Go语言是一种开源的编程语言...",
      "data": "{\"user_id\": \"user_123\", \"timestamp\": 1703123456790}",
      "created_at": 1703123456790
    }
  ]
}
```

**响应字段说明**:
- `id`: 消息ID
- `session_id`: 会话ID
- `message_id`: 消息唯一标识
- `event`: 消息类型（user/assistant/system）
- `content`: 消息内容
- `data`: 额外数据（JSON字符串）
- `created_at`: 创建时间戳

---

### 4. 创建新会话

**接口描述**: 创建一个新的聊天会话

**请求信息**:
- **URL**: `/api/v2/openai/chat/session`
- **方法**: `POST`
- **Content-Type**: `application/json`

**请求参数**:
```json
{
  "user_id": "string"  // 必填，用户ID
}
```

**请求示例**:
```bash
curl -X POST http://localhost:3008/api/v2/openai/chat/session \
  -H 'Content-Type: application/json' \
  -d '{
    "user_id": "user_123"
  }'
```

**响应格式**:
```json
{
  "code": 200,
  "result": {
    "session_id": "openai_1703123456789"
  }
}
```

---

### 5. 关闭会话

**接口描述**: 关闭指定的聊天会话

**请求信息**:
- **URL**: `/api/v2/openai/chat/session/{session_id}`
- **方法**: `DELETE`

**路径参数**:
- `session_id`: 会话ID

**请求示例**:
```bash
curl -X DELETE http://localhost:3008/api/v2/openai/chat/session/openai_1703123456789 \
  -H 'Content-Type: application/json'
```

**响应格式**:
```json
{
  "code": 200,
  "result": {
    "message": "会话已关闭"
  }
}
```

## 🤖 系统提示词功能

### 系统提示词选择机制

系统提示词支持三种选择方式，按优先级排序：

1. **直接内容**：通过 `system_prompt_content` 字段直接提供系统提示词内容
2. **指定ID**：通过 `system_prompt_id` 字段指定已创建的系统提示词ID
3. **默认提示词**：当以上两种都未提供时，自动使用名为"default"且激活状态的系统提示词

### 注入机制

- **注入时机**：在每次调用 `/api/v2/openai/chat/send` 或 `/api/v2/openai/chat/stream` 时
- **注入方式**：作为"system"角色的消息添加到对话上下文中
- **管理接口**：通过 `/api/v2/openai/system-prompts` 系列接口管理系统提示词

### 系统提示词管理

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/v2/openai/system-prompts` | POST | 创建系统提示词 |
| `/api/v2/openai/system-prompts` | GET | 获取系统提示词列表 |
| `/api/v2/openai/system-prompts/:id` | GET | 获取单个系统提示词 |
| `/api/v2/openai/system-prompts/:id` | PUT | 更新系统提示词 |
| `/api/v2/openai/system-prompts/:id` | DELETE | 删除系统提示词 |

详细文档请参考：[系统提示词管理 API v2 详细文档](system_prompts_api_v2.md)

## 🔧 配置说明

### OpenAI配置参数

在 `config.yaml` 中配置OpenAI相关参数：

```yaml
openai:
  # OpenAI API 密钥
  api_key: "sk-your-openai-api-key-here"
  # API 基础地址
  base_url: "https://api.openai.com/v1"
  # 请求超时时间（秒）
  timeout: 120
  # 最大token数量
  max_tokens: 4096
  # 温度参数（0-2，控制回复的随机性）
  temperature: 0.7
  # 默认模型
  model: "gpt-3.5-turbo"
```

### 支持的模型

- `gpt-3.5-turbo`: 默认模型，性价比高
- `gpt-4`: 更强大的模型
- `gpt-4-turbo`: 最新版本
- 其他OpenAI支持的模型

## 📊 错误码说明

| 错误码 | 说明 | 解决方案 |
|--------|------|----------|
| 400 | 请求参数错误 | 检查必填字段是否提供 |
| 500 | 服务器内部错误 | 检查服务状态和配置 |
| 401 | 认证失败 | 检查API密钥配置 |
| 429 | 请求频率限制 | 降低请求频率 |
| 503 | 服务不可用 | 检查OpenAI API状态 |

## 🔍 使用示例

### JavaScript示例

```javascript
// 发送聊天消息
async function sendMessage(message, userId, sessionId = null) {
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
  return data.result;
}

// 流式聊天
function streamChat(message, userId, sessionId = null) {
  const eventSource = new EventSource('/api/v2/openai/chat/stream');
  
  eventSource.onmessage = function(event) {
    const data = JSON.parse(event.data);
    if (data.event === 'message') {
      console.log('AI回复:', data.data.content);
    } else if (data.event === 'completed') {
      console.log('对话完成');
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
```

### Python示例

```python
import requests
import json

# 发送聊天消息
def send_message(message, user_id, session_id=None):
    url = "http://localhost:3008/api/v2/openai/chat/send"
    data = {
        "message": message,
        "user_id": user_id,
        "session_id": session_id
    }
    
    response = requests.post(url, json=data)
    return response.json()

# 获取聊天历史
def get_chat_history(session_id, limit=20):
    url = f"http://localhost:3008/api/v2/openai/chat/history/{session_id}"
    params = {"limit": limit}
    
    response = requests.get(url, params=params)
    return response.json()

# 使用示例
result = send_message("你好，请介绍一下Python", "user_123")
print(f"AI回复: {result['result']['message']}")
print(f"会话ID: {result['result']['session_id']}")

# 管理系统提示词
def create_system_prompt(name, content):
    url = "http://localhost:3008/api/v2/openai/system-prompts"
    data = {
        "name": name,
        "content": content,
        "is_active": True
    }
    response = requests.post(url, json=data)
    return response.json()

# 创建编程助手提示词
prompt_result = create_system_prompt(
    "programmer", 
    "你是一个专业的编程助手，擅长多种编程语言。请提供准确、详细的代码建议。"
)
print(f"创建提示词结果: {prompt_result}")
```

## 🚀 最佳实践

### 1. 会话管理
- 为每个用户或对话主题创建独立的会话
- 定期清理不需要的会话以节省存储空间
- 使用有意义的会话ID便于管理

### 2. 上下文优化
- 系统会自动管理上下文，保持最近20条消息
- 对于长对话，考虑定期创建新会话
- 使用系统提示词设置AI的行为模式和角色定义
- 通过系统提示词管理接口自定义AI的回复风格

### 3. 错误处理
- 实现重试机制处理临时错误
- 监控API使用量和成本
- 设置合理的超时时间

### 4. 性能优化
- 使用流式接口提升用户体验
- 合理设置max_tokens避免过长回复
- 根据需求选择合适的模型
- 合理使用系统提示词，避免过长的提示词影响性能

## 📈 监控和调试

### 日志记录
系统会记录以下信息：
- API调用次数和响应时间
- Token使用情况
- 错误信息和堆栈跟踪
- 系统提示词注入情况

### 性能监控
- 响应时间监控
- 并发请求处理能力
- 数据库查询性能

## 🔒 安全注意事项

1. **API密钥保护**: 确保OpenAI API密钥安全存储
2. **输入验证**: 对用户输入进行适当验证和过滤
3. **访问控制**: 根据需要实现用户认证和授权
4. **数据隐私**: 注意聊天内容的隐私保护

## 📞 技术支持

如有问题或建议，请联系开发团队或查看项目文档。

---

**版本**: v2.1  
**更新时间**: 2024年12月  
**维护者**: SenseCraft Voice Service Team  
**新增功能**: 系统提示词自动注入和管理
