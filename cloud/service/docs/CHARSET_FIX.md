# 数据库字符集问题修复指南

## 🚨 问题描述

在测试OpenAI API时遇到以下错误：

```
Error 1366: Incorrect string value: '\xE5\xB8\xAE\xE6\x88\x91...' for column 'content' at row 1
```

这是一个典型的MySQL字符集问题，数据库无法正确处理UTF-8编码的中文字符。

## 🔍 问题原因

1. **数据库连接字符集**: 连接字符串使用了 `charset=utf8`，只支持3字节UTF-8字符
2. **表字符集**: 数据库表可能使用了不完整的UTF-8字符集
3. **GORM模型**: 字段定义没有明确指定字符集

## ✅ 修复方案

### 1. 修复数据库连接字符集

**文件**: `cmd/app/options/options.go`

**修改前**:
```go
dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
    sqlConfig.User,
    sqlConfig.Password,
    sqlConfig.Host,
    sqlConfig.Port,
    sqlConfig.Name)
```

**修改后**:
```go
dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
    sqlConfig.User,
    sqlConfig.Password,
    sqlConfig.Host,
    sqlConfig.Port,
    sqlConfig.Name)
```

### 2. 修复GORM模型定义

**文件**: `pkg/db/model/chat.go`

**修改前**:
```go
type ChatMessage struct {
    // ...
    Content   string `gorm:"type:text"`
    Data      string `gorm:"type:longtext"`
    // ...
}
```

**修改后**:
```go
type ChatMessage struct {
    // ...
    Content   string `gorm:"type:text;charset:utf8mb4"`
    Data      string `gorm:"type:longtext;charset:utf8mb4"`
    // ...
}
```

### 3. 修复数据库表字符集

**方法1: 使用修复脚本**

```bash
# 运行数据库字符集修复脚本
./fix_database_charset.sh
```

**方法2: 手动执行SQL**

```sql
-- 修改数据库字符集
ALTER DATABASE sensecraft_voice CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 修改表字符集
ALTER TABLE chat_messages CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
ALTER TABLE chat_sessions CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
ALTER TABLE chat_stats CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

## 🧪 验证修复

### 1. 重新编译服务

```bash
go build -o server cmd/server.go
```

### 2. 启动服务

```bash
./server
```

### 3. 运行测试

```bash
# 测试中文聊天
./test_chinese_chat.sh

# 或运行完整测试
./test_openai_api.sh
```

## 📊 字符集对比

| 字符集 | 支持范围 | 字节长度 | 说明 |
|--------|----------|----------|------|
| `utf8` | 基本多文种平面 | 1-3字节 | 不支持emoji和4字节字符 |
| `utf8mb4` | 完整Unicode | 1-4字节 | 支持所有Unicode字符 |

## 🔧 配置检查

### 检查数据库字符集

```sql
-- 检查数据库字符集
SELECT DEFAULT_CHARACTER_SET_NAME, DEFAULT_COLLATION_NAME 
FROM information_schema.SCHEMATA 
WHERE SCHEMA_NAME = 'sensecraft_voice';

-- 检查表字符集
SELECT 
    TABLE_NAME,
    TABLE_COLLATION,
    CHARACTER_SET_NAME
FROM information_schema.TABLES t
JOIN information_schema.COLLATION_CHARACTER_SET_APPLICABILITY ccs
    ON t.TABLE_COLLATION = ccs.COLLATION_NAME
WHERE t.TABLE_SCHEMA = 'sensecraft_voice'
    AND t.TABLE_NAME IN ('chat_messages', 'chat_sessions', 'chat_stats');
```

### 检查列字符集

```sql
-- 检查列字符集
SELECT 
    TABLE_NAME,
    COLUMN_NAME,
    CHARACTER_SET_NAME,
    COLLATION_NAME
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = 'sensecraft_voice'
    AND TABLE_NAME IN ('chat_messages', 'chat_sessions', 'chat_stats')
    AND CHARACTER_SET_NAME IS NOT NULL;
```

## 🚀 最佳实践

### 1. 数据库设计

- 始终使用 `utf8mb4` 字符集
- 使用 `utf8mb4_unicode_ci` 排序规则
- 在连接字符串中指定 `charset=utf8mb4`

### 2. GORM模型

- 为文本字段明确指定字符集
- 使用 `gorm:"type:text;charset:utf8mb4"` 标签

### 3. 应用配置

- 确保数据库连接使用正确的字符集
- 在应用启动时验证字符集配置

## 📝 相关文件

- `cmd/app/options/options.go` - 数据库连接配置
- `pkg/db/model/chat.go` - GORM模型定义
- `fix_database_charset.sh` - 数据库修复脚本
- `test_chinese_chat.sh` - 中文测试脚本

## ✅ 修复验证

修复完成后，应该能够：

1. ✅ 正常保存和读取中文内容
2. ✅ 支持emoji和特殊字符
3. ✅ 流式聊天正常工作
4. ✅ 聊天历史正确显示

## 🔍 故障排除

如果问题仍然存在：

1. **检查MySQL版本**: 确保MySQL版本支持utf8mb4
2. **重启服务**: 修改连接配置后需要重启服务
3. **清理缓存**: 清理应用缓存和数据库连接池
4. **检查权限**: 确保数据库用户有修改表结构的权限

---

**修复时间**: 2024年12月  
**影响范围**: OpenAI API v2 聊天功能  
**修复状态**: ✅ 已完成
