# 阶段1：构建Vue3静态文件（适配context=项目根目录）
FROM node:18-alpine AS build-stage
WORKDIR /app
# 复制前端依赖：基于项目根目录，路径是 ./frontend/package*.json
COPY ./frontend/package*.json ./
# 国内源加速安装依赖
RUN npm install --registry=https://registry.npmmirror.com
# 复制整个前端代码：基于项目根目录，路径是 ./frontend/
COPY ./frontend/ ./

# 核心修复：强制赋予vite执行权限（关键！）
RUN chmod +x node_modules/.bin/vite || true

# 打包生成dist目录
RUN npm run build

# 阶段2：Nginx托管静态文件（适配context=项目根目录）
FROM nginx:alpine AS production-stage
# 清空默认nginx配置
RUN rm -f /etc/nginx/conf.d/default.conf
# 复制nginx.conf：基于项目根目录，路径是 ./deploy/docker/frontend/nginx.conf
COPY ./deploy/docker/frontend/nginx.conf /etc/nginx/conf.d/default.conf
# 复制dist文件（从构建阶段复制，路径不变）
COPY --from=build-stage /app/dist /usr/share/nginx/html
# 暴露80端口
EXPOSE 80
# 启动Nginx
CMD ["nginx", "-g", "daemon off;"]
