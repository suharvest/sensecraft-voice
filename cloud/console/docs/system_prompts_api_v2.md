# 系统提示词管理 API v2

## 概述

系统提示词管理API提供了对AI对话中系统级提示词的完整CRUD操作。系统提示词会在每次对话开始时自动注入到上下文中，为AI提供角色定义和行为指导。

## 数据模型

### SystemPrompt

```json
{
  "id": 1,
  "name": "default",
  "role": "system",
  "content": "你是一个有用的AI助手，请用中文回答问题。",
  "tags": "[\"default\", \"chinese\"]",
  "is_active": true,
  "version": 1,
  "created_at": 1703123456789,
  "updated_at": 1703123456789
}
```

### 字段说明

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int64 | - | 主键，自动生成 |
| name | string | 是 | 提示词名称，全局唯一 |
| role | string | 否 | 角色标识，默认为"system" |
| content | string | 是 | 提示词内容 |
| tags | string | 否 | JSON格式的标签数组 |
| is_active | bool | 否 | 是否激活，默认true |
| version | int | - | 版本号，更新时自动递增 |
| created_at | int64 | - | 创建时间戳(毫秒) |
| updated_at | int64 | - | 更新时间戳(毫秒) |

## API 接口

### 1. 创建系统提示词

**POST** `/api/v2/openai/system-prompts`

#### 请求体

```json
{
  "name": "default",
  "role": "system",
  "content": "你是一个有用的AI助手，请用中文回答问题。",
  "tags": "[\"default\", \"chinese\"]",
  "is_active": true
}
```

#### 响应

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "name": "default",
    "role": "system",
    "content": "你是一个有用的AI助手，请用中文回答问题。",
    "tags": "[\"default\", \"chinese\"]",
    "is_active": true,
    "version": 1,
    "created_at": 1703123456789,
    "updated_at": 1703123456789
  }
}
```

### 2. 获取系统提示词列表

**GET** `/api/v2/openai/system-prompts`

#### 查询参数

| 参数 | 类型 | 说明 |
|------|------|------|
| name | string | 按名称模糊搜索 |
| role | string | 按角色精确搜索 |
| active | bool | 按激活状态筛选 |
| offset | int | 偏移量，默认0 |
| limit | int | 限制数量，默认20，最大100 |

#### 示例请求

```
GET /api/v2/openai/system-prompts?name=default&active=true&limit=10
```

#### 响应

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "name": "default",
        "role": "system",
        "content": "你是一个有用的AI助手，请用中文回答问题。",
        "tags": "[\"default\", \"chinese\"]",
        "is_active": true,
        "version": 1,
        "created_at": 1703123456789,
        "updated_at": 1703123456789
      }
    ],
    "total": 1,
    "offset": 0,
    "limit": 10
  }
}
```

### 3. 获取单个系统提示词

**GET** `/api/v2/openai/system-prompts/{id}`

#### 路径参数

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int64 | 系统提示词ID |

#### 响应

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "name": "default",
    "role": "system",
    "content": "你是一个有用的AI助手，请用中文回答问题。",
    "tags": "[\"default\", \"chinese\"]",
    "is_active": true,
    "version": 1,
    "created_at": 1703123456789,
    "updated_at": 1703123456789
  }
}
```

### 4. 更新系统提示词

**PUT** `/api/v2/openai/system-prompts/{id}`

#### 路径参数

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int64 | 系统提示词ID |

#### 请求体

```json
{
  "name": "default_updated",
  "content": "你是一个有用的AI助手，请用中文回答问题。请保持友好和专业的态度。",
  "is_active": true
}
```

#### 响应

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "name": "default_updated",
    "role": "system",
    "content": "你是一个有用的AI助手，请用中文回答问题。请保持友好和专业的态度。",
    "tags": "[\"default\", \"chinese\"]",
    "is_active": true,
    "version": 2,
    "created_at": 1703123456789,
    "updated_at": 1703123456790
  }
}
```

### 5. 删除系统提示词

**DELETE** `/api/v2/openai/system-prompts/{id}`

#### 路径参数

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int64 | 系统提示词ID |

#### 响应

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "删除成功"
  }
}
```

## 系统提示词自动注入

系统提示词会在以下情况下自动注入到AI对话中：

1. **默认提示词注入**：当会话中没有绑定的系统提示词时，会自动使用名为"default"且激活状态的系统提示词
2. **注入时机**：在每次调用 `/api/v2/openai/chat/send` 或 `/api/v2/openai/chat/stream` 时
3. **注入方式**：作为"system"角色的消息添加到对话上下文中
4. **统一错误处理**：所有API都使用项目的统一错误处理机制，确保响应格式一致

### 注入逻辑

```go
// 在OpenAI控制器中的注入逻辑
if sp, err := c.factory.SystemPrompt().GetActiveDefault(ctx); err == nil && sp != nil && sp.Content != "" {
    _ = c.openaiClient.GetContextManager().SaveMessage(sessionID, "system", sp.Content, map[string]interface{}{
        "system_prompt_id":   sp.ID,
        "system_prompt_name": sp.Name,
    })
}
```

### 统一响应格式

所有API都遵循项目的统一响应格式：

#### 成功响应
```json
{
  "code": 200,
  "message": "success",
  "result": {
    // 具体数据
  }
}
```

#### 错误响应
```json
{
  "code": 400,
  "message": "错误信息描述"
}
```

## 使用示例

### 1. 创建不同类型的系统提示词

```bash
# 创建默认助手
curl -X POST "http://localhost:8080/api/v2/openai/system-prompts" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "default",
    "content": "你是一个有用的AI助手，请用中文回答问题。",
    "tags": "[\"default\", \"chinese\"]"
  }'

# 创建编程助手
curl -X POST "http://localhost:8080/api/v2/openai/system-prompts" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "programmer",
    "content": "你是一个专业的编程助手，擅长Go语言开发。请提供准确、详细的代码建议。",
    "tags": "[\"programming\", \"golang\"]"
  }'

# 创建创意写作助手
curl -X POST "http://localhost:8080/api/v2/openai/system-prompts" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "creative_writer",
    "content": "你是一个富有创意的写作助手，擅长创作故事、诗歌和散文。",
    "tags": "[\"writing\", \"creative\"]"
  }'
```

### 2. 测试系统提示词效果

```bash
# 发送消息，系统会自动注入默认提示词
curl -X POST "http://localhost:8080/api/v2/openai/chat/send" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "你好，请介绍一下自己",
    "user_id": "test_user"
  }'
```

### 3. 管理提示词

```bash
# 获取所有提示词
curl -X GET "http://localhost:8080/api/v2/openai/system-prompts"

# 按名称搜索
curl -X GET "http://localhost:8080/api/v2/openai/system-prompts?name=default"

# 更新提示词
curl -X PUT "http://localhost:8080/api/v2/openai/system-prompts/1" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "你是一个有用的AI助手，请用中文回答问题。请保持友好和专业的态度。"
  }'

# 删除提示词
curl -X DELETE "http://localhost:8080/api/v2/openai/system-prompts/1"
```

## 错误处理

### 常见错误码

| 错误码 | 说明 | 解决方案 |
|--------|------|----------|
| 400 | 请求参数错误 | 检查请求体格式和必填字段 |
| 404 | 资源不存在 | 检查ID是否正确 |
| 409 | 名称冲突 | 系统提示词名称必须唯一 |
| 500 | 服务器内部错误 | 查看服务器日志 |

### 错误响应格式

```json
{
  "code": 400,
  "message": "请求参数错误",
  "error": "参数不合法"
}
```

## 最佳实践

1. **命名规范**：使用有意义的名称，如"default"、"programmer"、"creative_writer"
2. **标签使用**：合理使用标签进行分类，便于管理和搜索
3. **版本控制**：系统会自动维护版本号，便于追踪变更
4. **激活状态**：合理使用is_active字段控制提示词的使用
5. **内容长度**：建议系统提示词内容控制在合理长度内，避免影响对话效果

## 注意事项

1. 系统提示词会在每次对话时注入，请确保内容简洁有效
2. 删除系统提示词前请确认没有正在使用的会话
3. 更新系统提示词会影响所有使用该提示词的对话
4. 建议保留一个名为"default"的激活提示词作为默认选项
