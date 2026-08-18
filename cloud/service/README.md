# SenseCraft Voice Service

[![Go Version](https://img.shields.io/badge/Go-1.23+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Supported-blue.svg)](Dockerfile)

SenseCraft Voice Service 是一个基于 Go 语言开发的现代化语音识别服务系统，提供完整的语音录制、识别、存储、智能对话和关键词匹配功能。

## 🎯 核心功能

### 🎤 语音处理
- **实时语音录制**: 支持 WebSocket 流式音频数据接收
- **语音识别**: 集成多种语音识别引擎
- **音频存储**: 支持音频文件上传和管理
- **录音管理**: 完整的录音记录查询和管理

### 🤖 智能对话
- **AI 聊天**: 集成 Dify API 和 OpenAI API
- **流式对话**: 支持实时流式响应
- **会话管理**: 聊天会话创建、管理和历史记录
- **系统提示词**: 可配置的 AI 系统提示词管理

### 🔍 关键词匹配
- **智能匹配**: 自动识别录音中的关键词
- **同义词支持**: 支持关键词同义词扩展
- **实时分析**: 异步关键词匹配处理
- **结果查询**: 多维度关键词匹配结果查询

### 🏪 门店管理
- **门店管理**: 门店信息的增删改查
- **点位管理**: 门店下点位的层级管理
- **设备管理**: 设备注册、状态监控和分配
- **关联查询**: 门店-点位-设备关联关系管理

### 👥 用户系统
- **用户认证**: 完整的用户注册、登录系统
- **权限控制**: 基于 JWT 的权限管理
- **会话管理**: 用户会话状态维护

## 🏗️ 技术架构

### 技术栈
- **语言**: Go 1.23+
- **Web框架**: Gin
- **数据库**: MySQL 8.0+
- **ORM**: GORM
- **WebSocket**: Gorilla WebSocket
- **配置管理**: Viper
- **日志**: klog
- **容器化**: Docker

### 系统架构
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   客户端层      │    │   API网关层     │    │  应用服务层     │
│                │    │                │    │                │
│ • Web客户端     │───▶│ • HTTP/HTTPS    │───▶│ • Gin引擎       │
│ • 移动端        │    │ • WebSocket     │    │ • 中间件        │
│ • IoT设备       │    │ • 健康检查      │    │ • 路由层        │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                                        │
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   数据存储层    │    │  数据访问层     │    │  业务逻辑层     │
│                │    │                │    │                │
│ • MySQL数据库   │◀───│ • DAO层         │◀───│ • Controller层  │
│ • 缓存层        │    │ • 数据模型      │    │ • Service层     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 🚀 快速开始

### 环境要求
- Go 1.23+
- MySQL 8.0+
- Docker (可选)

### 安装和运行

#### 1. 克隆项目
```bash
git clone https://github.com/YOUR-ORG/sensecraft-voice.git
cd sensecraft-voice-service
```

#### 2. 安装依赖
```bash
go mod download
```

#### 3. 配置数据库
```bash
# 创建数据库
mysql -u root -p -e "CREATE DATABASE sensecraft_voice CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 导入数据库结构
mysql -u root -p sensecraft_voice < docs/sql_chat.sql
```

#### 4. 配置文件
复制并修改配置文件：
```bash
cp config.yaml config-dev.yaml
# 编辑 config-dev.yaml 中的数据库连接信息
```

#### 5. 运行服务
```bash
# 使用 Makefile
make run

# 或直接运行
go run cmd/server.go --configfile ./config.yaml
```

#### 6. 使用 Docker
```bash
# 构建镜像
make image

# 运行容器
docker run -d -p 3008:3008 -v $(pwd)/config.yaml:/etc/sensecraft-voice/config.yaml sensecraft-voice-service
```

### 访问服务
- **API 服务**: http://localhost:3008
- **健康检查**: http://localhost:3008/healthz
- **API 文档**: http://localhost:3008/api-ref/index.html

## 📚 API 接口

### 核心模块 API

#### 录音管理 (`/api/v1/recordings`)
- `GET /stream` - WebSocket 流式录音
- `POST /` - 保存录音数据
- `GET /` - 查询录音列表
- `GET /keyword-matches` - 查询关键词匹配

#### 聊天对话 (`/api/v1/chat`)
- `POST /stream` - 流式聊天
- `POST /send` - 发送消息
- `GET /history/:session_id` - 获取聊天历史
- `GET /session/:session_id` - 获取会话信息

#### 设备管理 (`/api/v1/devices`)
- `POST /` - 创建设备
- `GET /` - 查询设备列表
- `PUT /:id/assign` - 分配设备到点位

#### 门店管理 (`/api/v1/stores`)
- `POST /` - 创建门店
- `GET /` - 查询门店列表
- `PUT /:id` - 更新门店
- `DELETE /:id` - 删除门店

#### 关键词管理 (`/api/v1/keywords`)
- `POST /` - 创建关键词
- `GET /` - 查询关键词列表
- `PUT /:id` - 更新关键词
- `DELETE /:id` - 删除关键词

### OpenAI 集成 (`/api/v2/openai`)
- `POST /chat/send` - 发送聊天消息
- `POST /chat/stream` - 流式聊天
- `GET /system-prompts` - 系统提示词管理

## 🔧 配置说明

### 主要配置项
```yaml
default:
  mode: debug                    # 运行模式
  listen: 3008                   # 服务端口
  jwt_key: CHANGE_ME             # JWT 密钥
  auto_migrate: true             # 自动迁移数据库

mysql:
  host: localhost                # 数据库主机
  user: root                     # 数据库用户
  password: password             # 数据库密码
  port: 3306                     # 数据库端口
  name: sensecraft_voice         # 数据库名

chat:
  base_url: "http://dify-api"    # Dify API 地址
  api_key: "your-api-key"        # API 密钥
  timeout: 60                    # 请求超时

openai:
  api_key: "your-openai-key"     # OpenAI API 密钥
  base_url: "https://api.openai.com/v1"
  model: "gpt-3.5-turbo"         # 默认模型
```

## 📁 项目结构

```
sensecraft-voice-service/
├── api/server/              # API 服务层
│   ├── router/             # 路由层
│   ├── middleware/         # 中间件
│   └── httputils/          # HTTP 工具
├── pkg/                    # 核心包
│   ├── controller/         # 业务逻辑层
│   ├── db/                # 数据访问层
│   ├── util/              # 工具包
│   ├── plugins/           # 插件
│   └── types/             # 类型定义
├── cmd/                   # 应用入口
├── docs/                  # 文档
├── config.yaml           # 配置文件
├── Dockerfile            # Docker 配置
└── Makefile             # 构建脚本
```

## 🛠️ 开发指南

### 构建命令
```bash
# 本地构建
make build

# Linux 构建
make build-linux

# Docker 镜像构建
make image

# 推送镜像
make push
```

### 测试
```bash
# 运行测试
go test ./...

# 运行特定模块测试
go test ./pkg/controller/chat/
```

### 代码规范
- 使用 `gofmt` 格式化代码
- 遵循 Go 语言官方编码规范
- 添加必要的注释和文档

## 🔒 安全特性

- **认证授权**: JWT Token 验证
- **权限控制**: 基于角色的访问控制
- **数据安全**: 密码加密存储
- **网络安全**: CORS 和请求限流
- **输入验证**: 完整的请求参数验证

## 📊 监控和日志

- **健康检查**: `/healthz` 端点
- **结构化日志**: JSON 格式日志输出
- **性能监控**: 请求链路追踪
- **错误处理**: 统一错误处理机制

## 🤝 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 📞 支持

如有问题或建议，请通过以下方式联系：

- 创建 [Issue](https://github.com/YOUR-ORG/sensecraft-voice/issues)
- 发送邮件到项目维护者

## 🙏 致谢

感谢所有为这个项目做出贡献的开发者和社区成员。

---

**SenseCraft Voice Service** - 让语音识别更简单、更智能！