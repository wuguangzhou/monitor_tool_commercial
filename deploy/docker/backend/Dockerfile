# 阶段1：编译Go代码（构建阶段）
FROM golang:1.25-alpine AS build-stage
WORKDIR /app
# 开启Go模块支持
ENV GO111MODULE=on
# 设置国内Go代理（解决依赖下载慢）
ENV GOPROXY=https://goproxy.cn,direct
# 复制go.mod和go.sum（路径修正：../../backend → 基于deploy/docker/找外层backend）
COPY ../../backend/go.mod ../../backend/go.sum ./
RUN go mod download
# 复制整个后端业务代码（路径修正）
COPY ../../backend ./
# 编译二进制文件（适配Linux系统，静态编译避免依赖）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o app ./cmd/server/main.go

# 阶段2：运行阶段（轻量镜像，仅几MB）
FROM alpine:3.19
# 设置时区（解决Go程序时间错误）
RUN apk add --no-cache tzdata
ENV TZ=Asia/Shanghai
# 设置工作目录
WORKDIR /app
# 从构建阶段复制二进制文件
COPY --from=build-stage /app/app /app/
# 暴露Go后端端口（和代码中一致，如8080）
EXPOSE 8080
# 启动Go程序（前台运行）
CMD ["./app"]
