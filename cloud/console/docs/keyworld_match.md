### 6) 关键词匹配记录查询

- 路径: `/api/v1/recordings/keyword-matches`
- 方法: `GET`
- 说明: 查询关键词匹配记录，支持多种查询条件和分页。返回包含录音详细信息的匹配记录。

请求参数（所有参数都是可选的）

- `offset`(int, 可选): 偏移量，默认 0
- `limit`(int, 可选): 每页数量，不传则查询所有记录
- `mac_address`(string, 可选): MAC地址，查询指定设备的匹配记录
- `keyword_id`(int64, 可选): 关键词ID，查询指定关键词的匹配记录
- `recording_id`(int64, 可选): 录音ID，查询指定录音的匹配记录
- `start_time`(int64, 可选): 开始时间戳（毫秒），时间范围查询
- `end_time`(int64, 可选): 结束时间戳（毫秒），时间范围查询

请求示例

```bash
# 1. 查询所有匹配记录
curl -X GET 'http://127.0.0.1:3008/api/v1/recordings/keyword-matches' \
  -H 'Content-Type: application/json'

# 2. 分页查询
curl -X GET 'http://127.0.0.1:3008/api/v1/recordings/keyword-matches?offset=0&limit=10' \
  -H 'Content-Type: application/json'

# 3. 根据关键词ID查询
curl -X GET 'http://127.0.0.1:3008/api/v1/recordings/keyword-matches?keyword_id=20220803' \
  -H 'Content-Type: application/json'

# 4. 根据MAC地址查询
curl -X GET 'http://127.0.0.1:3008/api/v1/recordings/keyword-matches?mac_address=d8:3a:dd:54:c2:5d' \
  -H 'Content-Type: application/json'

# 5. 根据录音ID查询
curl -X GET 'http://127.0.0.1:3008/api/v1/recordings/keyword-matches?recording_id=65927' \
  -H 'Content-Type: application/json'

# 6. 时间范围查询
curl -X GET 'http://127.0.0.1:3008/api/v1/recordings/keyword-matches?start_time=1757588400000&end_time=1757588500000' \
  -H 'Content-Type: application/json'

# 7. 组合查询：时间范围 + 关键词ID
curl -X GET 'http://127.0.0.1:3008/api/v1/recordings/keyword-matches?start_time=1757588400000&end_time=1757588500000&keyword_id=20220803' \
  -H 'Content-Type: application/json'

# 8. 组合查询：时间范围 + MAC地址
curl -X GET 'http://127.0.0.1:3008/api/v1/recordings/keyword-matches?start_time=1757588400000&end_time=1757588500000&mac_address=d8:3a:dd:54:c2:5d' \
  -H 'Content-Type: application/json'

# 9. 组合查询：所有条件
curl -X GET 'http://127.0.0.1:3008/api/v1/recordings/keyword-matches?start_time=1757588400000&end_time=1757588500000&keyword_id=20220803&mac_address=d8:3a:dd:54:c2:5d' \
  -H 'Content-Type: application/json'
```

响应示例

```json
{
  "code": 0,
  "message": "success",
  "result": {
    "total": 682,
    "items": [
      {
        "id": 20221482,
        "recording_id": 65927,
        "mac_address": "d8:3a:dd:54:c2:5d",
        "keyword_id": 20220803,
        "keyword": "测试",
        "matched_text": "我",
        "match_type": "synonym",
        "confidence": 0.8,
        "position": 0,
        "length": 3,
        "created_at": 1757588484090,
        "session_id": "f2ccc0de24f28cc04b7394237c013e6e",
        "audio_id": "rri8tl",
        "speaker_id": "speaker_06",
        "speaker_name": "用户_06",
        "text": "我要",
        "device_time": 1757588484089,
        "status": 1
      }
    ],
    "limit": 0,
    "offset": 0
  }
}
```

字段说明

**请求参数**
- `offset`: 分页偏移量，从0开始
- `limit`: 每页记录数，不传则返回所有记录
- `mac_address`: 设备MAC地址，支持精确匹配
- `keyword_id`: 关键词ID，支持精确匹配
- `recording_id`: 录音记录ID，支持精确匹配
- `start_time`: 开始时间戳（毫秒），与end_time配合使用
- `end_time`: 结束时间戳（毫秒），与start_time配合使用

**响应字段**
- `total`: 总记录数
- `items`: 匹配记录列表
- `limit`: 每页记录数（0表示无限制）
- `offset`: 分页偏移量

**匹配记录字段**
- `id`: 匹配记录ID
- `recording_id`: 关联的录音记录ID
- `mac_address`: 设备MAC地址
- `keyword_id`: 关键词ID
- `keyword`: 关键词内容
- `matched_text`: 匹配到的具体文本
- `match_type`: 匹配类型（exact=精确匹配，synonym=近义词匹配）
- `confidence`: 匹配置信度（0-1）
- `position`: 在文本中的位置
- `length`: 匹配文本长度
- `created_at`: 匹配记录创建时间（毫秒时间戳）

**录音信息字段**（通过JOIN获取）
- `session_id`: 会话ID
- `audio_id`: 音频ID
- `speaker_id`: 说话人ID
- `speaker_name`: 说话人姓名
- `text`: 录音文本内容
- `device_time`: 设备时间戳（毫秒）
- `status`: 录音状态

查询优先级

1. **录音ID查询**: 如果传了 `recording_id`，优先使用录音ID查询
2. **组合查询**: 如果有其他条件（时间范围、关键词ID、MAC地址），使用组合查询
3. **全量查询**: 如果没有任何条件，查询所有记录

注意事项

- 所有参数都是可选的，可以任意组合使用
- 时间戳使用毫秒格式
- MAC地址支持大小写，服务端自动转换为小写
- 查询结果按创建时间倒序排列
- 所有查询都会返回录音的详细信息（通过JOIN recording表）

---