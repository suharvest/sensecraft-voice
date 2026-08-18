# 关键词管理模块实现说明

## 概述

关键词管理模块已成功实现，提供完整的关键词增删改查功能，支持关键词、近义词和颜色标记的管理。

## 实现文件结构

```
pkg/db/model/keywords.go              # 数据模型定义
pkg/db/keywords.go                   # 数据库操作接口和实现
pkg/types/keywords.go                # 类型定义
pkg/controller/keywords/             # 控制器层
├── interface.go                     # 控制器接口定义
└── keywords.go                      # 控制器实现
api/server/router/keywords/          # 路由层
├── keywords.go                      # 路由定义
└── keywords_routes.go               # 路由处理函数
test_keywords_api.sh                 # API测试脚本
```

## 数据库表结构

```sql
CREATE TABLE `keywords` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `keyword` varchar(50) NOT NULL,
  `synonyms` varchar(500) NOT NULL,
  `mark_color` varchar(7) NOT NULL DEFAULT '#ff4d4f',
  `created_at` bigint NOT NULL,
  `updated_at` bigint NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_keywords_keyword` (`keyword`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

## API接口

### 基础路径
- **API版本**: `/api/v1/keywords`
- **认证方式**: Bearer Token
- **数据格式**: JSON

### 接口列表

| 方法 | 路径 | 功能 | 说明 |
|------|------|------|------|
| GET | `/api/v1/keywords` | 获取关键词列表 | 支持分页、搜索、颜色筛选 |
| GET | `/api/v1/keywords/:id` | 获取单个关键词 | 根据ID获取关键词详情 |
| POST | `/api/v1/keywords` | 创建关键词 | 创建新的关键词 |
| PUT | `/api/v1/keywords/:id` | 更新关键词 | 更新现有关键词 |
| DELETE | `/api/v1/keywords/:id` | 删除关键词 | 删除指定关键词 |
| DELETE | `/api/v1/keywords/batch` | 批量删除关键词 | 批量删除多个关键词 |

## 使用方法

### 1. 启动服务

确保数据库已正确配置，然后启动服务：

```bash
go run cmd/server.go
```

### 2. 测试API

使用提供的测试脚本：

```bash
# 使用默认配置测试
./test_keywords_api.sh

# 指定服务器地址和Token
./test_keywords_api.sh http://localhost:8080 your_token_here
```

### 3. 手动测试

#### 创建关键词
```bash
curl -X POST "http://localhost:8080/api/v1/keywords" \
  -H "Authorization: Bearer your_token_here" \
  -H "Content-Type: application/json" \
  -d '{
    "keyword": "投诉",
    "synonyms": "抱怨, 不满, 问题",
    "mark_color": "#ff4d4f"
  }'
```

#### 获取关键词列表
```bash
curl -X GET "http://localhost:8080/api/v1/keywords?offset=0&limit=10" \
  -H "Authorization: Bearer your_token_here"
```

#### 搜索关键词
```bash
curl -X GET "http://localhost:8080/api/v1/keywords?keyword=投诉" \
  -H "Authorization: Bearer your_token_here"
```

#### 按颜色筛选
```bash
curl -X GET "http://localhost:8080/api/v1/keywords?mark_color=%23ff4d4f" \
  -H "Authorization: Bearer your_token_here"
```

## 数据验证规则

### 关键词 (keyword)
- **必填**: 是
- **长度**: 1-50个字符
- **唯一性**: 在系统中必须唯一

### 近义词 (synonyms)
- **必填**: 是
- **长度**: 1-500个字符
- **格式**: 多个近义词用逗号分隔

### 标记颜色 (mark_color)
- **必填**: 是
- **格式**: 十六进制颜色值（如：#ff4d4f）
- **默认值**: #ff4d4f（红色）

## 错误处理

所有API都遵循项目的统一错误处理机制：
- HTTP状态码统一返回200
- 具体错误信息在响应的`code`和`message`字段中体现
- 常见错误码：400（请求参数错误、数据验证失败等）

## 注意事项

1. **数据库迁移**: 启动服务时会自动创建keywords表
2. **唯一性约束**: 关键词名称必须唯一，重复创建会返回错误
3. **颜色格式**: 颜色值必须使用十六进制格式，如#ff4d4f
4. **分页查询**: 建议使用分页查询，避免一次性获取大量数据
5. **批量操作**: 批量删除时，如果部分ID不存在，会返回成功删除的ID列表

## 扩展功能

如需扩展功能，可以考虑：
1. 添加关键词分类功能
2. 支持关键词权重设置
3. 添加关键词使用统计
4. 支持关键词导入导出
5. 添加关键词审核流程

## 更新日志

- **v1.0.0** (2024-01-25): 初始实现，支持基本的CRUD操作和批量删除功能
