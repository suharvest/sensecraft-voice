# SenseCraft Voice Service 系统架构文档

## 📋 系统概述

SenseCraft Voice Service 是一个基于Go语言开发的语音识别服务系统，采用分层架构设计，支持语音录制、识别、存储、聊天对话等核心功能。

## 🏗️ 系统架构

### 整体架构图

```mermaid
graph TB
    subgraph "客户端层"
        A[Web客户端] 
        B[移动端]
        C[IoT设备]
    end
    
    subgraph "API网关层"
        D[HTTP/HTTPS]
        E[WebSocket]
        F[健康检查]
    end
    
    subgraph "应用服务层"
        G[Gin HTTP引擎]
        H[中间件层]
        I[路由层]
    end
    
    subgraph "业务逻辑层"
        J[Controller层]
        K[Service层]
    end
    
    subgraph "数据访问层"
        L[DAO层]
        M[数据模型]
    end
    
    subgraph "数据存储层"
        N[MySQL数据库]
        O[缓存层]
    end
    
    subgraph "外部服务"
        P[Dify AI服务]
        Q[Seeed API]
    end
    
    A --> D
    B --> D
    C --> E
    D --> G
    E --> G
    F --> G
    G --> H
    H --> I
    I --> J
    J --> K
    K --> L
    L --> M
    M --> N
    L --> O
    J --> P
    J --> Q
```

## 🎯 核心模块

### 1. 录音管理模块 (Recording)

```mermaid
graph LR
    A[WebSocket流式输入] --> B[录音数据验证]
    C[HTTP POST接口] --> B
    B --> D[数据存储]
    D --> E[录音记录表]
    
    F[录音查询接口] --> G[条件查询]
    G --> E
    G --> H[分页结果]
```

**功能特性:**
- 支持WebSocket实时流式录音数据接收
- 支持HTTP POST批量录音数据提交
- 支持按设备、时间、门店等条件查询
- 自动处理中间结果和最终结果

### 2. 设备管理模块 (Device)

```mermaid
graph TD
    A[设备注册] --> B[设备信息管理]
    B --> C[设备状态监控]
    C --> D[设备分配]
    D --> E[点位关联]
    E --> F[门店关联]
    
    G[设备查询] --> H[多条件筛选]
    H --> I[关联查询]
    I --> J[分页结果]
```

### 3. 门店点位管理模块 (Store & Location)

```mermaid
graph TB
    A[门店管理] --> B[门店CRUD]
    B --> C[点位管理]
    C --> D[点位CRUD]
    D --> E[设备分配]
    E --> F[关联关系]
    
    G[层级结构] --> H[门店]
    H --> I[点位]
    I --> J[设备]
```

### 4. 聊天对话模块 (Chat)

```mermaid
graph LR
    A[用户输入] --> B[会话管理]
    B --> C[Dify API调用]
    C --> D[流式响应]
    D --> E[消息存储]
    E --> F[Seeed API集成]
    F --> G[结果返回]
```

### 5. 用户管理模块 (User)

```mermaid
graph TD
    A[用户注册] --> B[身份验证]
    B --> C[用户信息管理]
    C --> D[权限控制]
    D --> E[会话管理]
```

## 🔧 技术栈

### 后端技术
- **语言**: Go 1.19+
- **Web框架**: Gin
- **数据库**: MySQL 8.0+
- **ORM**: GORM
- **WebSocket**: Gorilla WebSocket
- **配置管理**: Viper
- **日志**: klog
- **容器化**: Docker

### 外部集成
- **AI服务**: Dify API
- **硬件集成**: Seeed API
- **文档**: Swagger

## 📁 代码结构

### 分层架构

```mermaid
graph TB
    subgraph "API层 (api/)"
        A1[路由定义]
        A2[请求处理]
        A3[中间件]
        A4[参数验证]
    end
    
    subgraph "业务逻辑层 (pkg/controller/)"
        B1[业务逻辑]
        B2[数据验证]
        B3[外部服务调用]
    end
    
    subgraph "数据访问层 (pkg/db/)"
        C1[DAO接口]
        C2[数据模型]
        C3[数据库操作]
    end
    
    subgraph "工具层 (pkg/util/)"
        D1[HTTP客户端]
        D2[日志工具]
        D3[加密工具]
        D4[缓存工具]
    end
    
    subgraph "插件层 (pkg/plugins/)"
        E1[外部API集成]
        E2[缓存工具]
    end
    
    A1 --> A2
    A2 --> B1
    B1 --> C1
    C1 --> C2
    B1 --> D1
    B1 --> E1
```

### 目录结构

```
sensecraft-voice-service/
├── api/server/           # API服务层
│   ├── router/          # 路由层
│   ├── middleware/      # 中间件
│   ├── httputils/       # HTTP工具
│   └── validator/       # 参数验证
├── pkg/                 # 核心包
│   ├── controller/      # 业务逻辑层
│   ├── db/             # 数据访问层
│   ├── util/           # 工具包
│   ├── plugins/        # 插件
│   └── types/          # 类型定义
├── cmd/app/            # 应用入口
├── docs/               # 文档
└── config.yaml         # 配置文件
```

## 🔄 数据流

### 录音数据流

```mermaid
sequenceDiagram
    participant C as 客户端
    participant W as WebSocket/HTTP
    participant R as Router
    participant Ctrl as Controller
    participant DB as Database
    
    C->>W: 发送录音数据
    W->>R: 路由到处理函数
    R->>Ctrl: 调用Save方法
    Ctrl->>Ctrl: 数据验证
    Ctrl->>DB: 存储录音记录
    DB-->>Ctrl: 返回存储结果
    Ctrl-->>R: 返回处理结果
    R-->>W: 返回响应
    W-->>C: 确认消息
```

### 聊天对话流

```mermaid
sequenceDiagram
    participant U as 用户
    participant C as 客户端
    participant S as 服务端
    participant D as Dify API
    participant Se as Seeed API
    
    U->>C: 输入消息
    C->>S: POST /api/v1/chat/send
    S->>S: 创建会话
    S->>D: 调用AI服务
    D-->>S: 流式响应
    S->>S: 存储消息
    S->>Se: 发送到Seeed API
    Se-->>S: 返回结果
    S-->>C: 返回对话结果
    C-->>U: 显示回复
```

## 🛡️ 安全机制

### 认证授权
- Token验证
- 权限控制 (Casbin)
- 请求限流
- CORS支持

### 数据安全
- 密码MD5加密
- 输入参数验证
- SQL注入防护
- 错误信息脱敏

## 📊 监控与日志

### 日志系统
- 结构化日志 (JSON格式)
- 分级日志 (Info/Warn/Error)
- 请求链路追踪
- 性能监控

### 健康检查
- `/healthz` 端点
- 数据库连接检查
- 外部服务状态检查

## 🚀 部署架构

### Docker部署

```mermaid
graph TB
    subgraph "Docker容器"
        A[SenseCraft Voice Service]
        B[MySQL数据库]
    end
    
    subgraph "外部服务"
        C[Dify AI服务]
        D[Seeed API]
    end
    
    A --> B
    A --> C
    A --> D
```

### 配置管理
- 环境变量支持
- 配置文件热加载
- 多环境配置

## 🔮 扩展性设计

### 水平扩展
- 无状态服务设计
- 数据库读写分离
- 缓存层支持
- 负载均衡友好

### 功能扩展
- 插件化架构
- 模块化设计
- 接口标准化
- 版本兼容性

## 📈 性能优化

### 数据库优化
- 索引优化
- 查询优化
- 连接池管理
- 分页查询

### 缓存策略
- LRU缓存
- Token缓存
- 用户信息缓存
- 查询结果缓存

## 🧪 测试策略

### 测试覆盖
- 单元测试
- 集成测试
- API测试
- 性能测试

### 测试工具
- 内置测试脚本
- 自动化测试
- 压力测试
- 监控告警

---

## 📝 总结

SenseCraft Voice Service 采用现代化的微服务架构设计，具有以下特点：

✅ **高可扩展性**: 模块化设计，易于扩展新功能  
✅ **高性能**: 异步处理，缓存优化，数据库优化  
✅ **高可用性**: 健康检查，错误处理，优雅关闭  
✅ **易维护性**: 清晰的分层架构，标准化接口  
✅ **安全性**: 完整的认证授权，数据加密，输入验证  

系统为语音识别和智能对话提供了完整的解决方案，支持多种客户端接入方式，具备良好的扩展性和维护性。
