# 用户管理API文档

## 概述
本文档描述了用户管理系统的API接口，包括用户注册、登录、更新、删除等功能。

## 数据模型

### 用户模型 (User)
- `id`: 用户ID，自增主键
- `username`: 用户名，唯一，3-20个字符
- `password`: 密码，MD5加密存储，6-20个字符
- `created_at`: 创建时间(毫秒)
- `updated_at`: 更新时间(毫秒)

## API接口

### 1. 用户注册
```http
POST /api/v1/users/register
Content-Type: application/json

{
  "username": "newuser",
  "password": "password123"
}
```

**响应示例**
```json
{
  "code": 200,
  "result": {
    "id": 3,
    "username": "newuser",
    "created_at": 1756388344546,
    "updated_at": 1756388344546
  }
}
```

### 2. 用户登录
```http
POST /api/v1/users/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123"
}
```

**响应示例**
```json
{
  "code": 200,
  "result": {
    "user": {
      "id": 1,
      "username": "admin",
      "created_at": 1756388344546,
      "updated_at": 1756388344546
    },
    "token": ""
  }
}
```

### 3. 查询用户列表
```http
GET /api/v1/users?offset=0&limit=20
```

**响应示例**
```json
{
  "code": 200,
  "result": {
    "total": 2,
    "items": [
      {
        "id": 1,
        "username": "admin",
        "created_at": 1756388344546,
        "updated_at": 1756388344546
      },
      {
        "id": 2,
        "username": "testuser",
        "created_at": 1756388344546,
        "updated_at": 1756388344546
      }
    ]
  }
}
```

### 4. 查询用户详情
```http
GET /api/v1/users/{id}
```

**响应示例**
```json
{
  "code": 200,
  "result": {
    "id": 1,
    "username": "admin",
    "created_at": 1756388344546,
    "updated_at": 1756388344546
  }
}
```

### 5. 更新用户信息
```http
PUT /api/v1/users/{id}
Content-Type: application/json

{
  "username": "admin_updated"
}
```

**响应示例**
```json
{
  "code": 200,
  "result": {
    "id": 1,
    "username": "admin_updated",
    "created_at": 1756388344546,
    "updated_at": 1756388344547
  }
}
```

### 6. 修改密码
```http
PUT /api/v1/users/{id}/password
Content-Type: application/json

{
  "old_password": "admin123",
  "new_password": "newpassword123"
}
```

**响应示例**
```json
{
  "code": 200,
  "result": {
    "message": "密码修改成功"
  }
}
```

### 7. 删除用户
```http
DELETE /api/v1/users/{id}
```

**响应示例**
```json
{
  "code": 200,
  "result": {
    "message": "删除成功"
  }
}
```

## 业务规则

### 1. 用户名规则
- 长度：3-20个字符
- 唯一性：用户名不能重复
- 更新时检查唯一性

### 2. 密码规则
- 长度：6-20个字符
- 存储：使用MD5加密
- 修改：需要验证旧密码

### 3. 数据完整性
- 创建和更新时间自动维护
- 删除操作不可恢复
- 密码字段不返回给前端

## 错误处理

### 常见错误码
- `400`: 参数验证失败
- `500`: 服务器内部错误

### 错误响应格式
```json
{
  "code": 400,
  "message": "错误信息"
}
```

## 使用示例

### 完整流程示例

1. **注册新用户**
```bash
curl -X POST http://localhost:8080/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "newuser",
    "password": "password123"
  }'
```

2. **用户登录**
```bash
curl -X POST http://localhost:8080/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "newuser",
    "password": "password123"
  }'
```

3. **查询用户列表**
```bash
curl "http://localhost:8080/api/v1/users?limit=10"
```

4. **修改密码**
```bash
curl -X PUT http://localhost:8080/api/v1/users/1/password \
  -H "Content-Type: application/json" \
  -d '{
    "old_password": "password123",
    "new_password": "newpassword456"
  }'
```

5. **删除用户**
```bash
curl -X DELETE http://localhost:8080/api/v1/users/1
```

## 安全考虑

1. **密码加密**: 使用MD5加密存储密码
2. **参数验证**: 严格的输入参数验证
3. **错误信息**: 不暴露敏感的系统信息
4. **唯一性约束**: 数据库级别的用户名唯一性约束

## 扩展功能

### 预留功能
- JWT Token认证
- 用户角色管理
- 密码强度验证
- 登录日志记录
- 用户状态管理
