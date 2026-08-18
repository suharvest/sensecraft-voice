# 环境配置说明

## 📁 环境配置文件

项目支持多环境配置，通过不同的环境文件来管理不同环境的配置。

### 环境文件

- `env.development` - 开发环境配置
- `env.test` - 测试环境配置  
- `env.production` - 生产环境配置

### 环境变量说明

| 变量名 | 说明 | 开发环境 | 测试环境 | 生产环境 |
|--------|------|----------|----------|----------|
| `VITE_API_BASE_URL` | API基础URL | http://localhost:8080 | https://test-api.sensecraft-voice.com | https://api.sensecraft-voice.com |
| `VITE_APP_TITLE` | 应用标题 | SenseCraft Voice | SenseCraft Voice (Test) | SenseCraft Voice |
| `VITE_APP_ENV` | 环境标识 | development | test | production |
| `VITE_API_TIMEOUT` | API超时时间(ms) | 10000 | 10000 | 15000 |

## 🚀 启动命令

### 开发环境
```bash
npm run dev
# 或
npm run dev:test
npm run dev:prod
```

### 构建命令
```bash
# 开发环境构建
npm run build

# 测试环境构建
npm run build:test

# 生产环境构建
npm run build:prod
```

## 🔧 配置使用

### 在代码中使用环境配置

```typescript
import { envConfig, isDev, isProd } from '@/config/env';

// 使用配置
console.log('API Base URL:', envConfig.API_BASE_URL);
console.log('App Title:', envConfig.APP_TITLE);

// 环境判断
if (isDev) {
  console.log('开发环境');
}

if (isProd) {
  console.log('生产环境');
}
```

### API服务自动使用环境配置

所有API服务都会自动使用环境配置中的 `API_BASE_URL` 和 `TIMEOUT` 设置。

## 📝 注意事项

1. **环境变量前缀**: 所有环境变量必须以 `VITE_` 开头
2. **配置文件**: 环境配置文件需要放在项目根目录
3. **构建时替换**: 生产构建时会自动替换环境变量
4. **安全性**: 不要在环境变量中存储敏感信息

## 🔄 切换环境

### 开发时切换
```bash
# 切换到测试环境
npm run dev:test

# 切换到生产环境
npm run dev:prod
```

### 部署时切换
```bash
# 构建测试版本
npm run build:test

# 构建生产版本
npm run build:prod
```

## 🛠️ 自定义环境

如果需要添加新的环境，可以：

1. 创建新的环境文件，如 `env.staging`
2. 在 `package.json` 中添加对应的脚本
3. 在 `src/config/env.ts` 中添加环境判断逻辑

```json
{
  "scripts": {
    "dev:staging": "vite --mode staging",
    "build:staging": "tsc && vite build --mode staging"
  }
}
```

