# 使用多阶段构建
FROM golang:1.21-alpine AS builder

# 安装必要的工具
RUN apk add --no-cache git

# 设置工作目录
WORKDIR /build

# 设置 Go 代理（使用中国镜像源）
ENV GOPROXY=https://goproxy.cn,direct

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o vllm-show .

# 运行阶段
FROM alpine:latest

# 安装ca证书（用于HTTPS请求）
RUN apk add --no-cache ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# 创建非root用户
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /build/vllm-show .

# 创建上传目录
RUN mkdir -p /app/uploads && \
    chown -R appuser:appuser /app /app/uploads

# 切换到非root用户
USER appuser

# 暴露端口
# 文件服务器端口（默认8080）
EXPOSE 8080
# API服务器端口（默认8081）
EXPOSE 8081

# 运行服务
CMD ["./vllm-show"]

