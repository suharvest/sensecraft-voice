# 关键词匹配 API 接口文档

## 概述

关键词匹配模块提供录音文本的关键词匹配功能，支持精确匹配和近义词匹配，并将匹配结果存储到数据库中。

## 基础信息

- **基础路径**: `/api/v1/recordings`
- **认证方式**: Bearer Token
- **数据格式**: JSON
- **字符编码**: UTF-8

## 数据模型

### KeywordMatch 关键词匹配记录

```typescript
interface KeywordMatch {
  id: number;              // 匹配记录ID，主键
  recording_id: number;    // 录音ID
  mac_address: string;     // MAC地址
  keyword_id: number;      // 关键词ID
  keyword: string;         // 关键词内容
  matched_text: string;    // 匹配到的具体文本
  match_type: string;      // 匹配类型：exact(精确) | synonym(近义词)
  confidence: number;      // 匹配置信度：1.0(精确) | 0.8(近义词)
  position: number;        // 在文本中的位置
  length: number;          // 匹配文本长度
  created_at: number;      // 创建时间戳
}
```

### 请求参数

#### 分页参数
```typescript
interface PaginationParams {
  offset?: number;         // 偏移量，默认0
  limit?: number;          // 每页数量，默认20，最大100
}
```

#### 关键词匹配查询参数
```typescript
interface KeywordMatchQueryParams extends PaginationParams {
  mac_address?: string;    // MAC地址筛选
  keyword_id?: number;     // 关键词ID筛选
  recording_id?: number;   // 录音ID筛选
}
```

## API 接口

### 1. 获取关键词匹配记录列表

**接口地址**: `GET /api/v1/recordings/keyword-matches`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 | 示例 |
|--------|------|------|------|------|
| offset | number | 否 | 偏移量，默认0 | 0 |
| limit | number | 否 | 每页数量，默认20，最大100 | 20 |
| mac_address | string | 否 | MAC地址筛选 | "aa:bb:cc:dd:ee:01" |
| keyword_id | number | 否 | 关键词ID筛选 | 1 |
| recording_id | number | 否 | 录音ID筛选 | 123 |

**请求示例**:
```http
GET /api/v1/recordings/keyword-matches?mac_address=aa:bb:cc:dd:ee:01&offset=0&limit=20
Authorization: Bearer your_token_here
```

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "result": {
    "total": 15,
    "items": [
      {
        "id": 1,
        "recording_id": 123,
        "mac_address": "aa:bb:cc:dd:ee:01",
        "keyword_id": 1,
        "keyword": "投诉",
        "matched_text": "投诉",
        "match_type": "exact",
        "confidence": 1.0,
        "position": 15,
        "length": 2,
        "created_at": 1703123456789
      },
      {
        "id": 2,
        "recording_id": 123,
        "mac_address": "aa:bb:cc:dd:ee:01",
        "keyword_id": 1,
        "keyword": "投诉",
        "matched_text": "不满",
        "match_type": "synonym",
        "confidence": 0.8,
        "position": 8,
        "length": 2,
        "created_at": 1703123456789
      }
    ],
    "limit": 20,
    "offset": 0
  }
}
```

### 2. 根据MAC地址获取匹配记录

**接口地址**: `GET /api/v1/recordings/keyword-matches`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 | 示例 |
|--------|------|------|------|------|
| mac_address | string | 是 | MAC地址 | "aa:bb:cc:dd:ee:01" |
| offset | number | 否 | 偏移量，默认0 | 0 |
| limit | number | 否 | 每页数量，默认20 | 20 |

**请求示例**:
```http
GET /api/v1/recordings/keyword-matches?mac_address=aa:bb:cc:dd:ee:01
Authorization: Bearer your_token_here
```

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "result": {
    "total": 8,
    "items": [
      {
        "id": 1,
        "recording_id": 123,
        "mac_address": "aa:bb:cc:dd:ee:01",
        "keyword_id": 1,
        "keyword": "投诉",
        "matched_text": "投诉",
        "match_type": "exact",
        "confidence": 1.0,
        "position": 15,
        "length": 2,
        "created_at": 1703123456789
      }
    ],
    "limit": 20,
    "offset": 0
  }
}
```

### 3. 根据关键词ID获取匹配记录

**接口地址**: `GET /api/v1/recordings/keyword-matches`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 | 示例 |
|--------|------|------|------|------|
| keyword_id | number | 是 | 关键词ID | 1 |
| offset | number | 否 | 偏移量，默认0 | 0 |
| limit | number | 否 | 每页数量，默认20 | 20 |

**请求示例**:
```http
GET /api/v1/recordings/keyword-matches?keyword_id=1
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
        "recording_id": 123,
        "mac_address": "aa:bb:cc:dd:ee:01",
        "keyword_id": 1,
        "keyword": "投诉",
        "matched_text": "投诉",
        "match_type": "exact",
        "confidence": 1.0,
        "position": 15,
        "length": 2,
        "created_at": 1703123456789
      },
      {
        "id": 3,
        "recording_id": 124,
        "mac_address": "aa:bb:cc:dd:ee:02",
        "keyword_id": 1,
        "keyword": "投诉",
        "matched_text": "抱怨",
        "match_type": "synonym",
        "confidence": 0.8,
        "position": 5,
        "length": 2,
        "created_at": 1703123456790
      }
    ],
    "limit": 20,
    "offset": 0
  }
}
```

### 4. 根据录音ID获取匹配记录

**接口地址**: `GET /api/v1/recordings/keyword-matches`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 | 示例 |
|--------|------|------|------|------|
| recording_id | number | 是 | 录音ID | 123 |
| offset | number | 否 | 偏移量，默认0 | 0 |
| limit | number | 否 | 每页数量，默认20 | 20 |

**请求示例**:
```http
GET /api/v1/recordings/keyword-matches?recording_id=123
Authorization: Bearer your_token_here
```

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "result": {
    "total": 3,
    "items": [
      {
        "id": 1,
        "recording_id": 123,
        "mac_address": "aa:bb:cc:dd:ee:01",
        "keyword_id": 1,
        "keyword": "投诉",
        "matched_text": "投诉",
        "match_type": "exact",
        "confidence": 1.0,
        "position": 15,
        "length": 2,
        "created_at": 1703123456789
      },
      {
        "id": 2,
        "recording_id": 123,
        "mac_address": "aa:bb:cc:dd:ee:01",
        "keyword_id": 1,
        "keyword": "投诉",
        "matched_text": "不满",
        "match_type": "synonym",
        "confidence": 0.8,
        "position": 8,
        "length": 2,
        "created_at": 1703123456789
      },
      {
        "id": 3,
        "recording_id": 123,
        "mac_address": "aa:bb:cc:dd:ee:01",
        "keyword_id": 2,
        "keyword": "满意",
        "matched_text": "很好",
        "match_type": "synonym",
        "confidence": 0.8,
        "position": 25,
        "length": 2,
        "created_at": 1703123456789
      }
    ],
    "limit": 20,
    "offset": 0
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
| 400 | 数据验证失败 | 检查数据格式和长度限制 |
| 400 | 服务器内部错误 | 联系技术支持 |

**注意**: 根据系统设计，所有错误都统一返回HTTP 200状态码，具体错误信息在响应的`code`和`message`字段中体现。

## 数据验证规则

### 分页参数
- **offset**: 必须 >= 0
- **limit**: 必须 >= 1 且 <= 100

### 查询参数
- **mac_address**: 格式为 "xx:xx:xx:xx:xx:xx"
- **keyword_id**: 必须为正整数
- **recording_id**: 必须为正整数

## 使用示例

### JavaScript/TypeScript 示例

```typescript
// 获取关键词匹配记录
const getKeywordMatches = async (params: KeywordMatchQueryParams) => {
  const response = await fetch('/api/v1/recordings/keyword-matches?' + new URLSearchParams(params), {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    }
  });
  return await response.json();
};

// 根据MAC地址查询匹配记录
const getMatchesByMac = async (macAddress: string, offset = 0, limit = 20) => {
  return await getKeywordMatches({
    mac_address: macAddress,
    offset,
    limit
  });
};

// 根据关键词ID查询匹配记录
const getMatchesByKeyword = async (keywordId: number, offset = 0, limit = 20) => {
  return await getKeywordMatches({
    keyword_id: keywordId,
    offset,
    limit
  });
};

// 根据录音ID查询匹配记录
const getMatchesByRecording = async (recordingId: number, offset = 0, limit = 20) => {
  return await getKeywordMatches({
    recording_id: recordingId,
    offset,
    limit
  });
};
```

### cURL 示例

```bash
# 获取所有匹配记录
curl -X GET "http://your-domain.com/api/v1/recordings/keyword-matches?offset=0&limit=20" \
  -H "Authorization: Bearer your_token_here"

# 根据MAC地址查询
curl -X GET "http://your-domain.com/api/v1/recordings/keyword-matches?mac_address=aa:bb:cc:dd:ee:01" \
  -H "Authorization: Bearer your_token_here"

# 根据关键词ID查询
curl -X GET "http://your-domain.com/api/v1/recordings/keyword-matches?keyword_id=1" \
  -H "Authorization: Bearer your_token_here"

# 根据录音ID查询
curl -X GET "http://your-domain.com/api/v1/recordings/keyword-matches?recording_id=123" \
  -H "Authorization: Bearer your_token_here"

# 组合查询
curl -X GET "http://your-domain.com/api/v1/recordings/keyword-matches?mac_address=aa:bb:cc:dd:ee:01&keyword_id=1&offset=0&limit=10" \
  -H "Authorization: Bearer your_token_here"
```

## 匹配类型说明

### exact (精确匹配)
- 直接匹配关键词本身
- 置信度: 1.0
- 示例: 关键词"投诉"匹配到文本中的"投诉"

### synonym (近义词匹配)
- 匹配关键词的近义词
- 置信度: 0.8
- 示例: 关键词"投诉"的近义词"抱怨"匹配到文本中的"抱怨"

## 性能说明

### 查询性能
- 使用复合索引 `(mac_address, keyword_id)` 提高查询效率
- 支持分页查询，避免大量数据加载
- 建议使用合适的查询条件缩小结果集

### 数据量建议
- 单次查询建议 limit <= 100
- 大量数据查询建议使用分页
- 定期清理旧的匹配记录

## 注意事项

1. **认证**: 所有接口都需要有效的Bearer Token
2. **分页**: 建议使用分页查询，避免一次性获取大量数据
3. **查询条件**: 建议至少指定一个查询条件（mac_address、keyword_id或recording_id）
4. **数据延迟**: 关键词匹配是异步处理，可能存在短暂延迟
5. **错误处理**: 请根据返回的错误码进行相应的错误处理

## 更新日志

- **v1.0.0** (2024-01-25): 初始版本，支持关键词匹配记录查询
- **v1.1.0** (2024-01-26): 添加多种查询条件支持
- **v1.2.0** (2024-01-27): 优化查询性能和分页支持
