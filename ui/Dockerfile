# --- 构建阶段 ---
FROM node:20-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制package文件
COPY package.json package-lock.json ./

# 安装依赖（优化缓存层）
RUN npm ci --only=production --no-audit --no-fund && \
    npm cache clean --force

# 复制源代码
COPY . .

# 构建应用
RUN npm run build && \
    rm -rf node_modules && \
    rm -rf /app/.npm

# --- 运行阶段 ---
FROM nginx:1.27-alpine

# 安装必要的工具
RUN apk add --no-cache curl && \
    rm -rf /var/cache/apk/*

# 复制构建产物
COPY --from=builder /app/dist /usr/share/nginx/html

# 复制nginx配置
COPY ./nginx.conf /etc/nginx/conf.d/default.conf

# 设置权限
RUN chown -R nginx:nginx /usr/share/nginx/html && \
    chown -R nginx:nginx /var/cache/nginx && \
    chown -R nginx:nginx /var/log/nginx && \
    chown -R nginx:nginx /etc/nginx/conf.d && \
    touch /var/run/nginx.pid && \
    chown -R nginx:nginx /var/run/nginx.pid

# 切换到非root用户
USER nginx

# 暴露端口
EXPOSE 80

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:80/ || exit 1

# 启动nginx
CMD ["nginx", "-g", "daemon off;"]