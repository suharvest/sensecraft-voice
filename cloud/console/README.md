# SenseCraft Voice Web

SenseCraft Voice 语音管理系统的前端应用，基于 React + TypeScript + Ant Design + Tailwind CSS 构建。

## 🚀 技术栈

- **前端框架**: React 18 + TypeScript
- **UI组件库**: Ant Design 5.x
- **样式框架**: Tailwind CSS 3.x
- **状态管理**: Zustand
- **路由管理**: React Router 6
- **HTTP客户端**: Axios
- **构建工具**: Vite
- **包管理**: npm

## 🎨 设计特色

- **主色调**: #8CC020 (自然绿色)
- **响应式设计**: 支持桌面端、平板、手机
- **现代化UI**: 卡片式布局、渐变背景、阴影效果
- **统一风格**: 结合 Ant Design 组件和 Tailwind CSS 样式

## 📁 项目结构

```
src/
├── components/          # 通用组件
│   └── Layout/         # 布局组件
├── pages/              # 页面组件
│   ├── Auth/           # 认证页面
│   ├── Dashboard/      # 仪表盘
│   └── Recordings/     # 录音管理
├── services/           # API服务
├── stores/             # 状态管理
├── types/              # TypeScript类型定义
├── utils/              # 工具函数
├── theme/              # 主题配置
├── App.tsx             # 主应用组件
├── main.tsx            # 应用入口
└── index.css           # 全局样式
```

## 🚀 快速开始

### 1. 安装依赖

```bash
npm install
```

### 2. 启动开发服务器

```bash
npm run dev
```

应用将在 http://localhost:3000 启动

### 3. 构建生产版本

```bash
npm run build
```

### 4. 预览生产版本

```bash
npm run preview
```

## 🔧 配置说明

### 环境配置

- **开发端口**: 3000
- **API代理**: 自动代理到 http://localhost:8080
- **浏览器**: 自动打开默认浏览器

### 主题配置

主色调 #8CC020 已配置到：
- Tailwind CSS 自定义颜色
- Ant Design 主题变量
- 全局CSS变量

## 📱 功能模块

### 1. 用户认证
- 用户登录/登出
- 路由保护
- 状态持久化

### 2. 仪表盘
- 数据统计展示
- 增长趋势图表
- 最近活动记录

### 3. 录音管理
- 录音记录列表
- 分页查询
- MAC地址过滤
- 数据导出
- 批量操作

### 4. 门店管理 (待实现)
- 门店CRUD操作
- 层级管理

### 5. 点位管理 (待实现)
- 点位CRUD操作
- 设备分配

### 6. 用户管理 (待实现)
- 用户CRUD操作
- 权限管理

## 🎯 API接口

系统已集成以下API服务：

- **录音管理**: `/api/v1/recordings`
- **门店管理**: `/api/v1/stores`
- **点位管理**: `/api/v1/locations`
- **用户管理**: `/api/v1/users`

## 🎨 自定义样式

### Tailwind CSS 扩展

```css
/* 主色调系列 */
.primary-50   /* #f7fbf0 */
.primary-500  /* #8CC020 */
.primary-600  /* #7aad1c */
.primary-700  /* #689a18 */

/* 自定义组件类 */
.btn-primary    /* 主要按钮样式 */
.card-primary   /* 卡片样式 */
.input-primary  /* 输入框样式 */
.table-primary  /* 表格样式 */
```

### Ant Design 主题定制

- 主色调: #8CC020
- 圆角: 8px
- 阴影: 自定义绿色阴影
- 组件样式: 统一覆盖

## 🔒 安全特性

- 路由保护
- Token认证
- 请求拦截器
- 错误处理

## 📱 响应式支持

- **桌面端**: 完整功能展示
- **平板端**: 适配中等屏幕
- **手机端**: 移动端优化

## 🚀 部署说明

### 构建

```bash
npm run build
```

### 部署

将 `dist` 目录部署到Web服务器即可。

## 🤝 开发规范

### 代码风格

- 使用 TypeScript 严格模式
- 组件使用函数式组件 + Hooks
- 状态管理使用 Zustand
- 样式优先使用 Tailwind CSS

### 文件命名

- 组件文件: PascalCase (如 `RecordingsPage.tsx`)
- 工具文件: camelCase (如 `dateUtils.ts`)
- 类型文件: camelCase (如 `index.ts`)

## 📝 更新日志

### v1.0.0 (2024-01-XX)
- ✅ 项目基础架构搭建
- ✅ 用户认证系统
- ✅ 仪表盘页面
- ✅ 录音管理页面
- ✅ 响应式布局
- ✅ 主题系统

## 📄 许可证

MIT License

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

---

**SenseCraft Voice Web** - 让语音管理更简单 🎵
