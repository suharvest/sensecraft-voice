# 关键词匹配 API 快速参考

## 接口概览

| 方法 | 路径 | 功能 | 说明 |
|------|------|------|------|
| GET | `/api/v1/recordings/keyword-matches` | 查询关键词匹配记录 | 支持多种查询条件 |

## 请求参数

| 参数名 | 类型 | 必填 | 说明 | 示例 |
|--------|------|------|------|------|
| mac_address | string | 否 | MAC地址筛选 | "aa:bb:cc:dd:ee:01" |
| keyword_id | number | 否 | 关键词ID筛选 | 1 |
| recording_id | number | 否 | 录音ID筛选 | 123 |
| offset | number | 否 | 偏移量，默认0 | 0 |
| limit | number | 否 | 每页数量，默认20，最大100 | 20 |

## 响应格式

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
      }
    ],
    "limit": 20,
    "offset": 0
  }
}
```

## 快速示例

### 1. 查询所有匹配记录
```bash
curl -X GET "http://localhost:3008/api/v1/recordings/keyword-matches" \
  -H "Authorization: Bearer your_token"
```

### 2. 根据MAC地址查询
```bash
curl -X GET "http://localhost:8080/api/v1/recordings/keyword-matches?mac_address=aa:bb:cc:dd:ee:01" \
  -H "Authorization: Bearer your_token"
```

### 3. 根据关键词ID查询
```bash
curl -X GET "http://localhost:8080/api/v1/recordings/keyword-matches?keyword_id=1" \
  -H "Authorization: Bearer your_token"
```

### 4. 根据录音ID查询
```bash
curl -X GET "http://localhost:8080/api/v1/recordings/keyword-matches?recording_id=123" \
  -H "Authorization: Bearer your_token"
```

### 5. 分页查询
```bash
curl -X GET "http://localhost:8080/api/v1/recordings/keyword-matches?offset=0&limit=10" \
  -H "Authorization: Bearer your_token"
```

### 6. 组合查询
```bash
curl -X GET "http://localhost:8080/api/v1/recordings/keyword-matches?mac_address=aa:bb:cc:dd:ee:01&keyword_id=1&offset=0&limit=10" \
  -H "Authorization: Bearer your_token"
```

## 匹配类型

- **exact**: 精确匹配，置信度 1.0
- **synonym**: 近义词匹配，置信度 0.8

## 错误码

- **200**: 成功
- **400**: 参数错误或服务器错误

## 注意事项

1. 至少需要指定一个查询条件
2. 建议使用分页查询
3. 数据可能存在异步处理延迟
