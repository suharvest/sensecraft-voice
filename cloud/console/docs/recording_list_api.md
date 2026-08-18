# Recording列表查询API文档

## 概述
本文档描述了recording模块新增的列表查询接口，支持分页查询和按mac地址过滤。

## API接口

### 查询录音记录列表
```http
GET /api/v1/recordings?offset=0&limit=20&mac_address=aa:bb:cc:dd:ee:ff
```

## 请求参数

### 查询参数 (Query Parameters)

| 参数名 | 类型 | 必填 | 默认值 | 说明 |
|--------|------|------|--------|------|
| `offset` | int | 否 | 0 | 偏移量，用于分页 |
| `limit` | int | 否 | 20 | 每页记录数，最大100 |
| `mac_address` | string | 否 | - | MAC地址过滤，支持模糊匹配 |

### 参数说明

- **offset**: 分页偏移量，从0开始
- **limit**: 每页记录数，范围1-100，默认20
- **mac_address**: 设备MAC地址，用于过滤特定设备的录音记录

## 响应格式

### 成功响应
```json
{
  "code": 200,
  "result": {
    "total": 150,
    "items": [
      {
        "id": 123,
        "mac_address": "aa:bb:cc:dd:ee:ff",
        "speaker_id": "speaker_05",
        "speaker_name": "用户_05",
        "text": "中学之前一直是在",
        "status": 1,
        "created_at": 1756388344546,
        "device_time": 1756388344540
      },
      {
        "id": 122,
        "mac_address": "aa:bb:cc:dd:ee:ff",
        "speaker_id": "speaker_03",
        "speaker_name": "用户_03",
        "text": "今天天气很好",
        "status": 1,
        "created_at": 1756388344000,
        "device_time": 1756388343990
      }
    ]
  }
}
```

### 响应字段说明

| 字段名 | 类型 | 说明 |
|--------|------|------|
| `total` | int64 | 总记录数 |
| `items` | array | 录音记录列表 |
| `items[].id` | int64 | 记录ID |
| `items[].mac_address` | string | 设备MAC地址 |
| `items[].speaker_id` | string | 说话人ID |
| `items[].speaker_name` | string | 说话人姓名 |
| `items[].text` | string | 录音文本内容 |
| `items[].status` | int8 | 状态（0=中间结果，1=最终结果） |
| `items[].created_at` | int64 | 创建时间（毫秒） |
| `items[].device_time` | int64 | 设备时间（毫秒） |

## 使用示例

### 1. 查询所有录音记录（分页）
```bash
curl "http://localhost:8080/api/v1/recordings?offset=0&limit=10"
```

### 2. 查询特定设备的录音记录
```bash
curl "http://localhost:8080/api/v1/recordings?mac_address=aa:bb:cc:dd:ee:ff"
```

### 3. 查询特定设备的分页录音记录
```bash
curl "http://localhost:8080/api/v1/recordings?mac_address=aa:bb:cc:dd:ee:ff&offset=20&limit=15"
```

### 4. 查询第二页数据
```bash
curl "http://localhost:8080/api/v1/recordings?offset=20&limit=20"
```

## 业务规则

### 1. 分页规则
- offset从0开始计数
- limit范围：1-100，默认20
- 按创建时间倒序排列（最新的在前）

### 2. 过滤规则
- mac_address支持精确匹配
- mac_address自动转换为小写
- 如果不提供mac_address，则查询所有设备的记录

### 3. 排序规则
- 默认按`created_at`字段倒序排列
- 最新创建的记录排在前面

## 性能考虑

### 1. 索引优化
- `mac_address`字段已建立索引
- `created_at`字段已建立索引
- 支持高效的过滤和排序查询

### 2. 分页建议
- 建议limit不要超过100，避免单次查询数据过多
- 大数据量场景下，建议使用offset+limit进行分页

## 错误处理

### 常见错误码
- `400`: 参数验证失败（如limit超出范围）
- `500`: 服务器内部错误

### 错误响应格式
```json
{
  "code": 400,
  "message": "参数验证失败"
}
```

## 完整示例

### 前端分页实现示例
```javascript
// 获取录音记录列表
async function getRecordings(page = 1, pageSize = 20, macAddress = '') {
  const offset = (page - 1) * pageSize;
  const params = new URLSearchParams({
    offset: offset.toString(),
    limit: pageSize.toString()
  });
  
  if (macAddress) {
    params.append('mac_address', macAddress);
  }
  
  const response = await fetch(`/api/v1/recordings?${params}`);
  const data = await response.json();
  
  if (data.code === 200) {
    return {
      total: data.result.total,
      items: data.result.items,
      currentPage: page,
      totalPages: Math.ceil(data.result.total / pageSize)
    };
  }
  
  throw new Error(data.message);
}

// 使用示例
try {
  const result = await getRecordings(1, 20, 'aa:bb:cc:dd:ee:ff');
  console.log(`总记录数: ${result.total}`);
  console.log(`当前页: ${result.currentPage}/${result.totalPages}`);
  console.log(`记录列表:`, result.items);
} catch (error) {
  console.error('获取录音记录失败:', error.message);
}
```

## 总结

新增的recording列表查询接口提供了：

✅ **分页支持**: 支持offset/limit分页查询
✅ **MAC过滤**: 支持按设备MAC地址过滤
✅ **性能优化**: 利用现有索引，查询效率高
✅ **灵活查询**: 支持组合查询条件
✅ **标准响应**: 统一的响应格式和错误处理

现在您可以方便地查询和管理录音记录了！
