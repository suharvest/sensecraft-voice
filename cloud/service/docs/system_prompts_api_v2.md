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
  "is_default": true,
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
| is_default | bool | 否 | 是否为默认提示词，全局互斥，默认false |
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
  "is_active": true,
  "is_default": true
}
```

**注意**: 状态字段支持两种格式：
- 小写：`"is_active": true/false`, `"is_default": true/false`
- 大写：`"IsActive": true/false`, `"IsDefault": true/false`

如果不提供这些字段，默认为 `is_active: true`, `is_default: false`。

**重要**: `is_default` 字段具有全局互斥性，当设置某个提示词为默认时，会自动将其他所有提示词的 `is_default` 设为 `false`。

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
    "is_default": true,
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

### 3. 按名称模糊搜索系统提示词

**GET** `/api/v2/openai/system-prompts/search`

#### 查询参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 搜索关键词，支持模糊匹配 |
| limit | int | 否 | 限制数量，默认20，最大100 |

#### 示例请求

```
GET /api/v2/openai/system-prompts/search?name=会议&limit=10
```

#### 响应

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 2,
        "name": "会议纪要助手",
        "role": "system",
        "content": "你叫Seeed 智能会议助手，负责总结会议纪要",
        "tags": "[\"Seeed\",\"会议\"]",
        "is_active": true,
        "version": 1,
        "created_at": 1703123456789,
        "updated_at": 1703123456789
      },
      {
        "id": 3,
        "name": "会议记录员",
        "role": "system",
        "content": "你是一个专业的会议记录员",
        "tags": "[\"会议\",\"记录\"]",
        "is_active": false,
        "version": 1,
        "created_at": 1703123456790,
        "updated_at": 1703123456790
      }
    ],
    "count": 2,
    "name": "会议",
    "limit": 10
  }
}
```

### 4. 获取单个系统提示词

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
    "is_default": true,
    "version": 1,
    "created_at": 1703123456789,
    "updated_at": 1703123456789
  }
}
```

### 5. 更新系统提示词

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

**注意**: `is_active` 字段支持两种格式：
- 小写：`"is_active": true/false`
- 大写：`"IsActive": true/false`

如果提供了该字段，会更新激活状态；如果不提供，则不会影响现有的激活状态。

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

### 6. 更新系统提示词激活状态

**PATCH** `/api/v2/openai/system-prompts/{id}/status`

#### 路径参数

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int64 | 系统提示词ID |

#### 请求体

**激活系统提示词**:
```json
{
  "is_active": true
}
```

**停用系统提示词**:
```json
{
  "is_active": false
}
```

#### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| is_active | bool | 是 | 激活状态，true为激活，false为停用 |

**注意**: 该字段为必填字段，必须明确指定 true 或 false，不能省略。

#### 响应

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "系统提示词状态已更新为激活",
    "id": 1,
    "is_active": true
  }
}
```

#### 停用响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "系统提示词状态已更新为停用",
    "id": 1,
    "is_active": false
  }
}
```

### 7. 设置系统提示词为默认

**PATCH** `/api/v2/openai/system-prompts/{id}/default`

#### 路径参数

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int64 | 系统提示词ID |

#### 功能说明

将指定的系统提示词设置为默认提示词。此操作具有全局互斥性：
- 设置成功后，该提示词将成为唯一的默认提示词
- 其他所有提示词的 `is_default` 字段将自动设为 `false`
- 只有激活状态（`is_active = true`）的提示词才能被设为默认

#### 响应

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "设置成功"
  }
}
```

#### 错误情况

- **400**: 参数不合法（ID必须大于0）
- **404**: 系统提示词不存在
- **400**: 系统提示词未激活，无法设为默认

### 8. 删除系统提示词

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

### 8. 批量删除系统提示词

**DELETE** `/api/v2/openai/system-prompts`

#### 请求体

```json
{
  "ids": [1, 2, 3, 4, 5]
}
```

#### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| ids | array | 是 | 要删除的系统提示词ID列表，最多100个 |

#### 响应

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "成功删除 5 个系统提示词",
    "count": 5
  }
}
```

#### 错误响应

```json
{
  "code": 400,
  "message": "ID列表不能为空"
}
```

```json
{
  "code": 400,
  "message": "批量删除数量不能超过100个"
}
```

## 系统提示词自动注入

系统提示词会在以下情况下自动注入到AI对话中：

1. **默认提示词注入**：当会话中没有绑定的系统提示词时，会自动使用 `is_default = true` 且 `is_active = true` 的系统提示词
2. **注入时机**：在每次调用 `/api/v2/openai/chat/send` 或 `/api/v2/openai/chat/stream` 时
3. **注入方式**：作为"system"角色的消息添加到对话上下文中
4. **统一错误处理**：所有API都使用项目的统一错误处理机制，确保响应格式一致

### 注入逻辑

```go
// 在OpenAI控制器中的注入逻辑
// 查找 is_default = true 且 is_active = true 的系统提示词
if sp, err := c.factory.SystemPrompt().GetActiveDefault(ctx); err == nil && sp != nil && sp.Content != "" {
    _ = c.openaiClient.GetContextManager().SaveMessage(sessionID, "system", sp.Content, map[string]interface{}{
        "system_prompt_id":   sp.ID,
        "system_prompt_name": sp.Name,
        "system_prompt_source": "default",
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
# 创建默认助手（使用小写字段）
curl -X POST "http://localhost:8080/api/v2/openai/system-prompts" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "default",
    "content": "你是一个有用的AI助手，请用中文回答问题。",
    "tags": "[\"default\", \"chinese\"]",
    "is_active": true,
    "is_default": true
  }'

# 创建停用的助手（使用大写字段）
curl -X POST "http://localhost:8080/api/v2/openai/system-prompts" \
  -H "Content-Type: application/json" \
  -d '{
    "Name": "会议纪要",
    "Role": "system",
    "Content": "你叫Seeed 智能会议助手，负责总结会议纪要",
    "Tags": "[\"Seeed\",\"会议\"]",
    "IsActive": false
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

# 按名称搜索（列表接口）
curl -X GET "http://localhost:8080/api/v2/openai/system-prompts?name=default"

# 按名称模糊搜索（专用搜索接口）
curl -X GET "http://localhost:8080/api/v2/openai/system-prompts/search?name=会议&limit=10"

# 设置某个提示词为默认（全局互斥）
curl -X PATCH "http://localhost:8080/api/v2/openai/system-prompts/2/default"

# 更新提示词，同时设置为默认
curl -X PUT "http://localhost:8080/api/v2/openai/system-prompts/2" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "你是一个专业的编程助手，擅长多种编程语言。",
    "is_default": true
  }'

# 更新提示词
curl -X PUT "http://localhost:8080/api/v2/openai/system-prompts/1" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "你是一个有用的AI助手，请用中文回答问题。请保持友好和专业的态度。"
  }'

# 激活提示词
curl -X PATCH "http://localhost:8080/api/v2/openai/system-prompts/1/status" \
  -H "Content-Type: application/json" \
  -d '{
    "is_active": true
  }'

# 停用提示词
curl -X PATCH "http://localhost:8080/api/v2/openai/system-prompts/1/status" \
  -H "Content-Type: application/json" \
  -d '{
    "is_active": false
  }'

# 删除提示词
curl -X DELETE "http://localhost:8080/api/v2/openai/system-prompts/1"

# 批量删除提示词
curl -X DELETE "http://localhost:8080/api/v2/openai/system-prompts" \
  -H "Content-Type: application/json" \
  -d '{
    "ids": [1, 2, 3, 4, 5]
  }'
```

## 错误处理

### 常见错误码

| 错误码 | 说明 | 解决方案 |
|--------|------|----------|
| 400 | 请求参数错误 | 检查请求体格式和必填字段 |
| 400 | 名称冲突 | 系统提示词名称必须唯一，请使用其他名称 |
| 404 | 资源不存在 | 检查ID是否正确 |
| 500 | 服务器内部错误 | 查看服务器日志 |

### 错误响应格式

#### 参数错误
```json
{
  "code": 400,
  "message": "请求参数错误"
}
```

#### 名称冲突错误
```json
{
  "code": 400,
  "message": "系统提示词名称 '测试' 已存在"
}
```

#### 资源不存在错误
```json
{
  "code": 400,
  "message": "资源不存在"
}
```

## 最佳实践

1. **命名规范**：使用有意义的名称，如"default"、"programmer"、"creative_writer"
2. **名称唯一性**：系统提示词名称必须全局唯一，创建和更新时会自动检查重复
3. **标签使用**：合理使用标签进行分类，便于管理和搜索
4. **版本控制**：系统会自动维护版本号，便于追踪变更
5. **激活状态**：合理使用is_active字段控制提示词的使用
6. **内容长度**：建议系统提示词内容控制在合理长度内，避免影响对话效果

## 注意事项

1. 系统提示词会在每次对话时注入，请确保内容简洁有效
2. 删除系统提示词前请确认没有正在使用的会话
3. 更新系统提示词会影响所有使用该提示词的对话
4. 建议保留一个名为"default"的激活提示词作为默认选项
