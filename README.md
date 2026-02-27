# LLMGate - 企业内部 LLM 管理平台

LLMGate 是一个为企业内部提供统一 LLM 服务入口的管理平台，实现用户管理、权限控制、配额计费和审计追踪。

## 核心功能

- **用户管理**：JWT 认证、角色管理（管理员/经理/普通用户）
- **API Key 管理**：用户自助创建/删除 API Key
- **模型管理**：支持多个后端 LLM 服务器的负载均衡
- **配额控制**：速率限制、Token 配额、并发控制
- **审计日志**：完整的请求日志（7天自动清理）

## 技术栈

- **后端**: Go + Gin
- **数据库**: PostgreSQL
- **缓存**: Redis
- **部署**: Docker + Docker Compose

## 快速开始

### 使用 Docker Compose

```bash
# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f server
```

### 本地开发

```bash
# 安装依赖
go mod tidy

# 配置数据库
cp config.yaml config.local.yaml
# 编辑 config.local.yaml 设置数据库连接

# 运行服务
go run cmd/server/main.go
```

### 默认管理员账号

- **邮箱**: admin@llmgate.local
- **密码**: admin123

## API 接口

### 认证接口

```bash
# 登录
POST /api/v1/auth/login
{
  "email": "user@example.com",
  "password": "password"
}

# 注册
POST /api/v1/auth/register
{
  "email": "user@example.com",
  "password": "password",
  "name": "User Name"
}
```

### API Key 管理

```bash
# 创建 API Key
POST /api/v1/user/keys
{
  "name": "开发测试"
}

# 列出 API Keys
GET /api/v1/user/keys

# 删除 API Key
DELETE /api/v1/user/keys/:id
```

### LLM 代理接口（OpenAI 兼容）

```bash
# 列出模型
GET /v1/models

# 聊天补全
POST /v1/chat/completions
Authorization: Bearer your-api-key
{
  "model": "llama3-70b",
  "messages": [{"role": "user", "content": "Hello"}]
}
```

### 管理接口

```bash
# 用户管理
GET    /api/v1/admin/users
POST   /api/v1/admin/users
PUT    /api/v1/admin/users/:id
DELETE /api/v1/admin/users/:id

# 模型管理
GET    /api/v1/admin/models
POST   /api/v1/admin/models
PUT    /api/v1/admin/models/:id
DELETE /api/v1/admin/models/:id

# 配额策略
GET    /api/v1/admin/policies
POST   /api/v1/admin/policies
PUT    /api/v1/admin/policies/:name
DELETE /api/v1/admin/policies/:name
```

## 配置说明

### config.yaml

```yaml
server:
  port: 8080
  mode: "release"  # debug 或 release

database:
  host: "localhost"
  port: 5432
  user: "llmgate"
  password: "llmgate_pass"
  dbname: "llmgate"

redis:
  host: "localhost"
  port: 6379

jwt:
  secret: "your-jwt-secret-key"
  expire_hours: 24

# LLM 后端配置
models:
  - id: "llama3-70b"
    name: "Llama 3 70B"
    backend: "http://llm-server-1:8000"
    enabled: true
    weight: 1
  - id: "qwen-72b"
    name: "Qwen 72B"
    backend: "http://llm-server-2:8000"
    enabled: true
    weight: 1

# 配额策略
quota_policies:
  - name: "default"
    rate_limit: 60          # 每分钟请求数
    rate_limit_window: 60   # 窗口秒数
    token_quota_daily: 100000
    models: ["llama3-70b", "qwen-72b"]
```

## 项目结构

```
llmgate/
├── cmd/server/           # 主入口
├── internal/
│   ├── auth/            # JWT 认证
│   ├── user/            # 用户管理
│   ├── apikey/          # API Key 管理
│   ├── model/           # 模型管理
│   ├── quota/           # 配额检查
│   ├── proxy/           # LLM 代理
│   ├── usage/           # 使用记录
│   ├── db/              # 数据库连接
│   ├── middleware/      # 中间件
│   ├── config/          # 配置管理
│   └── models/          # 数据模型
├── migrations/          # 数据库迁移
├── config.yaml          # 配置文件
├── docker-compose.yml   # Docker Compose 配置
└── README.md
```

## 数据库表结构

- **users**: 用户表
- **api_keys**: API Key 表
- **models**: 模型配置表
- **quota_policies**: 配额策略表
- **usage_records**: 使用记录表（分区表，7天保留）
- **quota_usage_daily**: 每日配额使用统计

## 开发计划

- [x] 用户认证（JWT）
- [x] API Key 管理
- [x] 模型管理
- [x] 配额检查
- [x] 负载均衡
- [x] 使用记录
- [ ] Web 管理界面（React）
- [ ] 企业 SSO 集成
- [ ] 实时监控仪表盘

## License

MIT
