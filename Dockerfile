# 后端 Dockerfile
FROM golang:1.24.9-alpine AS builder

RUN apk add --no-cache \
    git \
    gcc \
    g++ \
    make \

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go run github.com/google/wire/cmd/wire ./wire
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o yikou-ai-go ./main.go

FROM alpine:latest

RUN apk add --no-cache \
    chromium \
    chromium-chromedriver \
    ttf-freefont \
    ca-certificates \
    libstdc++ \
    libgcc

ENV CHROME_BIN=/usr/bin/chromium-browser
ENV CHROME_DRIVER=/usr/bin/chromedriver

WORKDIR /app

COPY --from=builder /app/yikou-ai-go /app/
COPY --from=builder /app/config/config.yml /app/config/
COPY --from=builder /app/config/config-prod.yml /app/config/

EXPOSE 8888

CMD ["/app/yikou-ai-go", "-env=prod"]
