# 关键词匹配系统设计文档

## 概述

本文档描述了关键词匹配系统的完整设计方案，包括正则匹配、数据存储和查询接口。

## 系统架构

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   录音文本输入    │───▶│   关键词匹配服务   │───▶│   匹配结果存储    │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │                        │
                                ▼                        ▼
                       ┌──────────────────┐    ┌─────────────────┐
                       │   异步处理        │    │   数据库存储     │
                       └──────────────────┘    └─────────────────┘
```

## 核心组件

### 1. 数据模型

#### KeywordMatch 关键词匹配记录
```go
type KeywordMatch struct {
    ID          int64   `gorm:"column:id;primaryKey;autoIncrement;not null"`
    RecordingID int64   `gorm:"column:recording_id;index:idx_recording_id;not null"`
    MacAddress  string  `gorm:"column:mac_address;type:varchar(32);index:idx_mac_keyword,priority:1;not null"`
    KeywordID   int64   `gorm:"column:keyword_id;index:idx_mac_keyword,priority:2;not null"`
    Keyword     string  `gorm:"column:keyword;type:varchar(50);not null"`
    MatchedText string  `gorm:"column:matched_text;type:text;not null"`
    MatchType   string  `gorm:"column:match_type;type:varchar(20);default:'exact';not null"`
    Confidence  float64 `gorm:"column:confidence;type:decimal(3,2);default:1.00;not null"`
    Position    int     `gorm:"column:position;not null"`
    Length      int     `gorm:"column:length;not null"`
    CreatedAt   int64   `gorm:"column:created_at;type:bigint;not null"`
}
```

### 2. 匹配算法

#### 精确匹配
- 使用正则表达式进行精确匹配
- 支持中文和英文文本
- 大小写不敏感匹配

```go
func (km *KeywordMatcher) exactMatch(text, keyword string) []TextMatch {
    var matches []TextMatch
    keywordLower := strings.ToLower(keyword)
    
    pattern := regexp.MustCompile(regexp.QuoteMeta(keywordLower))
    indices := pattern.FindAllStringIndex(text, -1)
    
    for _, index := range indices {
        matches = append(matches, TextMatch{
            Text:     text[index[0]:index[1]],
            Position: index[0],
            Length:   index[1] - index[0],
        })
    }
    
    return matches
}
```

#### 近义词匹配
- 将关键词的近义词按逗号分割
- 对每个近义词进行精确匹配
- 匹配置信度设置为0.8

```go
func (km *KeywordMatcher) synonymMatch(text, synonyms, originalKeyword string) []TextMatch {
    var matches []TextMatch
    synonymList := strings.Split(synonyms, ",")
    
    for _, synonym := range synonymList {
        synonym = strings.TrimSpace(synonym)
        if synonym == "" {
            continue
        }
        
        synonymMatches := km.exactMatch(text, synonym)
        matches = append(matches, synonymMatches...)
    }
    
    return matches
}
```

### 3. 异步处理机制

#### 异步处理设计
- 关键词匹配在独立的goroutine中执行
- 不阻塞录音保存的主流程
- 完整的错误处理和日志记录

```go
func (r *recording) processKeywordMatching(ctx context.Context, recordingID int64, macAddress, text string) {
    // 创建关键词匹配服务
    matcher := service.NewKeywordMatcher(r.factory)
    
    // 执行关键词匹配
    results, err := matcher.MatchKeywords(ctx, text, macAddress)
    if err != nil {
        klog.Errorf("Failed to match keywords: %v", err)
        return
    }

    // 保存匹配结果到数据库
    if err := matcher.SaveMatches(ctx, recordingID, macAddress, results); err != nil {
        klog.Errorf("Failed to save matches: %v", err)
        return
    }
}
```

### 4. 数据库设计

#### 表结构
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

#### 索引设计
- `idx_recording_id`: 根据录音ID快速查询
- `idx_mac_keyword`: 复合索引，支持按MAC地址和关键词ID查询

### 5. 集成流程

#### 录音保存时的关键词匹配
```go
func (r *recording) Save(ctx context.Context, in SaveRequest) (*model.Recording, error) {
    // 1. 保存录音记录
    out, err := r.factory.Recording().Create(ctx, obj)
    if err != nil {
        return nil, errors.ErrServerInternal
    }

    // 2. 异步执行关键词匹配
    go r.processKeywordMatching(context.Background(), out.ID, mac, in.Text)

    return out, nil
}
```

#### 关键词匹配处理
```go
func (r *recording) processKeywordMatching(ctx context.Context, recordingID int64, macAddress, text string) {
    // 1. 创建匹配服务
    matcher := service.NewKeywordMatcher(r.factory)
    
    // 2. 执行匹配
    results, err := matcher.MatchKeywords(ctx, text, macAddress)
    if err != nil {
        klog.Errorf("Failed to match keywords: %v", err)
        return
    }

    // 3. 保存匹配结果
    if err := matcher.SaveMatches(ctx, recordingID, macAddress, results); err != nil {
        klog.Errorf("Failed to save matches: %v", err)
        return
    }
}
```

## API接口

### 查询关键词匹配记录

#### 请求
```http
GET /api/v1/recordings/keyword-matches?mac_address=aa:bb:cc:dd:ee:01&offset=0&limit=20
Authorization: Bearer your_token_here
```

#### 响应
```json
{
  "code": 200,
  "message": "success",
  "result": {
    "total": 5,
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

### 查询参数
- `mac_address`: MAC地址筛选
- `keyword_id`: 关键词ID筛选
- `recording_id`: 录音ID筛选
- `offset`: 分页偏移量
- `limit`: 每页数量

## 性能优化

### 1. 异步处理
- 关键词匹配在独立的goroutine中执行
- 不阻塞录音保存的主流程

### 2. 批量操作
- 支持批量保存匹配结果
- 减少数据库连接次数

### 3. 索引优化
- 为常用查询字段建立复合索引
- 提高查询性能

### 4. 数据库优化
- 使用复合索引提高查询效率
- 支持分页查询避免大量数据加载

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

### 2. 保存录音（自动触发关键词匹配）
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

# 根据关键词ID查询
curl -X GET "http://localhost:8080/api/v1/recordings/keyword-matches?keyword_id=1" \
  -H "Authorization: Bearer your_token"
```

## 监控和日志

### 1. 日志记录
- 匹配成功/失败的日志
- 性能指标记录
- 错误信息记录

### 2. 统计信息
- 关键词匹配次数统计
- 按MAC地址的匹配统计
- 按时间段的匹配统计

## 扩展功能

### 1. 高级匹配算法
- 支持模糊匹配
- 支持语义匹配
- 支持机器学习模型

### 2. 实时通知
- 匹配到特定关键词时发送通知
- 支持WebSocket实时推送

### 3. 数据分析
- 关键词热度分析
- 匹配趋势分析
- 用户行为分析

## 注意事项

1. **性能考虑**: 关键词数量过多时可能影响匹配性能
2. **数据一致性**: 异步处理可能导致数据延迟
3. **错误处理**: 匹配失败不应影响录音保存
4. **扩展性**: 设计时考虑了水平扩展的可能性
