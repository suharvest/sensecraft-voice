# 门店和点位管理API文档

## 概述
本文档描述了门店->点位->设备的三级层级管理系统的API接口。

## 数据模型

### 门店 (Store)
- 门店是最高层级，包含基本信息如名称、代码、地址、联系人等
- 一个门店可以有多个点位

### 点位 (Location)
- 点位属于门店，是门店内的具体位置
- 一个点位可以有多个设备
- 点位包含名称、代码、描述等信息

### 设备 (Device)
- 设备可以分配到具体的点位
- 设备自动关联到点位所属的门店

## API接口

### 门店管理

#### 1. 创建门店
```http
POST /api/v1/stores
Content-Type: application/json

{
  "name": "北京朝阳店",
  "code": "BJ_CY_001",
  "address": "北京市朝阳区xxx街道xxx号",
  "contact": "张经理",
  "phone": "13800138000",
  "status": 1
}
```

#### 2. 查询门店列表
```http
GET /api/v1/stores?offset=0&limit=20
```

#### 3. 查询门店详情
```http
GET /api/v1/stores/{id}
```

#### 4. 更新门店
```http
PUT /api/v1/stores/{id}
Content-Type: application/json

{
  "name": "北京朝阳店(更新)",
  "address": "北京市朝阳区xxx街道xxx号(更新)"
}
```

#### 5. 删除门店
```http
DELETE /api/v1/stores/{id}
```

### 点位管理

#### 1. 创建点位
```http
POST /api/v1/locations
Content-Type: application/json

{
  "store_id": 1,
  "name": "一楼大厅",
  "code": "L1_HALL",
  "description": "一楼主要营业区域",
  "status": 1
}
```

#### 2. 查询点位列表
```http
GET /api/v1/locations?offset=0&limit=20
```

#### 3. 按门店查询点位
```http
GET /api/v1/stores/{storeId}/locations?offset=0&limit=20
```

#### 4. 查询点位详情
```http
GET /api/v1/locations/{id}
```

#### 5. 更新点位
```http
PUT /api/v1/locations/{id}
Content-Type: application/json

{
  "name": "一楼大厅(更新)",
  "description": "一楼主要营业区域(更新)"
}
```

#### 6. 删除点位
```http
DELETE /api/v1/locations/{id}
```

### 设备管理扩展

#### 1. 设备分配到点位
```http
PUT /api/v1/devices/{id}/assign
Content-Type: application/json

{
  "location_id": 1
}
```

#### 2. 查询点位的设备
```http
GET /api/v1/locations/{locationId}/devices?offset=0&limit=20
```

#### 3. 查询门店的所有设备
```http
GET /api/v1/stores/{storeId}/devices?offset=0&limit=20
```

## 业务规则

1. **层级关系**: 门店 -> 点位 -> 设备
2. **数据完整性**: 
   - 删除门店前必须先删除所有相关点位
   - 删除点位前必须先解除所有设备的关联
3. **设备分配**: 设备分配到点位后，自动关联到点位所属的门店

## 响应格式

### 成功响应
```json
{
  "code": 200,
  "result": {
    // 具体数据
  }
}
```

### 错误响应
```json
{
  "code": 400,
  "message": "错误信息"
}
```

## 使用示例

### 完整流程示例

1. **创建门店**
```bash
curl -X POST http://localhost:8080/api/v1/stores \
  -H "Content-Type: application/json" \
  -d '{
    "name": "北京朝阳店",
    "code": "BJ_CY_001",
    "address": "北京市朝阳区xxx街道xxx号",
    "contact": "张经理",
    "phone": "13800138000"
  }'
```

2. **创建点位**
```bash
curl -X POST http://localhost:8080/api/v1/locations \
  -H "Content-Type: application/json" \
  -d '{
    "store_id": 1,
    "name": "一楼大厅",
    "code": "L1_HALL",
    "description": "一楼主要营业区域"
  }'
```

3. **设备分配到点位**
```bash
curl -X PUT http://localhost:8080/api/v1/devices/1/assign \
  -H "Content-Type: application/json" \
  -d '{
    "location_id": 1
  }'
```

4. **查询点位的设备**
```bash
curl "http://localhost:8080/api/v1/locations/1/devices?limit=10"
```

5. **查询门店的所有设备**
```bash
curl "http://localhost:8080/api/v1/stores/1/devices?limit=10"
```
