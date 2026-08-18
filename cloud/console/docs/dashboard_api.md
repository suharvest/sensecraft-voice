### 8) 统计接口

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
- `weekly_record_trend`(array): 本周录音记录增长趋势
  - `date`(string): 日期，格式 YYYY-MM-DD
  - `count`(int64): 当日录音记录数量
- `today_active_devices`(int64): 今日产生录音记录的设备数量
- `today_hourly_distribution`(array): 今日每小时采集分布（00:00-23:59）
  - `hour`(int): 小时，0-23
  - `count`(int64): 该小时的记录数量
- `today_keyword_triggers`(int64): 今日关键词触发总次数
- `today_keyword_matches`(array): 今日关键词匹配详情
  - `keyword_id`(int64): 关键词ID
  - `keyword`(string): 关键词内容
  - `mark_color`(string): 关键词标记颜色
  - `match_count`(int64): 匹配次数
  - `record_count`(int64): 涉及到的记录数

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
- 关键词匹配统计按匹配次数降序排列