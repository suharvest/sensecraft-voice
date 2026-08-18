# Chat 模块配置说明

## 📋 配置文件结构

在 `config.yaml` 中添加以下配置：

```yaml
# 聊天服务配置
chat:
  # Dify API 基础地址
  base_url: "https://dify.example.com"
  # API 密钥
  api_key: "your_dify_api_key_here"
  # 请求超时时间（秒）
  timeout: 60
  # 是否启用调试模式
  enable_debug: false
```

## 🔑 配置项说明

### base_url
- **类型**: string
- **必填**: 是
- **说明**: Dify API 的基础URL地址
- **示例**: `"https://dify.example.com"`

### api_key
- **类型**: string
- **必填**: 是
- **说明**: Dify API 的访问密钥
- **获取方式**: 登录Dify平台，在API设置中获取
- **示例**: `"sk-xxxxxxxxxxxxxxxxxxxxxxxx"`

### timeout
- **类型**: int
- **必填**: 否
- **默认值**: 60
- **说明**: HTTP请求超时时间，单位为秒
- **建议值**: 30-120秒

### enable_debug
- **类型**: bool
- **必填**: 否
- **默认值**: false
- **说明**: 是否启用调试模式，启用后会输出详细的请求日志

## 🚀 配置示例

### 1. 基础配置
```yaml
chat:
  base_url: "https://dify.example.com"
  api_key: "sk-1234567890abcdef"
  timeout: 60
  enable_debug: false
```

### 2. 开发环境配置
```yaml
chat:
  base_url: "http://localhost:5001"
  api_key: "dev_api_key"
  timeout: 30
  enable_debug: true
```

### 3. 生产环境配置
```yaml
chat:
  base_url: "https://api.dify.ai"
  api_key: "sk-prod-xxxxxxxxxxxxxxxxxxxxxxxx"
  timeout: 120
  enable_debug: false
```

## 🔒 安全配置建议

### 1. API密钥管理
- 不要在代码中硬编码API密钥
- 使用环境变量或配置文件管理敏感信息
- 定期轮换API密钥

### 2. 网络配置
- 生产环境使用HTTPS
- 配置适当的防火墙规则
- 限制API访问来源

### 3. 超时配置
- 根据网络环境调整超时时间
- 避免设置过短的超时时间
- 考虑添加重试机制

## 🌍 环境变量支持

你也可以通过环境变量来配置chat模块：

```bash
export CHAT_BASE_URL="https://dify.example.com"
export CHAT_API_KEY="your_api_key"
export CHAT_TIMEOUT="60"
export CHAT_ENABLE_DEBUG="false"
```

## ⚠️ 注意事项

1. **API密钥安全**: 确保API密钥的安全性，不要提交到版本控制系统
2. **网络访问**: 确保服务器能够访问配置的Dify API地址
3. **超时设置**: 根据实际网络情况调整超时时间
4. **调试模式**: 生产环境建议关闭调试模式
5. **配置验证**: 启动服务时会自动验证配置的有效性

## 🔍 配置验证

启动服务后，可以通过以下方式验证配置：

1. 检查服务日志，确认chat模块配置加载成功
2. 尝试调用chat接口，验证API连接正常
3. 检查数据库连接，确认chat相关表创建成功

## 📞 故障排除

### 常见问题

1. **API密钥无效**
   - 检查API密钥是否正确
   - 确认API密钥是否过期
   - 验证API密钥权限

2. **网络连接失败**
   - 检查网络连接
   - 验证防火墙设置
   - 确认API地址可访问

3. **超时错误**
   - 增加超时时间
   - 检查网络延迟
   - 考虑使用CDN

### 调试步骤

1. 启用调试模式：`enable_debug: true`
2. 检查服务日志
3. 验证配置文件格式
4. 测试网络连接
5. 检查API密钥有效性
