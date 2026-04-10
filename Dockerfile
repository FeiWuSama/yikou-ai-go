# 构建阶段
FROM golang:1.24.9-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache \
    git \
    gcc \
    g++ \
    make \
    ca-certificates \
    nodejs \
    npm

# 设置工作目录
WORKDIR /app

# 复制依赖文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源码
COPY . .

# 构建前端
RUN ./build-frontend.sh

# 生成 wire 依赖注入代码
RUN go run github.com/google/wire/cmd/wire ./wire

# 编译应用（使用项目名作为可执行文件名）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o yikou-ai-go ./main.go

# 运行阶段
FROM alpine:latest

# 安装 Chrome 和依赖
RUN apk add --no-cache \
    chromium \
    chromium-chromedriver \
    ttf-freefont \
    ca-certificates \
    libstdc++ \
    libgcc

# 设置 Chrome 环境变量
ENV CHROME_BIN=/usr/bin/chromium-browser
ENV CHROME_DRIVER=/usr/bin/chromedriver

# 创建应用目录
WORKDIR /app

# 复制编译产物
COPY --from=builder /app/yikou-ai-go /app/
# 复制配置文件
COPY --from=builder /app/config/config.yml /app/config/
COPY --from=builder /app/config/config-prod.yml /app/config/
# 复制前端文件
COPY --from=builder /app/yikou-ai-feiwu-front/dist /app/yikou-ai-feiwu-front/dist/

# 暴露端口
EXPOSE 8888

# 启动应用（使用生产环境配置）
CMD ["/app/yikou-ai-go", "-env=prod"]
