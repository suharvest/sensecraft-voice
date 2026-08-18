# 关键词匹配系统 - 简化版设计

## 概述

这是一个简化的关键词匹配系统，不使用缓存，只依赖数据库存储。系统在录音文本存储时自动进行关键词匹配，并将匹配结果存储到数据库中。

## 系统架构

```
录音文本 → 关键词匹配服务 → 数据库存储
    ↓              ↓              ↓
异步处理        正则匹配        批量保存
```

## 核心组件

### 1. 数据模型
- `KeywordMatch`: 关键词匹配记录
- 包含录音ID、MAC地址、关键词ID、匹配文本、位置等信息

### 2. 匹配算法
- **精确匹配**: 使用正则表达式匹配关键词
- **近义词匹配**: 匹配关键词的近义词
- **位置记录**: 记录匹配文本在原文中的位置和长度

### 3. 数据库设计
```sql
CREATE TABLE `keyword_matches` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `recording_id` bigint NOT NULL,
  `mac_address` varchar(32) NOT NULL,
  `keyword_id` bigint NOT NULL,
  `keyword` varchar(50) NOT NULL,
  `matched_text` text NOT NULL,
  `match_type` varchar(20) NOT NULL DEFAULT 'exact',
  `confidence` decimal(3,2) NOT NULL DEFAULT '1.00',
  `position` int NOT NULL,
  `length` int NOT NULL,
  `created_at` bigint NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_recording_id` (`recording_id`),
  KEY `idx_mac_keyword` (`mac_address`,`keyword_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

## 工作流程

1. **录音保存**: 用户保存录音文本
2. **异步匹配**: 在独立goroutine中执行关键词匹配
3. **结果存储**: 将匹配结果批量保存到数据库
4. **查询接口**: 提供多种查询方式

## API接口

### 查询关键词匹配记录
```http
GET /api/v1/recordings/keyword-matches?mac_address=aa:bb:cc:dd:ee:01&offset=0&limit=20
```

### 查询参数
- `mac_address`: MAC地址筛选
- `keyword_id`: 关键词ID筛选
- `recording_id`: 录音ID筛选
- `offset`: 分页偏移量
- `limit`: 每页数量

## 性能特点

- **异步处理**: 不阻塞录音保存流程
- **批量操作**: 减少数据库连接次数
- **索引优化**: 提高查询性能
- **错误处理**: 匹配失败不影响录音保存

## 使用示例

### 1. 创建关键词
```bash
curl -X POST "http://localhost:8080/api/v1/keywords" \
  -H "Authorization: Bearer your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "keyword": "投诉",
    "synonyms": "抱怨, 不满, 问题, 意见",
    "mark_color": "#ff4d4f"
  }'
```

### 2. 保存录音（自动触发匹配）
```bash
curl -X POST "http://localhost:8080/api/v1/recordings" \
  -H "Authorization: Bearer your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "test_session_001",
    "audio_id": "test_audio_001",
    "mac_address": "aa:bb:cc:dd:ee:01",
    "speaker": {
      "speaker_id": "speaker_001",
      "speaker_name": "客户A"
    },
    "text": "我对这个产品的质量很不满，想要投诉，服务态度也有问题",
    "timestamp": 1703123456789,
    "status": 1
  }'
```

### 3. 查询匹配结果
```bash
# 根据MAC地址查询
curl -X GET "http://localhost:8080/api/v1/recordings/keyword-matches?mac_address=aa:bb:cc:dd:ee:01" \
  -H "Authorization: Bearer your_token"
```

## 测试

运行测试脚本：
```bash
./test_keyword_matching.sh
```

## 优势

- **简单**: 不依赖Redis等外部缓存
- **可靠**: 数据持久化存储
- **高效**: 异步处理，不阻塞主流程
- **易维护**: 代码结构清晰，文档完整

## 注意事项

1. 关键词数量过多时可能影响匹配性能
2. 异步处理可能导致数据延迟
3. 匹配失败不应影响录音保存
4. 建议定期清理旧的匹配记录
