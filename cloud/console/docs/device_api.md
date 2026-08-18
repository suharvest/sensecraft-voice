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
        "version": "v1.2.3",
        "cpu_usage_percent": 23.45,
        "memory_used_bytes": 536870912,
        "disk_used_bytes": 2147483648,
        "swap_used_bytes": 0,
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
    "version": "v1.2.3",
    "cpu_usage_percent": 23.45,
    "memory_used_bytes": 536870912,
    "disk_used_bytes": 2147483648,
    "swap_used_bytes": 0,
    "created_at": 1756388344600,
    "updated_at": 1756388344600
  }
}
```

错误码

- 400: 参数不合法
- 404: 设备不存在
- 500: 服务端内部错误