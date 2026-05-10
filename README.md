# 易扣AI - 智能代码生成平台

<div align="center">

[![Go Version](https://img.shields.io/badge/Go-1.24.9-blue.svg)](https://golang.org)
[![Docker](https://img.shields.io/badge/Docker-Ready-brightgreen.svg)](https://www.docker.com)

</div>

---

## 🚀 项目介绍

**易扣AI** 是一个基于 Go 语言开发的智能代码生成平台，采用现代化的技术栈和架构设计。对于想要入门 Go 语言的程序员来说，这是一个绝佳的实战学习项目 ！该项目同时还开源了配套的[文字教程](https://github.com/FeiWuSama/yikou-ai-go-teach)，手把手教导你如何入手GoWeb项目，从萌新蜕变成大佬
![项目介绍](image/img2.png)

### 📦 架构版本

本项目提供两种架构版本，满足不同场景需求：

| 版本 | 目录 | 适用场景 | 特点 |
|------|------|----------|------|
| 单体架构 | `/` (根目录) | 快速开发、学习入门、小规模应用 | 部署简单、调试方便、资源占用少 |
| 微服务架构 | `/yikou-ai-go-microservice` | 企业级生产、高并发、团队协作 | 服务解耦、独立部署、易于扩展 |

---

## 🎯 为什么选择这个项目学习 Go？

### 1. 完整的全栈项目经验

**单体架构版本：**
- 后端：Go + Hertz + GORM + Redis + MySQL + Eino
- 前端：Vue3 + TypeScript + Ant Design Vue
- 从零到一构建完整的企业级应用

**微服务架构版本：**
- 后端：Go + Kitex + Hertz + GORM + Redis + MySQL + Eino + Nacos
- 服务网关：统一入口、路由转发
- 服务治理：服务注册发现、RPC 通信

### 2. 主流技术栈覆盖

| 技术领域 | 单体架构技术 | 微服务架构技术 | 学习价值 |
| -------- | ------------ | -------------- | -------- |
| Web框架 | Hertz (字节跳动开源) | Hertz + Kitex | 学习高性能HTTP服务和RPC服务开发 |
| AI集成 | Eino (AI工作流框架) | Eino (AI工作流框架) | 掌握AI应用开发范式 |
| 依赖注入 | Wire (Google开源) | Wire (Google开源) | 理解依赖注入设计模式 |
| ORM | GORM | GORM | 掌握数据库操作最佳实践 |
| 缓存 | Redis | Redis | 学习缓存策略和会话管理 |
| 配置管理 | Viper | Viper | 掌握多环境配置管理 |
| 服务发现 | - | Nacos | 学习微服务注册与发现 |
| RPC框架 | - | Kitex | 掌握高性能RPC通信 |
| 容器化 | Docker | Docker Compose | 学习容器编排和部署 |

### 3. 企业级架构设计

**单体架构：**
```
├── biz/           # 业务逻辑层 - 学习分层架构
├── pkg/           # 公共工具包 - 学习代码复用
└── config/        # 配置管理 - 学习工程化思维
```

**微服务架构：**
```
├── services/              # 微服务集合
│   ├── gateway/          # API网关服务
│   ├── user/             # 用户服务
│   ├── app/              # 应用服务
│   ├── ai/               # AI服务
│   └── screenshot/       # 截图服务
├── pkg/                  # 公共工具包
│   ├── commonmiddleware/ # 公共中间件
│   ├── commonenum/       # 公共枚举
│   └── errors/           # 统一错误处理
└── idl/                  # 接口定义语言 (IDL)
```

---

## ✨ 功能特性

### 🤖 AI 代码生成

- **自然语言编程**：通过对话方式描述需求，AI 自动生成代码
  ![AI 代码生成功能](image/img7.png)

### 🔄 工作流编排

- **节点组合**：灵活组合图片收集、代码生成、质量检查等节点
- **状态管理**：支持复杂的状态流转和条件分支
- **实时调试**：集成 Eino DevTools，实时调试工作流
  ![Eino DevTools](image/img8.png)

### 📦 应用管理

- **应用部署**：一键部署应用到云端
  ![应用部署](image/img4.png)
- **在线预览**：实时预览生成的应用效果
  ![应用部署](image/img5.png)

### 💬 对话历史

- **历史记录**：完整保存用户与 AI 的对话历史
- **上下文记忆**：AI 记住对话上下文，提供连贯的交互体验
- **会话管理**：支持多会话管理，隔离不同项目

### 🖼️ 图片处理

- **图片收集**：自动从网络收集相关图片素材
- **图片搜索**：集成 Pexels API，搜索高质量图片
- **图片生成**：使用 AI 生成图片素材
- **图片存储**：集成腾讯云 COS，高效存储图片资源

### 🔒 安全特性

- **企业级监控管理**
  ![应用部署](image/img1.png)
- **限流保护**：防止 API 滥用，保护系统稳定性
- **内容审核**：自动审核生成内容，过滤敏感信息
- **权限控制**：基于用户的权限管理体系
- **数据加密**：敏感数据加密存储和传输

---

## 🏗️ 项目架构

### 项目架构图

![架构图](image/yikouai-construction.drawio.png)

### 微服务架构图

```mermaid
graph TB
    subgraph "前端层"
        FE[前端应用<br/>Vue3 + TypeScript]
    end

    subgraph "网关层"
        GW[Gateway 网关<br/>Port: 8142<br/>路由转发]
    end

    subgraph "服务层"
        USR[User Service<br/>用户服务<br/>Port: 8881]
        APP[App Service<br/>应用服务<br/>Port: 8882]
        AI[AI Service<br/>AI服务<br/>Port: 8883]
        SS[Screenshot Service<br/>截图服务<br/>Port: 8884]
    end

    subgraph "基础设施层"
        NACOS[Nacos 注册中心<br/>服务发现]
        MYSQL[(MySQL<br/>数据库)]
        REDIS[(Redis<br/>缓存)]
        AIMODEL[AI Models<br/>DeepSeek]
    end

    FE -->|HTTP| GW
    GW -->|路由转发| USR
    GW -->|路由转发| APP
    GW -->|路由转发| AI

    USR -->|RPC| APP
    APP -->|RPC| USR
    APP -->|RPC| AI
    APP -->|RPC| SS

    USR --> NACOS
    APP --> NACOS
    AI --> NACOS
    SS --> NACOS

    USR --> MYSQL
    USR --> REDIS
    APP --> MYSQL
    APP --> REDIS
    AI --> AIMODEL

    style FE fill:#e1f5fe
    style GW fill:#fff3e0
    style USR fill:#f3e5f5
    style APP fill:#e8f5e9
    style AI fill:#fce4ec
    style SS fill:#fff8e1
    style NACOS fill:#e3f2fd
    style MYSQL fill:#efebe9
    style REDIS fill:#ffebee
    style AIMODEL fill:#f3e5f5
```

### 微服务说明

| 服务名称 | 端口   | 职责 | 技术栈 |
|----------|------|------|--------|
| Gateway | 8142 | API网关、路由转发、负载均衡、认证鉴权 | Hertz |
| User | 8881 | 用户管理、登录认证、会话管理 | Kitex + Hertz + GORM |
| App | 8882 | 应用管理、代码生成、部署服务 | Kitex + Hertz + GORM |
| AI | 8883 | AI对话、工作流编排、代码生成 | Kitex + Hertz + Eino |
| Screenshot | 8884 | 网页截图、图片处理 | Kitex + Chromedp |

### 服务间通信

```mermaid
graph LR
    GW[Gateway 网关]

    subgraph "服务调用关系"
        GW -->|认证鉴权| USR[User Service]
        GW -->|应用管理| APP[App Service]
        GW -->|AI对话| AI[AI Service]

        APP -->|获取用户信息| USR
        APP -->|AI对话| AI
        APP -->|网页截图| SS[Screenshot Service]
    end

    style GW fill:#fff3e0
    style USR fill:#f3e5f5
    style APP fill:#e8f5e9
    style AI fill:#fce4ec
    style SS fill:#fff8e1
```

---

## 🚀 快速开始

### 前置要求

- Go 1.24.9+
- Node.js 16+
- MySQL 8.0+
- Redis 7.0+
- Docker & Docker Compose（可选）
- Nacos 2.x（微服务版本需要）

---

## 📦 单体架构 - 快速开始

### 本地开发

#### 1. 克隆项目

```bash
git clone https://github.com/your-username/yikou-ai-go.git
cd yikou-ai-go
```

#### 2. 配置数据库

```bash
# 创建数据库
mysql -u root -p
CREATE DATABASE yikou_ai CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

# 导入表结构
mysql -u root -p yikou_ai < sql/create_table.sql
```

#### 3. 配置文件

编辑 `config/config-local.yml`，填写必要的配置：

```yaml
database:
  host: localhost
  port: 3306
  username: root
  password: your_password
  dbname: yikou_ai

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

ai:
  chat-model:
    api-key: your_deepseek_api_key
  
dashscope:
  api-key: your_dashscope_api_key
```

#### 4. 安装依赖

```bash
# 后端依赖
go mod download

# 前端依赖
cd yikou-ai-feiwu-front
npm install
```

#### 5. 启动服务

```bash
# 启动后端服务
go run main.go

# 启动前端服务（新终端）
cd yikou-ai-feiwu-front
npm run dev
```

访问 http://localhost:5173 即可操作。

### Docker 部署

```bash
# 构建镜像
docker build -t yikou-ai-go:latest .

# 运行容器
docker run -d \
  --name yikou-ai-go \
  -p 8123:8123 \
  yikou-ai-go:latest
```

---

## 🔧 微服务架构 - 快速开始

### 本地开发

#### 1. 克隆项目

```bash
git clone https://github.com/your-username/yikou-ai-go.git
cd yikou-ai-go/yikou-ai-go-microservice
```

#### 2. 启动基础设施

```bash
# 启动 Nacos、MySQL、 Redis
docker-compose up -d nacos mysql redis

# 等待服务启动完成（约30秒）
# 访问 Nacos 控制台: http://localhost:8848/nacos (账号密码: nacos/nacos)
```

#### 3. 初始化数据库

```bash
# 导入表结构
mysql -h 127.0.0.1 -P 3306 -u root -p < init/sql/init.sql
```

#### 4. 配置文件

各服务配置文件位于 `services/{service}/config/config.yml`：

```yaml
# 示例：用户服务配置
database:
  host: localhost
  port: 3306
  username: root
  password: your_password
  dbname: yikou_ai

redis:
  host: localhost
  port: 6379

nacos:
  server_addr: 127.0.0.1:8848
  namespace: public
  group: DEFAULT_GROUP
```

#### 5. 启动微服务

```bash
# 方式一：依次启动各服务（开发调试）
# 终端1 - 启动用户服务
cd services/user && go run main.go

# 终端2 - 启动应用服务
cd services/app && go run main.go

# 终端3 - 启动AI服务
cd services/ai && go run main.go

# 终端4 - 启动截图服务
cd services/screenshot && go run main.go

# 终端5 - 启动网关
cd services/gateway && go run main.go

# 方式二：使用 Docker Compose 一键启动（生产部署）
docker-compose up -d
```

#### 6. 启动前端

```bash
cd yikou-ai-feiwu-front
npm install
npm run dev
```

访问 http://localhost:5173 即可操作。

### Docker Compose 部署（微服务）

```bash
# 一键启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f gateway

# 停止所有服务
docker-compose down
```

---

## 📁 项目结构

### 单体架构项目结构

```
yikou-ai-go/
├── biz/                    # 业务逻辑层
│   ├── ai/                # AI 相关功能
│   │   ├── agent/         # AI Agent 实现
│   │   ├── aimodel/       # AI 模型定义
│   │   ├── aitools/       # AI 工具集成
│   │   └── llm/           # LLM 模型封装
│   ├── core/              # 核心业务逻辑
│   │   ├── messagehandler/ # 消息处理器
│   │   ├── parser/        # 代码解析器
│   │   └── saver/         # 文件保存器
│   ├── dal/               # 数据访问层
│   ├── graph/             # 工作流图定义
│   │   ├── node/          # 工作流节点
│   │   └── state/         # 工作流状态
│   ├── handler/           # HTTP 处理器
│   ├── manager/           # 第三方客户端管理器
│   ├── middleware/        # 中间件
│   ├── model/             # 数据模型
│   ├── router/            # 路由定义
│   ├── logic/             # 业务服务具体实现
|   └── service/           # 业务接口定义
├── config/                # 配置文件
│   ├── config.go          # 配置加载逻辑
│   ├── config.yml         # 基础配置
│   └── config-prod.yml    # 生产环境配置
├── docs/                  # Swagger 文档
├── pkg/                   # 公共工具包
│   ├── constants/         # 常量定义
│   ├── errors/            # 错误处理
│   ├── myfile/            # 文件操作
│   ├── myutils/           # 工具函数
│   ├── random/            # 随机数生成
│   └── snowflake/         # 雪花算法
├── sql/                   # SQL 脚本
├── wire/                  # 依赖注入
├── grafana/               # Grafana看板配置
├── yikou-ai-feiwu-front/  # 前端项目
├── Dockerfile             # Docker 构建文件
├── prometheus.yml         # Prometheus启动配置文件
├── go.mod                 # Go 模块定义
├── go.sum                 # Go 依赖锁定
└── main.go                # 应用入口
```

### 微服务架构项目结构

```
yikou-ai-go-microservice/
├── services/                      # 微服务集合
│   ├── gateway/                  # API 网关服务
│   │   ├── config/               # 网关配置
│   │   ├── proxy/                # 反向代理逻辑
│   │   └── main.go               # 网关入口
│   │
│   ├── user/                     # 用户服务
│   │   ├── config/               # 用户服务配置
│   │   ├── handler/              # HTTP 处理器
│   │   ├── logic/                # 业务逻辑
│   │   ├── dal/                  # 数据访问层
│   │   ├── middleware/           # 服务内中间件
│   │   ├── router/               # 路由定义
│   │   ├── kitex_gen/            # Kitex 生成代码
│   │   └── main.go               # 服务入口
│   │
│   ├── app/                      # 应用服务
│   │   ├── config/               # 应用服务配置
│   │   ├── handler/              # HTTP 处理器
│   │   ├── logic/                # 业务逻辑
│   │   ├── dal/                  # 数据访问层
│   │   ├── middleware/           # 服务内中间件
│   │   ├── router/               # 路由定义
│   │   ├── kitex_gen/            # Kitex 生成代码
│   │   └── main.go               # 服务入口
│   │
│   ├── ai/                       # AI 服务
│   │   ├── config/               # AI 服务配置
│   │   ├── handler/              # HTTP 处理器
│   │   ├── agent/                # AI Agent 实现
│   │   ├── graph/                # 工作流图定义
│   │   ├── kitex_gen/            # Kitex 生成代码
│   │   └── main.go               # 服务入口
│   │
│   └── screenshot/               # 截图服务
│       ├── config/               # 截图服务配置
│       ├── handler/              # HTTP 处理器
│       ├── kitex_gen/            # Kitex 生成代码
│       └── main.go               # 服务入口
│
├── pkg/                          # 公共工具包
│   ├── commonmiddleware/         # 公共中间件 (认证、限流等)
│   ├── commonenum/               # 公共枚举定义
│   ├── constants/                # 常量定义
│   ├── errors/                   # 统一错误处理
│   └── utils/                    # 工具函数
│
├── init/                         # 初始化脚本
│   └── sql/                      # 数据库初始化脚本
│
├── docker-compose.yml            # Docker Compose 配置
└── go.mod                        # Go 模块定义
```

---

## 🔧 配置说明

### 多环境配置

项目支持多环境配置，通过 `-env` 参数切换：

```bash
# 使用本地配置
go run main.go -env=local

# 使用生产配置
go run main.go -env=prod
```

### 核心配置项

```yaml
server:
  port: 8123              # 服务端口
  context-path: /api      # API 上下文路径

database:
  host: localhost         # 数据库地址
  port: 3306             # 数据库端口
  username: root         # 数据库用户名
  password: password     # 数据库密码
  dbname: yikou_ai       # 数据库名称

redis:
  host: localhost        # Redis 地址
  port: 6379            # Redis 端口
  password: ""          # Redis 密码
  db: 0                 # Redis 数据库

ai:
  chat-model:
    base-url: https://dashscope.aliyuncs.com/compatible-mode/v1
    api-key: your_api_key
    model-name: deepseek-v3.2
    memory-store: redis
    memory-ttl: 3600

cos:
  host: your_cos_host
  secret-id: your_secret_id
  secret-key: your_secret_key
  region: your_region
  bucket: your_bucket
```

### 微服务额外配置

```yaml
# Nacos 服务发现配置
nacos:
  server_addr: 127.0.0.1:8848
  namespace: public
  group: DEFAULT_GROUP
  service_name: user-service

# RPC 服务配置
rpc:
  user_service: 127.0.0.1:8881
  app_service: 127.0.0.1:8882
  ai_service: 127.0.0.1:8883
  screenshot_service: 127.0.0.1:8884
```

---

## 📚 API 文档

### 单体架构

启动服务后，访问以下地址查看 API 文档：

- Swagger UI: http://localhost:8123/swagger/index.html

### 主要 API 端点

| 方法 | 路径                        | 说明            | 服务 |
| ---- | --------------------------- | --------------- | ---- |
| POST | /api/user/register          | 用户注册        | User |
| POST | /api/user/login             | 用户登录        | User |
| GET  | /api/user/get/login         | 获取登录用户    | User |
| POST | /api/user/logout            | 用户登出        | User |
| POST | /api/app/add                | 创建应用        | App |
| POST | /api/app/update             | 更新应用        | App |
| POST | /api/app/delete             | 删除应用        | App |
| GET  | /api/app/get                | 获取应用详情    | App |
| POST | /api/app/list               | 获取应用列表    | App |
| POST | /api/app/deploy             | 部署应用        | App |
| GET  | /api/app/chat/gen/code      | AI 对话生成代码 | App |
| POST | /api/workflow/execute       | 执行工作流      | AI |
| GET  | /api/chatHistory/app/:appId | 获取聊天历史    | App |

---

## 🙏 致谢

感谢以下开源项目和组织：

- [CloudWeGo](https://www.cloudwego.io/) - 提供强大的 Hertz、Kitex 和 Eino 框架
- [Ant Design](https://ant.design/) - 提供精美的 UI 组件库
- [Vue.js](https://vuejs.org/) - 提供优秀的前端框架
- [Nacos](https://nacos.io/) - 提供服务发现和配置管理

---

## 📞 联系方式

- 个人Github主页: https://github.com/FeiWuSama
- 邮箱: 1825578184@qq.com

---

<div align="center">

**⭐ 如果这个项目对您有帮助，请给我们一个 Star！⭐**

Made with ❤️ by FeiWuSama

</div>
