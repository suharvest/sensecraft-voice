# 多阶段构建 Dockerfile
# 阶段1: 构建阶段
FROM --platform=linux/amd64 node:18-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 package.json 和 package-lock.json (利用Docker缓存层)
COPY package*.json ./

# 安装依赖
RUN npm ci --only=production=false && npm cache clean --force

# 复制源代码
COPY . .

# 复制生产环境配置文件
COPY .env.production .env.production

# 构建应用 (生产环境模式)
RUN npm run build:prod

# 阶段2: 生产运行阶段
FROM --platform=linux/amd64 node:18-alpine AS production

# 安装必要的工具
RUN apk add --no-cache curl

# 设置工作目录
WORKDIR /app

# 创建非 root 用户
RUN addgroup -g 1001 -S nodejs && \
    adduser -S nodejs -u 1001

# 复制构建产物和必要文件
COPY --from=builder /app/dist ./dist
COPY --from=builder /app/package*.json ./
COPY --from=builder /app/.env.production ./.env.production

# 全局安装serve用于托管静态文件
RUN npm install -g serve

# 设置正确的权限
RUN chown -R nodejs:nodejs /app

# 切换到非 root 用户
USER nodejs

# 暴露端口
EXPOSE 3000

# 健康检查 (检查主页是否可访问)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:3000/ || exit 1

# 启动应用 (使用serve托管静态文件)
CMD ["serve", "-s", "dist", "-l", "3000"]
