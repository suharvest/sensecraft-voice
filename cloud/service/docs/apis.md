## 设备与录音接口说明

### 1) 设备注册

- 路径: `/api/v1/devices/register`
- 方法: `POST`
- 说明: 按 `mac_address` 幂等注册/更新设备信息；资源占用为最新一次上报值。

请求示例

```json
{
  "mac_address": "aa:bb:cc:dd:ee:ff",
  "name": "客厅语音网关",
  "ip_address": "192.168.1.100",
  "version": "v1.2.3",
  "cpu_usage_percent": 23.45,
  "memory_used_bytes": 536870912,
  "memory_total_bytes": 17179869184,
  "disk_used_bytes": 2147483648,
  "disk_total_bytes": 999995129856,
  "swap_used_bytes": 0,
  "swap_total_bytes": 18253611008
}
```

响应示例（成功）

```json
{
  "code": 0,
  "message": "success",
  "result": {
    "id": 123,
    "mac_address": "aa:bb:cc:dd:ee:ff",
    "name": "客厅语音网关",
    "ip_address": "192.168.1.100",
    "version": "v1.2.3",
    "cpu_usage_percent": 23.45,
    "memory_used_bytes": 536870912,
    "memory_total_bytes": 17179869184,
    "disk_used_bytes": 2147483648,
    "disk_total_bytes": 999995129856,
    "swap_used_bytes": 0,
    "swap_total_bytes": 18253611008
  }
}
```

字段说明

- `mac_address`(string, 必填): 设备 MAC，小写、冒号分隔
- `name`(string, 可选): 设备名称，服务端可设置或留空
- `ip_address`(string, 可选): 设备IP地址
- `cpu_usage_percent`(decimal): CPU 占用百分比 0-100.00
- `memory_used_bytes`/`memory_total_bytes`(int64): 内存使用/总字节数
- `disk_used_bytes`/`disk_total_bytes`(int64): 磁盘使用/总字节数
- `swap_used_bytes`/`swap_total_bytes`(int64): 交换空间使用/总字节数

错误码

- 400: 参数不合法（如缺少 `mac_address`）
- 500: 服务端内部错误

curl 示例

```bash
curl -X POST http://127.0.0.1:8080/api/v1/devices/register \
  -H 'Content-Type: application/json' \
  -d '{
    "mac_address":"aa:bb:cc:dd:ee:ff",
    "name":"客厅语音网关",
    "ip_address":"192.168.1.100",
    "version":"v1.2.3",
    "cpu_usage_percent":23.45,
    "memory_used_bytes":536870912,
    "memory_total_bytes":17179869184,
    "disk_used_bytes":2147483648,
    "disk_total_bytes":999995129856,
    "swap_used_bytes":0,
    "swap_total_bytes":18253611008
  }'
```

---

### 2) 设备列表查询

- 路径: `/api/v1/devices`
- 方法: `GET`
- 说明: 获取设备列表，支持分页查询。

请求参数

- `offset`(int, 可选): 偏移量，默认 0
- `limit`(int, 可选): 每页数量，默认 20，最大 100

请求示例

```bash
curl -X GET 'http://127.0.0.1:8081/api/v1/devices?offset=0&limit=20' \
  -H 'Content-Type: application/json'
```

响应示例

```json
{
  "code": 0,
  "message": "success",
  "result": {
    "total": 45,
    "items": [
      {
        "id": 123,
        "mac_address": "aa:bb:cc:dd:ee:ff",
        "name": "客厅语音网关",
        "ip_address": "192.168.1.100",
        "version": "v1.2.3",
        "cpu_usage_percent": 23.45,
        "memory_used_bytes": 536870912,
        "memory_total_bytes": 17179869184,
        "disk_used_bytes": 2147483648,
        "disk_total_bytes": 999995129856,
        "swap_used_bytes": 0,
        "swap_total_bytes": 18253611008,
        "created_at": 1756388344600,
        "updated_at": 1756388344600
      }
    ]
  }
}
```

---

### 3) 根据 MAC 查询设备

- 路径: `/api/v1/devices/:mac`
- 方法: `GET`
- 说明: 根据 MAC 地址查询单台设备信息。

路径参数

- `mac`(string, 必填): 设备 MAC 地址，支持大小写

请求示例

```bash
curl -X GET http://127.0.0.1:8081/api/v1/devices/aa:bb:cc:dd:ee:ff \
  -H 'Content-Type: application/json'
```

响应示例

```json
{
  "code": 0,
  "message": "success",
  "result": {
    "id": 123,
    "mac_address": "aa:bb:cc:dd:ee:ff",
    "name": "客厅语音网关",
    "ip_address": "192.168.1.100",
    "version": "v1.2.3",
    "cpu_usage_percent": 23.45,
    "memory_used_bytes": 536870912,
    "memory_total_bytes": 17179869184,
    "disk_used_bytes": 2147483648,
    "disk_total_bytes": 999995129856,
    "swap_used_bytes": 0,
    "swap_total_bytes": 18253611008,
    "created_at": 1756388344600,
    "updated_at": 1756388344600
  }
}
```

错误码

- 400: 参数不合法
- 404: 设备不存在
- 500: 服务端内部错误

---

### 4) 更新设备名称

- 路径: `/api/v1/devices/:id/name`
- 方法: `PUT`
- 说明: 更新指定设备的名称。

路径参数

- `id`(int64, 必填): 设备ID

请求体

```json
{
  "name": "新的设备名称"
}
```

字段说明

- `name`(string, 必填): 新的设备名称

请求示例

```bash
curl -X PUT http://127.0.0.1:8081/api/v1/devices/123/name \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "客厅智能音箱"
  }'
```

响应示例（成功）

```json
{
  "code": 0,
  "message": "success",
  "result": {
    "id": 123,
    "mac_address": "aa:bb:cc:dd:ee:ff",
    "name": "客厅智能音箱",
    "ip_address": "192.168.1.100",
    "version": "v1.2.3",
    "cpu_usage_percent": 23.45,
    "memory_used_bytes": 536870912,
    "memory_total_bytes": 17179869184,
    "disk_used_bytes": 2147483648,
    "disk_total_bytes": 999995129856,
    "swap_used_bytes": 0,
    "swap_total_bytes": 18253611008,
    "location_id": 0,
    "store_id": 0,
    "created_at": 1756388344600,
    "updated_at": 1756388344600
  }
}
```

错误码

- 400: 参数不合法（如设备ID无效或缺少名称字段）
- 404: 设备不存在
- 500: 服务端内部错误

注意事项

- 设备ID必须大于0
- 名称字段为必填项
- 更新成功后会自动更新 `updated_at` 时间戳
- 返回完整的设备信息

---

### 5) 录音记录列表

- 路径: `/api/v1/recordings`
- 方法: `GET`
- 说明: 获取录音记录列表，支持多种查询条件和分页。

请求参数（所有参数都是可选的）

- `offset`(int, 可选): 偏移量，默认 0
- `limit`(int, 可选): 每页数量，默认 50，最大 100
- `start_time`(int64, 可选): 开始时间戳（毫秒），基于创建时间过滤
- `end_time`(int64, 可选): 结束时间戳（毫秒），基于创建时间过滤
- `device_start_time`(int64, 可选): 设备开始时间戳（毫秒），基于设备时间过滤
- `device_end_time`(int64, 可选): 设备结束时间戳（毫秒），基于设备时间过滤
- `store_id`(int, 可选): 门店ID，查询指定门店下所有设备的录音记录
- `location_id`(int, 可选): 点位ID，查询指定点位下所有设备的录音记录
- `mac_address`(string[], 可选): 设备MAC地址，支持多个，查询指定设备的录音记录
- `type`(int, 可选): 类型参数，1: 查询音频记录并返回play_url, 0或不传: 不查询音频记录
- `status`(int8, 可选): 状态过滤，查询指定状态的录音记录

请求示例

```bash
# 1. 默认查询（前50条记录）
curl -X GET 'http://127.0.0.1:8081/api/v1/recordings' \
  -H 'Content-Type: application/json'

# 2. 分页查询
curl -X GET 'http://127.0.0.1:8081/api/v1/recordings?offset=0&limit=10' \
  -H 'Content-Type: application/json'

# 3. 时间范围查询（基于创建时间）
curl -X GET 'http://127.0.0.1:8081/api/v1/recordings?start_time=1756656000&end_time=1756742400' \
  -H 'Content-Type: application/json'

# 4. 设备时间范围查询
curl -X GET 'http://127.0.0.1:8081/api/v1/recordings?device_start_time=1758902400000&device_end_time=1758988800000' \
  -H 'Content-Type: application/json'

# 5. 门店查询
curl -X GET 'http://127.0.0.1:8081/api/v1/recordings?store_id=3' \
  -H 'Content-Type: application/json'

# 6. 点位查询
curl -X GET 'http://127.0.0.1:8081/api/v1/recordings?location_id=4' \
  -H 'Content-Type: application/json'

# 7. 单个MAC地址查询
curl -X GET 'http://127.0.0.1:8081/api/v1/recordings?mac_address=aa:bb:cc:dd:ee:ff' \
  -H 'Content-Type: application/json'

# 8. 多个MAC地址查询
curl -X GET 'http://127.0.0.1:8081/api/v1/recordings?mac_address=aa:bb:cc:dd:ee:ff&mac_address=aa:bb:cc:dd:ee:gg' \
  -H 'Content-Type: application/json'

# 9. 组合查询（包含设备时间范围）
curl -X GET 'http://127.0.0.1:8081/api/v1/recordings?offset=0&limit=20&device_start_time=1758902400000&device_end_time=1758988800000&type=1' \
  -H 'Content-Type: application/json'

# 10. 完整组合查询
curl -X GET 'http://127.0.0.1:8081/api/v1/recordings?offset=0&limit=10&start_time=1756656000&end_time=1756742400&store_id=3&location_id=4&mac_address=aa:bb:cc:dd:ee:ff' \
  -H 'Content-Type: application/json'
```

响应示例

```json
{
  "code": 0,
  "message": "success",
  "result": {
    "total": 156,
    "items": [
      {
        "id": 123,
        "mac_address": "aa:bb:cc:dd:ee:ff",
        "speaker_id": "speaker_05",
        "speaker_name": "用户_05",
        "text": "中学之前一直是在",
        "status": 1,
        "created_at": 1756388344600,
        "device_time": 1756388344546
      }
    ]
  }
}
```

字段说明

- `total`(int64): 符合条件的录音记录总数
- `items`(array): 录音记录列表
  - `id`(int64): 记录ID
  - `mac_address`(string): 设备MAC地址
  - `speaker_id`(string): 说话人ID
  - `speaker_name`(string): 说话人姓名
  - `text`(string): 识别文本内容
  - `status`(int8): 状态（0=中间结果，1=最终结果）
  - `created_at`(int64): 创建时间（毫秒时间戳）
  - `device_time`(int64): 设备时间（毫秒时间戳）

查询逻辑说明

- **门店查询**: 通过门店ID查询该门店下所有点位的所有设备的录音记录
- **点位查询**: 通过点位ID查询该点位下所有设备的录音记录
- **MAC地址查询**: 支持多个MAC地址，使用IN查询
- **时间范围查询**: 基于`created_at`字段进行过滤
- **组合查询**: 支持任意组合的查询条件，使用AND逻辑
- **关联查询**: 使用JOIN查询确保数据关联正确

注意事项

- 所有参数都是可选的，可以任意组合使用
- 默认分页参数：offset=0, limit=50
- 时间戳使用毫秒格式
- MAC地址支持大小写，服务端自动转换为小写
- 查询结果按创建时间倒序排列

---

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

### 7) 录音记录查询

- 路径: `/api/recordings/query`
- 方法: `POST`
- 说明: 根据设备ID（MAC地址）、时间范围和会话ID查询录音记录。查询完成后会异步调用Dify API处理录音内容，然后调用Seeed API发送"hello world"内容。
- 认证: 需要 Bearer Token，固定值为 `voiceai_secure_token_2025`

请求头

```
authorization: Bearer voiceai_secure_token_2025
Content-Type: application/json
```

请求体

```json
{
    "deviceId": "aa:bb:cc:dd:ee:ff",
    "startTimestamp": 1755587886000,
    "endTimestamp": 1755759910445,
    "sid": "test-session-1755759910445"
}
```

字段说明

- `deviceId`(string, 必填): 设备 MAC 地址
- `startTimestamp`(int64, 可选): 开始时间戳（毫秒）
- `endTimestamp`(int64, 可选): 结束时间戳（毫秒）
- `sid`(string, 可选): 会话ID（当前版本暂不支持过滤）

响应示例

```json
{
    "code": 0,
    "message": "success",
    "result": {
        "total": 156,
        "items": [
            {
                "id": 123,
                "mac_address": "aa:bb:cc:dd:ee:ff",
                "speaker_id": "speaker_05",
                "speaker_name": "用户_05",
                "text": "中学之前一直是在",
                "status": 1,
                "created_at": 1756388344600,
                "device_time": 1756388344546
            }
        ]
    }
}
```

curl 示例

```bash
curl --location 'http://127.0.0.1:8081/api/recordings/query' \
--header 'authorization: Bearer voiceai_secure_token_2025' \
--header 'Content-Type: application/json' \
--data '{
    "deviceId": "aa:bb:cc:dd:ee:ff",
    "startTimestamp": 1755587886000,
    "endTimestamp": 1755759910445,
    "sid": "test-session-1755759910445"
}'
```

错误码

- 401: 缺少或无效的认证 token
- 400: 请求参数不合法
- 500: 服务端内部错误

注意事项

- 时间戳使用毫秒格式
- 设备ID 为 MAC 地址，支持大小写（服务端自动转换为小写）
- 会话ID 过滤功能当前版本暂不支持，将在后续版本中实现
- **异步API调用流程**: 每次查询完成后会自动执行以下异步流程：
  1. 压缩录音数据（合并同一speaker的文本，简化speaker_id）
  2. 调用Dify API处理压缩后的数据（blocking模式）
  3. 调用Seeed API发送"hello world"内容到指定的会话ID
- **重试机制**: Dify API和Seeed API都支持失败重试，每隔1秒重试一次，最多重试3次
- **数据压缩**: 使用基础压缩策略，将speaker_id简化为S1、S2等格式，合并同一speaker的连续文本
- **Seeed API配置**: 使用MD5(固定秘钥+当天日期)规则生成Token，所有配置项都从配置文件读取


#### 异步API调用详情

当调用 `/api/recordings/query` 接口时，系统会执行以下流程：

1. **查询录音记录**: 根据请求参数查询数据库中的录音记录
2. **返回查询结果**: 立即返回查询到的录音记录数据
3. **异步处理流程**:
   a) **数据压缩**: 将录音数据压缩为简单格式
     - 简化speaker_id: `speaker_01` → `S1`, `speaker_02` → `S2`
     - 合并同一speaker的连续文本
     - 限制总长度在2000字符内
   
   b) **调用Dify API**: 使用blocking模式处理压缩后的数据
     - **API地址**: `https://dify.example.com/v1/chat-messages`
     - **认证**: `Bearer ***REMOVED***`
     - **模式**: `blocking`
     - **重试**: 失败时1秒间隔重试，最多3次
   
   c) **调用Seeed API**: 在Dify API成功后调用
     - **API地址**: `https://api.example.com/api/SF/ModifyRecordingMeetingContent`
     - **Token**: 使用MD5(固定秘钥+当天日期)规则生成
     - **会话ID**: 使用请求中的 `sid` 参数
     - **内容**: 使用Dify API返回的answer内容
     - **重试**: 失败时1秒间隔重试，最多3次

#### 数据压缩示例

**原始录音数据**:
```json
[
  {"speaker_id": "speaker_01", "text": "你好"},
  {"speaker_id": "speaker_01", "text": "今天天气不错"},
  {"speaker_id": "speaker_02", "text": "是的"},
  {"speaker_id": "speaker_01", "text": "我们去散步吧"}
]
```

**压缩后的数据**:
```
S1: 你好 今天天气不错; S2: 是的; S1: 我们去散步吧
```

**完整流程示例**:
```
1. 录音查询结果: [{"speaker_id": "speaker_01", "text": "你好"}, ...]
2. 数据压缩: "S1: 你好 今天天气不错; S2: 是的; S1: 我们去散步吧"
3. Dify API处理: 返回answer "根据录音内容分析，这是一个关于天气和散步的对话"
4. Seeed API调用: 发送answer内容到Seeed API
```

#### 配置说明

所有外部API的配置都通过 `config.yaml` 文件管理：

```yaml
# 聊天服务配置
chat:
  base_url: "https://dify.example.com"
  api_key: ***REMOVED***
  timeout: 60

# Seeed API配置
seeed:
  base_url: "https://api.example.com"
  secret_key: "***REMOVED***"
  timeout: 30
```

#### 错误处理

系统会自动处理以下错误情况：
- **401 Unauthorized**: 检查Dify API密钥是否正确
- **Timeout**: 网络超时，会自动重试
- **Connection refused**: Dify服务不可用，会自动重试
- **Seeed API错误**: 根据返回的错误码进行相应处理

#### 测试方法

可以使用提供的测试脚本验证功能：
```bash
./test_seeed_api_integration.sh
```

该脚本会：
1. 调用录音查询接口
2. 等待异步API调用完成
3. 显示接口调用结果

---

### 8) 录音记录存储（WebSocket）

- 路径: `WS /api/v1/recordings/stream`
- 说明: 客户端持续发送识别结果片段；服务端校验并入库。每条消息返回 `{ "ack": true|false }`。

入站消息示例

```json
{
  "sessionID": "2d435c1db3cc1aa2589c769bffb672dc",
  "mac_address": "aa:bb:cc:dd:ee:ff",
  "speaker": {
    "confidence": 0.73748225,
    "identified": true,
    "speaker_id": "speaker_05",
    "speaker_name": "用户_05"
  },
  "text": "中学之前一直是在",
  "textLength": 24,
  "timestamp": 1756388344546,
  "type": "final",
  "wordCount": 1,
  "status": 1
}
```

出站确认示例

```json
{ "ack": true }
```

字段与落库映射

- `mac_address` → `recordings.mac_address`
- `speaker.speaker_id` → `recordings.speaker_id`
- `speaker.speaker_name` → `recordings.speaker_name`
- `text` → `recordings.text`
- `status` → `recordings.status`（0=中间结果，1=最终结果）
- `device_time` ← `timestamp`（设备侧毫秒）
- `created_at`（服务端入库时间毫秒，服务端自动填充）

wscat 示例

```bash
wscat -c ws://127.0.0.1:8080/api/v1/recordings/stream
```

注意事项

- 连接鉴权（可选）：可在首帧携带 `sessionID` 做简单校验
- 心跳（可选）：建议客户端定期发送 ping/pong 保持连接
- 幂等（可选）：如端侧有 `msg_id`，可在后续扩展为去重键

---

### 9) 统计接口

- 路径: `/api/v1/stats/dashboard`
- 方法: `GET`
- 说明: 获取仪表板统计数据，包含各种总数和趋势分析，支持按门店筛选。

请求参数

- `store_id`(可选): 门店ID，用于筛选特定门店的数据。不传则统计所有门店。

响应示例

```json
{
  "code": 0,
  "message": "success",
  "result": {
    "total_records": 12580,
    "total_devices": 45,
    "total_stores": 12,
    "total_users": 156,
    "total_location": 28,
    "weekly_record_trend": [
      {
        "date": "2025-08-25",
        "count": 1205
      },
      {
        "date": "2025-08-26", 
        "count": 1356
      },
      {
        "date": "2025-08-27",
        "count": 1189
      },
      {
        "date": "2025-08-28",
        "count": 1423
      },
      {
        "date": "2025-08-29",
        "count": 987
      }
    ],
    "today_active_devices": 23,
    "today_hourly_distribution": [
      {
        "hour": 0,
        "count": 45
      },
      {
        "hour": 1,
        "count": 32
      },
      {
        "hour": 2,
        "count": 18
      }
    ],
    "today_keyword_triggers": 156,
    "today_keyword_matches": [
      {
        "keyword_id": 1,
        "keyword": "投诉",
        "mark_color": "#ff4d4f",
        "match_count": 45,
        "record_count": 38
      },
      {
        "keyword_id": 2,
        "keyword": "满意",
        "mark_color": "#52c41a",
        "match_count": 32,
        "record_count": 28
      }
    ]
  }
}
```

字段说明

- `total_records`(int64): 录音记录总数
- `total_devices`(int64): 设备总数
- `total_stores`(int64): 门店总数
- `total_users`(int64): 用户总数（通过录音记录中的 speaker_id 去重统计）
- `total_location`(int64): 点位总数（支持按门店筛选）
- `weekly_record_trend`(array): 本周录音记录增长趋势
  - `date`(string): 日期，格式 YYYY-MM-DD
  - `count`(int64): 当日录音记录数量
- `today_active_devices`(int64): 今日产生录音记录的设备数量
- `today_hourly_distribution`(array): 今日每小时采集分布（00:00-23:59）
  - `hour`(int): 小时，0-23
  - `count`(int64): 该小时的记录数量
- `today_keyword_triggers`(int64): 今日关键词触发总次数
- `today_keyword_matches`(array): 今日关键词匹配详情（包含所有关键词，匹配次数为0的也会显示）
  - `keyword_id`(int64): 关键词ID
  - `keyword`(string): 关键词内容
  - `mark_color`(string): 关键词标记颜色
  - `match_count`(int64): 匹配次数（0表示今日无匹配）
  - `record_count`(int64): 涉及到的记录数（0表示今日无匹配）

curl 示例

```bash
# 获取所有门店的统计数据
curl -X GET http://127.0.0.1:8081/api/v1/stats/dashboard \
  -H 'Content-Type: application/json'

# 获取指定门店的统计数据
curl -X GET "http://127.0.0.1:8081/api/v1/stats/dashboard?store_id=1" \
  -H 'Content-Type: application/json'
```

注意事项

- 门店筛选通过设备表关联实现，确保数据准确性
- 用户统计基于录音记录中的 speaker_id 去重
- 本周趋势从本周一开始计算，包含今天
- 今日活跃设备统计从今日 00:00:00 开始计算
- 今日每小时分布确保24小时都有数据，无数据的小时显示为0
- 关键词匹配统计包含所有关键词，匹配次数为0的也会显示
- 关键词匹配统计按匹配次数降序排列

---

## 聊天接口说明

### 10) 流式发送聊天消息（推荐）

- 路径: `/api/v1/chat/stream`
- 方法: `POST`
- 说明: 发送聊天消息到Dify API，返回真正的流式响应（Server-Sent Events）。

请求示例

```json
{
  "inputs": {},
  "query": "What are the specs of the iPhone 13 Pro Max?",
  "response_mode": "streaming",
  "conversation_id": "",
  "user": "abc-123",
  "files": [
    {
      "type": "image",
      "transfer_method": "remote_url",
      "url": "https://cloud.dify.ai/logo/logo-site.png"
    }
  ]
}
```

响应格式（Server-Sent Events）

```
data: {"event": "workflow_started", "task_id": "5ad4cb98-f0c7-4085-b384-88c403be6290", "workflow_run_id": "5ad498-f0c7-4085-b384-88be6290", "data": {"id": "5ad498-f0c7-4085-b384-88be6290", "workflow_id": "dfjasklfjdslag", "sequence_number": 1, "created_at": 1679586595}}

data: {"event": "message", "message_id": "5ad4cb98-f0c7-4085-b384-88c403be6290", "conversation_id": "45701982-8118-4bc5-8e9b-64562b4555f2", "answer": " I", "created_at": 1679586595}

data: {"event": "message", "message_id": "5ad4cb98-f0c7-4085-b384-88c403be6290", "conversation_id": "45701982-8118-4bc5-8e9b-64562b4555f2", "answer": "'m", "created_at": 1679586595}

data: {"event": "completed", "conversation_id": "45701982-8118-4bc5-8e9b-64562b4555f2"}
```

字段说明

- `inputs`(object, 可选): 输入参数
- `query`(string, 必填): 查询内容
- `response_mode`(string, 可选): 响应模式，默认 "streaming"
- `conversation_id`(string, 可选): 对话ID，为空时自动生成
- `user`(string, 必填): 用户标识
- `files`(array, 可选): 文件列表
  - `type`(string): 文件类型
  - `transfer_method`(string): 传输方式
  - `url`(string): 文件URL

响应事件类型

- `workflow_started`: 工作流开始

---

## OpenAI API v2 接口说明

OpenAI API v2 提供了完整的AI聊天功能，包括消息发送、流式响应、会话管理、上下文保持和系统提示词管理。

### 聊天接口

### 1) 发送聊天消息

- 路径: `/api/v2/openai/chat/send`
- 方法: `POST`
- 说明: 发送消息给OpenAI并获取回复，支持上下文管理和系统提示词自动注入

请求示例

```json
{
  "message": "你好，请介绍一下Go语言的特点",
  "user_id": "user_123",
  "session_id": "openai_1703123456789"
}
```

响应示例

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

### 2) 流式聊天消息

- 路径: `/api/v2/openai/chat/stream`
- 方法: `POST`
- 说明: 发送消息给OpenAI并获取流式回复（Server-Sent Events），支持系统提示词自动注入

请求示例

```json
{
  "message": "请详细解释一下微服务架构",
  "user_id": "user_123",
  "session_id": "openai_1703123456789"
}
```

响应格式（Server-Sent Events）

```
data: {"event": "message", "data": {"session_id": "openai_1703123456789", "content": "微服务架构是一种", "timestamp": 1703123456789}}

data: {"event": "message", "data": {"session_id": "openai_1703123456789", "content": "将应用程序构建为", "timestamp": 1703123456790}}

data: {"event": "completed", "data": {"session_id": "openai_1703123456789", "timestamp": 1703123456791}}
```

### 3) 获取聊天历史

- 路径: `/api/v2/openai/chat/history/{session_id}`
- 方法: `GET`
- 说明: 获取指定会话的聊天历史记录

请求示例

```bash
curl -X GET 'http://localhost:3008/api/v2/openai/chat/history/openai_1703123456789?limit=10' \
  -H 'Content-Type: application/json'
```

响应示例

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

### 4) 创建新会话

- 路径: `/api/v2/openai/chat/session`
- 方法: `POST`
- 说明: 创建一个新的聊天会话

请求示例

```json
{
  "user_id": "user_123"
}
```

响应示例

```json
{
  "code": 200,
  "result": {
    "session_id": "openai_1703123456789"
  }
}
```

### 5) 关闭会话

- 路径: `/api/v2/openai/chat/session/{session_id}`
- 方法: `DELETE`
- 说明: 关闭指定的聊天会话

响应示例

```json
{
  "code": 200,
  "result": {
    "message": "会话已关闭"
  }
}
```

字段说明

- `message`(string, 必填): 用户消息内容
- `user_id`(string, 必填): 用户ID
- `session_id`(string, 可选): 会话ID，不提供则自动创建
- `limit`(int, 可选): 历史记录数量限制，默认20，最大100

注意事项

- OpenAI API需要配置有效的API密钥
- 支持上下文管理，自动保持最近20条消息
- 流式接口使用Server-Sent Events格式
- 所有聊天数据都会保存到数据库
- `node_started`: 节点开始
- `node_finished`: 节点完成
- `workflow_finished`: 工作流完成
- `message`: 流式消息片段
- `message_end`: 消息完成
- `completed`: 流式响应完成

curl 示例

```bash
curl -X POST http://127.0.0.1:8081/api/v1/chat/stream \
  -H 'Content-Type: application/json' \
  -d '{
    "query": "What are the specs of the iPhone 13 Pro Max?",
    "user": "abc-123",
    "response_mode": "streaming"
  }'
```

---

### 11) 同步发送聊天消息

- 路径: `/api/v1/chat/send`
- 方法: `POST`
- 说明: 发送聊天消息到Dify API，同步处理完成后返回结果。

请求示例

```json
{
  "inputs": {},
  "query": "What are the specs of the iPhone 13 Pro Max?",
  "response_mode": "streaming",
  "conversation_id": "",
  "user": "abc-123",
  "files": [
    {
      "type": "image",
      "transfer_method": "remote_url",
      "url": "https://cloud.dify.ai/logo/logo-site.png"
    }
  ]
}
```

响应示例（成功）

```json
{
  "code": 0,
  "message": "success",
  "result": {
    "conversation_id": "session_1703123456789",
    "status": "sent"
  }
}
```

字段说明

- `inputs`(object, 可选): 输入参数
- `query`(string, 必填): 查询内容
- `response_mode`(string, 可选): 响应模式，默认 "streaming"
- `conversation_id`(string, 可选): 对话ID，为空时自动生成
- `user`(string, 必填): 用户标识
- `files`(array, 可选): 文件列表
  - `type`(string): 文件类型
  - `transfer_method`(string): 传输方式
  - `url`(string): 文件URL

错误码

- 400: 参数不合法（如缺少 query 或 user）
- 500: 服务端内部错误

curl 示例

```bash
curl -X POST http://127.0.0.1:8081/api/v1/chat/send \
  -H 'Content-Type: application/json' \
  -d '{
    "query": "What are the specs of the iPhone 13 Pro Max?",
    "user": "abc-123",
    "response_mode": "streaming"
  }'
```

---

### 12) 获取聊天历史

- 路径: `/api/v1/chat/history/:session_id`
- 方法: `GET`
- 说明: 获取指定会话的聊天历史记录。

请求参数

- `session_id`(string, 必填): 会话ID（路径参数）
- `limit`(int, 可选): 返回消息数量，默认 50，最大 100

请求示例

```bash
curl -X GET 'http://127.0.0.1:8081/api/v1/chat/history/session_1703123456789?limit=20' \
  -H 'Content-Type: application/json'
```

响应示例

```json
{
  "code": 0,
  "message": "success",
  "result": {
    "session_id": "session_1703123456789",
    "messages": [
      {
        "id": 1,
        "session_id": "session_1703123456789",
        "message_id": "msg_1703123456789",
        "event": "message",
        "content": "Hello, how are you?",
        "data": "{\"answer\": \"Hello, how are you?\"}",
        "created_at": 1703123456789
      }
    ],
    "total": 1
  }
}
```

字段说明

- `session_id`(string): 会话ID
- `messages`(array): 消息列表
  - `id`(int64): 消息ID
  - `message_id`(string): 消息唯一标识
  - `event`(string): 事件类型
  - `content`(string): 消息内容
  - `data`(string): 额外数据（JSON格式）
  - `created_at`(int64): 创建时间（毫秒时间戳）
- `total`(int): 消息总数

---

### 13) 获取聊天会话

- 路径: `/api/v1/chat/session/:session_id`
- 方法: `GET`
- 说明: 获取指定会话的基本信息。

请求参数

- `session_id`(string, 必填): 会话ID（路径参数）

请求示例

```bash
curl -X GET http://127.0.0.1:8081/api/v1/chat/session/session_1703123456789 \
  -H 'Content-Type: application/json'
```

响应示例

```json
{
  "code": 0,
  "message": "success",
  "result": {
    "id": 1,
    "session_id": "session_1703123456789",
    "user_id": "abc-123",
    "status": "active",
    "created_at": 1703123456789,
    "updated_at": 1703123456789
  }
}
```

字段说明

- `id`(int64): 会话ID
- `session_id`(string): 会话唯一标识
- `user_id`(string): 用户ID
- `status`(string): 会话状态（active-活跃, closed-关闭）
- `created_at`(int64): 创建时间（毫秒时间戳）
- `updated_at`(int64): 更新时间（毫秒时间戳）

注意事项

- 聊天接口使用Dify API作为后端服务，需要配置正确的API密钥
- 所有聊天消息和统计信息都会保存到数据库中
- 支持流式响应处理，实时保存每个事件
- 会话ID自动生成，格式为 session_时间戳
- 消息ID自动生成，格式为 msg_时间戳

---

## Chat 模块配置说明

### 配置文件结构

在 `config.yaml` 中添加以下配置：

```yaml
# 聊天服务配置
chat:
  # Dify API 基础地址
  base_url: "https://dify.example.com"
  # API 密钥
  api_key: "***REMOVED***"
  # 请求超时时间（秒）
  timeout: 60
  # 是否启用调试模式
  enable_debug: false
```

### 配置项说明

| 配置项 | 类型 | 必填 | 默认值 | 说明 |
|--------|------|------|--------|------|
| `base_url` | string | 是 | - | Dify API 的基础URL地址 |
| `api_key` | string | 是 | - | Dify API 的访问密钥 |
| `timeout` | int | 否 | 60 | HTTP请求超时时间（秒） |
| `enable_debug` | bool | 否 | false | 是否启用调试模式 |

### 当前配置

- **API地址**: `https://dify.example.com`
- **API密钥**: `***REMOVED***`
- **超时时间**: 60秒
- **调试模式**: 关闭

### 配置验证

启动服务后，可以通过以下方式验证配置：

1. **检查服务日志**: 确认chat模块配置加载成功
2. **测试API连接**: 尝试调用chat接口，验证API连接正常
3. **检查数据库**: 确认chat相关表创建成功

### 安全建议

1. **API密钥管理**: 确保API密钥的安全性，不要提交到版本控制系统
2. **网络访问**: 确保服务器能够访问配置的Dify API地址
3. **超时设置**: 根据实际网络情况调整超时时间
4. **调试模式**: 生产环境建议关闭调试模式

### 故障排除

#### 常见问题

1. **API密钥无效**
   - 检查API密钥是否正确
   - 确认API密钥是否过期
   - 验证API密钥权限

2. **网络连接失败**
   - 检查网络连接
   - 验证防火墙设置
   - 确认API地址可访问

3. **超时错误**
   - 增加超时时间
   - 检查网络延迟
   - 考虑使用CDN

#### 调试步骤

1. 启用调试模式：`enable_debug: true`
2. 检查服务日志
3. 验证配置文件格式
4. 测试网络连接
5. 检查API密钥有效性

### 系统提示词管理接口

- **POST** `/api/v2/openai/system-prompts` - 创建系统提示词
- **GET** `/api/v2/openai/system-prompts` - 获取系统提示词列表
- **GET** `/api/v2/openai/system-prompts/search` - 按名称模糊搜索系统提示词
- **GET** `/api/v2/openai/system-prompts/:id` - 获取单个系统提示词
- **PUT** `/api/v2/openai/system-prompts/:id` - 更新系统提示词
- **PATCH** `/api/v2/openai/system-prompts/:id/status` - 更新系统提示词激活状态
- **DELETE** `/api/v2/openai/system-prompts/:id` - 删除系统提示词
- **DELETE** `/api/v2/openai/system-prompts` - 批量删除系统提示词

详细文档请参考：
- [OpenAI API v2 详细文档](openai_api_v2.md)
- [系统提示词管理 API v2 详细文档](system_prompts_api_v2.md)


