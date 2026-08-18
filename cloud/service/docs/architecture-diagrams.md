# SenseCraft Voice Service 架构图表

## 1. 整体系统架构

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

## 2. 录音管理模块架构

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

## 3. 设备管理模块架构

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

## 4. 门店点位管理架构

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

## 5. 聊天对话模块架构

```mermaid
graph LR
    A[用户输入] --> B[会话管理]
    B --> C[Dify API调用]
    C --> D[流式响应]
    D --> E[消息存储]
    E --> F[Seeed API集成]
    F --> G[结果返回]
```

## 6. 用户管理模块架构

```mermaid
graph TD
    A[用户注册] --> B[身份验证]
    B --> C[用户信息管理]
    C --> D[权限控制]
    D --> E[会话管理]
```

## 7. 代码分层架构

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

## 8. 录音数据流时序图

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

## 9. 聊天对话流时序图

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

## 10. 部署架构图

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

## 11. 模块关系图

```mermaid
graph TD
    A[录音管理] --> B[设备管理]
    B --> C[门店管理]
    C --> D[点位管理]
    D --> B
    
    E[用户管理] --> F[聊天对话]
    F --> G[Dify AI集成]
    F --> H[Seeed API集成]
    
    I[统计模块] --> A
    I --> B
    I --> C
    I --> D
    
    J[审计模块] --> A
    J --> B
    J --> C
    J --> D
    J --> E
    J --> F
```

## 12. 数据模型关系图

```mermaid
erDiagram
    STORES ||--o{ LOCATIONS : contains
    LOCATIONS ||--o{ DEVICES : contains
    DEVICES ||--o{ RECORDINGS : generates
    USERS ||--o{ CHAT_SESSIONS : creates
    CHAT_SESSIONS ||--o{ CHAT_MESSAGES : contains
    CHAT_SESSIONS ||--o{ CHAT_STATS : generates
    
    STORES {
        int id PK
        string name
        string address
        datetime created_at
    }
    
    LOCATIONS {
        int id PK
        int store_id FK
        string name
        string description
        datetime created_at
    }
    
    DEVICES {
        int id PK
        int location_id FK
        int store_id FK
        string mac_address
        string device_name
        datetime created_at
    }
    
    RECORDINGS {
        int id PK
        string mac_address
        string speaker_id
        string speaker_name
        text text
        int status
        bigint created_at
        bigint device_time
    }
    
    USERS {
        int id PK
        string username
        string password
        datetime created_at
        datetime updated_at
    }
    
    CHAT_SESSIONS {
        int id PK
        int user_id FK
        string session_id
        string status
        datetime created_at
    }
    
    CHAT_MESSAGES {
        int id PK
        int session_id FK
        string message_id
        string role
        text content
        datetime created_at
    }
    
    CHAT_STATS {
        int id PK
        int session_id FK
        int message_count
        int token_count
        datetime created_at
    }
```
