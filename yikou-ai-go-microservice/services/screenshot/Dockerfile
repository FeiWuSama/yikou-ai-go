# 构建阶段
FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-checklinkname=0" -o /main .

# 运行阶段
FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /main /app/main

ENV TZ=Asia/Shanghai

EXPOSE 8081 9091

CMD ["./main"]
