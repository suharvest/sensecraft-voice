# 关键词管理 API 接口文档

## 概述

关键词管理模块提供关键词的增删改查功能，支持关键词、近义词和颜色标记的管理。

## 基础信息

- **基础路径**: `/api/v1/keywords`
- **认证方式**: Bearer Token
- **数据格式**: JSON
- **字符编码**: UTF-8

## 数据模型

### Keyword 关键词对象

```typescript
interface Keyword {
  id: number;           // 关键词ID，主键
  keyword: string;      // 关键词内容
  synonyms: string;     // 近义词，多个用逗号分隔
  mark_color: string;   // 标记颜色，十六进制格式（如：#ff4d4f）
  created_at: number;   // 创建时间戳
  updated_at: number;   // 更新时间戳
}
```

### 请求参数

#### 分页参数
```typescript
interface PaginationParams {
  offset?: number;      // 偏移量，默认0
  limit?: number;       // 每页数量，默认10，最大100
}
```

#### 查询参数
```typescript
interface KeywordQueryParams extends PaginationParams {
  keyword?: string;     // 关键词搜索
  mark_color?: string;  // 颜色筛选
}
```

#### 创建请求
```typescript
interface CreateKeywordRequest {
  keyword: string;      // 关键词，必填，1-50字符
  synonyms: string;     // 近义词，必填，多个用逗号分隔
  mark_color: string;   // 标记颜色，必填，十六进制格式
}
```

#### 更新请求
```typescript
interface UpdateKeywordRequest {
  keyword?: string;     // 关键词，可选
  synonyms?: string;    // 近义词，可选
  mark_color?: string;  // 标记颜色，可选
}
```

## API 接口

### 1. 获取关键词列表

**接口地址**: `GET /api/v1/keywords`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 | 示例 |
|--------|------|------|------|------|
| offset | number | 否 | 偏移量，默认0 | 0 |
| limit | number | 否 | 每页数量，默认10，最大100 | 10 |
| keyword | string | 否 | 关键词搜索 | "投诉" |
| mark_color | string | 否 | 颜色筛选 | "#ff4d4f" |

**请求示例**:
```http
GET /api/v1/keywords?offset=0&limit=10&keyword=投诉
Authorization: Bearer your_token_here
```

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "result": {
    "total": 25,
    "items": [
      {
        "id": 1,
        "keyword": "投诉",
        "synonyms": "抱怨, 不满, 问题",
        "mark_color": "#ff4d4f",
        "created_at": 1703123456789,
        "updated_at": 1703123456789
      },
      {
        "id": 2,
        "keyword": "满意",
        "synonyms": "开心, 认可, 不错",
        "mark_color": "#52c41a",
        "created_at": 1703123456790,
        "updated_at": 1703123456790
      }
    ],
    "limit": 10,
    "offset": 0
  }
}
```

### 2. 获取单个关键词

**接口地址**: `GET /api/v1/keywords/:id`

**路径参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | number | 是 | 关键词ID |

**请求示例**:
```http
GET /api/v1/keywords/1
Authorization: Bearer your_token_here
```

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "result": {
    "id": 1,
    "keyword": "投诉",
    "synonyms": "抱怨, 不满, 问题",
    "mark_color": "#ff4d4f",
    "created_at": 1703123456789,
    "updated_at": 1703123456789
  }
}
```

### 3. 创建关键词

**接口地址**: `POST /api/v1/keywords`

**请求体**:
```json
{
  "keyword": "价格",
  "synonyms": "多少钱, 怎么卖, 费用",
  "mark_color": "#1890ff"
}
```

**请求示例**:
```http
POST /api/v1/keywords
Authorization: Bearer your_token_here
Content-Type: application/json

{
  "keyword": "价格",
  "synonyms": "多少钱, 怎么卖, 费用",
  "mark_color": "#1890ff"
}
```

**响应示例**:
```json
{
  "code": 200,
  "message": "创建成功",
  "result": {
    "id": 3,
    "keyword": "价格",
    "synonyms": "多少钱, 怎么卖, 费用",
    "mark_color": "#1890ff",
    "created_at": 1703123456791,
    "updated_at": 1703123456791
  }
}
```

### 4. 更新关键词

**接口地址**: `PUT /api/v1/keywords/:id`

**路径参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | number | 是 | 关键词ID |

**请求体**:
```json
{
  "keyword": "价格问题",
  "synonyms": "多少钱, 怎么卖, 费用, 贵",
  "mark_color": "#fa8c16"
}
```

**请求示例**:
```http
PUT /api/v1/keywords/3
Authorization: Bearer your_token_here
Content-Type: application/json

{
  "keyword": "价格问题",
  "synonyms": "多少钱, 怎么卖, 费用, 贵",
  "mark_color": "#fa8c16"
}
```

**响应示例**:
```json
{
  "code": 200,
  "message": "更新成功",
  "result": {
    "id": 3,
    "keyword": "价格问题",
    "synonyms": "多少钱, 怎么卖, 费用, 贵",
    "mark_color": "#fa8c16",
    "created_at": 1703123456791,
    "updated_at": 1703123456792
  }
}
```

### 5. 删除关键词

**接口地址**: `DELETE /api/v1/keywords/:id`

**路径参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | number | 是 | 关键词ID |

**请求示例**:
```http
DELETE /api/v1/keywords/3
Authorization: Bearer your_token_here
```

**响应示例**:
```json
{
  "code": 200,
  "message": "删除成功",
  "result": null
}
```

### 6. 批量删除关键词

**接口地址**: `DELETE /api/v1/keywords/batch`

**请求体**:
```json
{
  "ids": [1, 2, 3]
}
```

**请求示例**:
```http
DELETE /api/v1/keywords/batch
Authorization: Bearer your_token_here
Content-Type: application/json

{
  "ids": [1, 2, 3]
}
```

**响应示例**:
```json
{
  "code": 200,
  "message": "批量删除成功",
  "result": {
    "deleted_count": 3,
    "deleted_ids": [1, 2, 3]
  }
}
```

## 错误码说明

| 错误码 | 说明 | 解决方案 |
|--------|------|----------|
| 200 | 请求成功 | - |
| 400 | 请求参数错误 | 检查请求参数格式和必填字段 |
| 400 | 未授权 | 检查Token是否有效 |
| 400 | 权限不足 | 联系管理员获取权限 |
| 400 | 关键词不存在 | 检查关键词ID是否正确 |
| 400 | 关键词已存在 | 关键词名称重复，请修改 |
| 400 | 数据验证失败 | 检查数据格式和长度限制 |
| 400 | 服务器内部错误 | 联系技术支持 |

**注意**: 根据系统设计，所有错误都统一返回HTTP 200状态码，具体错误信息在响应的`code`和`message`字段中体现。

## 数据验证规则

### 关键词 (keyword)
- **必填**: 是
- **长度**: 1-50个字符
- **格式**: 字符串，不能为空
- **唯一性**: 在系统中必须唯一

### 近义词 (synonyms)
- **必填**: 是
- **长度**: 1-500个字符
- **格式**: 字符串，多个近义词用逗号分隔
- **示例**: "抱怨, 不满, 问题"

### 标记颜色 (mark_color)
- **必填**: 是
- **格式**: 十六进制颜色值
- **示例**: "#ff4d4f", "#52c41a", "#1890ff"
- **默认值**: "#ff4d4f" (红色)

## 使用示例

### JavaScript/TypeScript 示例

```typescript
// 获取关键词列表
const getKeywords = async (params: KeywordQueryParams) => {
  const response = await fetch('/api/v1/keywords?' + new URLSearchParams(params), {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    }
  });
  return await response.json();
};

// 创建关键词
const createKeyword = async (data: CreateKeywordRequest) => {
  const response = await fetch('/api/v1/keywords', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(data)
  });
  return await response.json();
};

// 更新关键词
const updateKeyword = async (id: number, data: UpdateKeywordRequest) => {
  const response = await fetch(`/api/v1/keywords/${id}`, {
    method: 'PUT',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(data)
  });
  return await response.json();
};

// 删除关键词
const deleteKeyword = async (id: number) => {
  const response = await fetch(`/api/v1/keywords/${id}`, {
    method: 'DELETE',
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });
  return await response.json();
};

// 批量删除关键词
const batchDeleteKeywords = async (ids: number[]) => {
  const response = await fetch('/api/v1/keywords/batch', {
    method: 'DELETE',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ ids })
  });
  return await response.json();
};
```

### cURL 示例

```bash
# 获取关键词列表
curl -X GET "http://your-domain.com/api/v1/keywords?offset=0&limit=10" \
  -H "Authorization: Bearer your_token_here"

# 创建关键词
curl -X POST "http://your-domain.com/api/v1/keywords" \
  -H "Authorization: Bearer your_token_here" \
  -H "Content-Type: application/json" \
  -d '{
    "keyword": "投诉",
    "synonyms": "抱怨, 不满, 问题",
    "mark_color": "#ff4d4f"
  }'

# 更新关键词
curl -X PUT "http://your-domain.com/api/v1/keywords/1" \
  -H "Authorization: Bearer your_token_here" \
  -H "Content-Type: application/json" \
  -d '{
    "keyword": "投诉问题",
    "synonyms": "抱怨, 不满, 问题, 意见",
    "mark_color": "#ff4d4f"
  }'

# 删除关键词
curl -X DELETE "http://your-domain.com/api/v1/keywords/1" \
  -H "Authorization: Bearer your_token_here"

# 批量删除关键词
curl -X DELETE "http://your-domain.com/api/v1/keywords/batch" \
  -H "Authorization: Bearer your_token_here" \
  -H "Content-Type: application/json" \
  -d '{"ids": [1, 2, 3]}'
```

## 注意事项

1. **认证**: 所有接口都需要有效的Bearer Token
2. **分页**: 建议使用分页查询，避免一次性获取大量数据
3. **颜色格式**: 颜色值必须使用十六进制格式，如 #ff4d4f
4. **近义词分隔**: 多个近义词使用逗号分隔，不要使用其他分隔符
5. **唯一性**: 关键词名称在系统中必须唯一
6. **默认颜色**: 新增关键词时默认使用红色 (#ff4d4f)
7. **批量操作**: 批量删除时，如果部分ID不存在，会返回成功删除的ID列表
8. **错误处理**: 请根据返回的错误码进行相应的错误处理

## 实现结构

根据项目代码风格，关键词管理模块需要实现以下文件结构：

```
pkg/db/model/keywords.go          # 数据模型定义
pkg/db/keywords.go               # 数据库操作接口和实现
pkg/controller/keywords/         # 控制器层
├── interface.go                 # 控制器接口定义
└── keywords.go                  # 控制器实现
api/server/router/keywords/      # 路由层
├── keywords.go                  # 路由定义
└── keywords_routes.go           # 路由处理函数
pkg/types/keywords.go            # 类型定义（可选）
```

### 数据模型示例

```go
// pkg/db/model/keywords.go
type Keyword struct {
    ID        int64  `gorm:"column:id;primaryKey;autoIncrement;not null" json:"id"`
    Keyword   string `gorm:"column:keyword;type:varchar(50);uniqueIndex;not null" json:"keyword"`
    Synonyms  string `gorm:"column:synonyms;type:varchar(500);not null" json:"synonyms"`
    MarkColor string `gorm:"column:mark_color;type:varchar(7);default:'#ff4d4f';not null" json:"mark_color"`
    CreatedAt int64  `gorm:"column:created_at;type:bigint;not null" json:"created_at"`
    UpdatedAt int64  `gorm:"column:updated_at;type:bigint;not null" json:"updated_at"`
}

func (k *Keyword) TableName() string { return "keywords" }
```

### 控制器接口示例

```go
// pkg/controller/keywords/interface.go
type Interface interface {
    Create(ctx context.Context, req *CreateKeywordRequest) (*model.Keyword, error)
    GetById(ctx context.Context, id int64) (*model.Keyword, error)
    List(ctx context.Context, req *ListKeywordsRequest) (*ListKeywordsResponse, error)
    Update(ctx context.Context, id int64, req *UpdateKeywordRequest) (*model.Keyword, error)
    Delete(ctx context.Context, id int64) error
    BatchDelete(ctx context.Context, ids []int64) (*BatchDeleteResponse, error)
}
```

### 路由定义示例

```go
// api/server/router/keywords/keywords.go
func (r *keywordsRouter) initRoutes(httpEngine *gin.Engine) {
    group := httpEngine.Group("/api/v1/keywords")
    {
        group.POST("", r.create)
        group.GET("", r.list)
        group.GET("/:id", r.getById)
        group.PUT("/:id", r.update)
        group.DELETE("/:id", r.delete)
        group.DELETE("/batch", r.batchDelete)
    }
}
```

## 更新日志

- **v1.0.0** (2024-01-01): 初始版本，支持基本的CRUD操作
- **v1.1.0** (2024-01-15): 添加批量删除功能
- **v1.2.0** (2024-01-20): 添加颜色筛选功能，支持调色盘选择
- **v1.3.0** (2024-01-25): 更新API路径为v1版本，统一错误处理机制
