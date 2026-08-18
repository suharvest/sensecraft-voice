# 关键词管理 API 集成指南

## 概述

本文档说明如何测试和验证关键词管理API的集成情况。

## 已完成的集成

### ✅ 前端集成完成

1. **API服务更新** (`src/services/keywords.ts`)
   - 更新所有API路径为 `/api/v1/keywords`
   - 匹配文档中的v1版本API规范
   - 支持统一错误处理机制

2. **页面功能集成** (`src/pages/Settings/SettingsPage.tsx`)
   - 关键词管理标签页
   - 调色盘颜色选择器
   - 默认红色设置
   - 批量删除功能

3. **类型定义** (`src/types/index.ts`)
   - 完整的TypeScript类型定义
   - 匹配API文档规范

## 测试方法

### 方法1: 浏览器控制台测试

1. 打开浏览器开发者工具 (F12)
2. 在控制台中运行：
   ```javascript
   testKeywordsAPI()
   ```
3. 查看测试结果和API调用情况

### 方法2: 页面功能测试

1. 访问设置页面 (`/settings`)
2. 切换到"关键词"标签页
3. 尝试以下操作：
   - 点击"新增关键词"按钮
   - 填写关键词信息
   - 选择颜色（默认红色）
   - 保存关键词
   - 编辑现有关键词
   - 批量删除关键词

### 方法3: 网络面板监控

1. 打开浏览器开发者工具
2. 切换到"Network"标签页
3. 在设置页面进行操作
4. 观察API请求：
   - `GET /api/v1/keywords` - 获取列表
   - `POST /api/v1/keywords` - 创建关键词
   - `PUT /api/v1/keywords/:id` - 更新关键词
   - `DELETE /api/v1/keywords/:id` - 删除关键词
   - `DELETE /api/v1/keywords/batch` - 批量删除

## API集成状态

### 前端状态: ✅ 完成

- [x] API服务层集成
- [x] 页面UI集成
- [x] 类型定义
- [x] 错误处理
- [x] 测试工具

### 后端状态: ⏳ 待实现

根据 `docs/keywords_api.md` 文档，需要实现以下后端接口：

- [ ] `GET /api/v1/keywords` - 获取关键词列表
- [ ] `GET /api/v1/keywords/:id` - 获取单个关键词
- [ ] `POST /api/v1/keywords` - 创建关键词
- [ ] `PUT /api/v1/keywords/:id` - 更新关键词
- [ ] `DELETE /api/v1/keywords/:id` - 删除关键词
- [ ] `DELETE /api/v1/keywords/batch` - 批量删除关键词

## 后端实现指南

### 1. 数据库表结构

```sql
CREATE TABLE keywords (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    keyword VARCHAR(50) UNIQUE NOT NULL,
    synonyms VARCHAR(500) NOT NULL,
    mark_color VARCHAR(7) DEFAULT '#ff4d4f' NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);
```

### 2. 实现文件结构

```
pkg/db/model/keywords.go          # 数据模型定义
pkg/db/keywords.go               # 数据库操作接口和实现
pkg/controller/keywords/         # 控制器层
├── interface.go                 # 控制器接口定义
└── keywords.go                  # 控制器实现
api/server/router/keywords/      # 路由层
├── keywords.go                  # 路由定义
└── keywords_routes.go           # 路由处理函数
```

### 3. 关键实现要点

- 统一返回格式：`{code: 200, message: "success", result: data}`
- 错误处理：所有错误都返回HTTP 200，错误信息在code字段中
- 分页支持：offset/limit参数
- 搜索功能：keyword和mark_color筛选
- 批量操作：支持批量删除

## 测试数据

### 示例关键词数据

```json
[
  {
    "id": 1,
    "keyword": "投诉",
    "synonyms": "抱怨, 不满, 问题",
    "mark_color": "#ff4d4f",
    "created_at": 1703123456789,
    "updated_at": 1703123456789
  },
  {
    "id": 2,
    "keyword": "满意",
    "synonyms": "开心, 认可, 不错",
    "mark_color": "#52c41a",
    "created_at": 1703123456790,
    "updated_at": 1703123456790
  },
  {
    "id": 3,
    "keyword": "价格",
    "synonyms": "多少钱, 怎么卖",
    "mark_color": "#1890ff",
    "created_at": 1703123456791,
    "updated_at": 1703123456791
  }
]
```

## 故障排除

### 常见问题

1. **API调用失败**
   - 检查后端服务是否启动
   - 确认API路径是否正确 (`/api/v1/keywords`)
   - 检查认证Token是否有效

2. **颜色选择器问题**
   - 确认ColorPicker组件已正确导入
   - 检查默认颜色设置

3. **批量删除失败**
   - 检查API响应格式
   - 确认返回的deleted_count字段

### 调试技巧

1. 使用浏览器控制台查看详细错误信息
2. 检查Network面板中的API请求和响应
3. 使用 `testKeywordsAPI()` 函数进行API测试
4. 查看控制台日志了解API调用流程

## 下一步

1. 实现后端API接口
2. 测试完整的CRUD功能
3. 优化用户体验
4. 添加更多验证和错误处理
5. 性能优化和缓存策略

## 联系支持

如有问题，请参考：
- API文档：`docs/keywords_api.md`
- 前端代码：`src/pages/Settings/SettingsPage.tsx`
- API服务：`src/services/keywords.ts`
