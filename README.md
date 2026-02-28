# LLMGate - 企业内部 LLM 管理平台

[![Go Version](https://img.shields.io/badge/Go-1.22+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

LLMGate 是一个为企业内部提供统一 LLM 服务入口的管理平台，实现用户管理、权限控制、配额计费和审计追踪。

## 核心功能

- **用户管理**：JWT 认证、角色管理（管理员/经理/普通用户）
- **API Key 管理**：用户自助创建/删除 API Key，支持过期时间和模型限制
- **模型管理**：支持多个后端 LLM 服务器的负载均衡
- **配额控制**：速率限制、Token 配额、并发控制
- **审计日志**：完整的请求日志（7天自动清理）
- **OpenAI 兼容**：提供与 OpenAI API 兼容的接口
- **单文件部署**：前端资源嵌入二进制，仅需一个可执行文件

## 技术栈

- **后端**: Go 1.22+ + Gin
- **数据库**: SQLite（单文件，零配置）
- **缓存**: 内存（内置实现）
- **前端**: React + TypeScript + Ant Design
- **部署**: 单二进制文件，无需 Docker

## 快速开始

### 环境要求

- Go 1.22 或更高版本（仅编译时需要）
- Node.js 18+（仅编译前端时需要）

### 构建

```bash
# 完整构建（前端 + 后端）
make build

# 构建多平台发布包
make release
```

构建产物：
- `llmgate` - 单个可执行文件（包含前端资源，约 12MB）
- `config.yaml` - 配置文件

### 部署

```bash
# 仅需两个文件即可部署
./llmgate
```

访问 http://localhost:8080 即可使用 Web 管理界面。

### 开发模式

```bash
# 同时启动前端开发服务器和后端服务
make dev

# 前端: http://localhost:5173
# 后端: http://localhost:8080
```

### 默认管理员账号

- **邮箱**: admin@llmgate.local
- **密码**: admin123

**注意**：首次登录后请立即修改默认密码。

## 配置说明

### config.yaml

```yaml
server:
  port: 8080              # 服务端口
  mode: "release"         # debug 或 release

database:
  path: "llmgate.db"      # SQLite 数据库文件路径

logs:
  path: "./logs"          # 日志目录
  retention_days: 7       # 日志保留天数

jwt:
  secret: "your-jwt-secret-change-in-production"
  expire_hours: 24

admin:
  default_email: "admin@llmgate.local"
  default_password: "admin123"

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
    rate_limit: 60              # 每分钟请求数
    rate_limit_window: 60       # 窗口秒数
    token_quota_daily: 100000   # 每日 Token 上限
    models: ["*"]               # "*" 表示所有模型
```

## API 接口

### 认证接口

```bash
# 登录
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password"
}

# 响应
{
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": "...",
      "email": "user@example.com",
      "name": "User Name",
      "role": "user"
    }
  }
}
```

### API Key 管理

```bash
# 创建 API Key（需登录）
POST /api/v1/user/keys
Authorization: Bearer <jwt-token>
Content-Type: application/json

{
  "name": "开发测试",
  "models": ["llama3-70b"],
  "expires_at": "2024-12-31T23:59:59Z"
}

# 列出 API Keys
GET /api/v1/user/keys
Authorization: Bearer <jwt-token>

# 删除 API Key
DELETE /api/v1/user/keys/:id
Authorization: Bearer <jwt-token>
```

### LLM 代理接口（OpenAI 兼容）

```bash
# 列出可用模型
GET /v1/models
Authorization: Bearer <api-key>

# 聊天补全
POST /v1/chat/completions
Authorization: Bearer <api-key>
Content-Type: application/json

{
  "model": "llama3-70b",
  "messages": [
    {"role": "user", "content": "Hello, how are you?"}
  ],
  "temperature": 0.7,
  "max_tokens": 1000
}
```

### 管理接口

```bash
# 用户管理
GET    /api/v1/admin/users          # 列出用户
POST   /api/v1/admin/users          # 创建用户
PUT    /api/v1/admin/users/:id      # 更新用户
DELETE /api/v1/admin/users/:id      # 删除用户

# 模型管理
GET    /api/v1/admin/models
POST   /api/v1/admin/models
PUT    /api/v1/admin/models/:id
DELETE /api/v1/admin/models/:id

# 配额策略管理
GET    /api/v1/admin/policies
POST   /api/v1/admin/policies
PUT    /api/v1/admin/policies/:name
DELETE /api/v1/admin/policies/:name
```

## 项目结构

```
llmgate/
├── cmd/
│   └── server/              # 主程序入口
├── internal/                # 内部包
│   ├── apikey/             # API Key 管理
│   ├── auth/               # JWT 认证
│   ├── config/             # 配置管理
│   ├── db/                 # 数据库（SQLite）
│   ├── logger/             # 日志记录
│   ├── middleware/         # HTTP 中间件
│   ├── model/              # 模型管理
│   ├── models/             # 数据模型定义
│   ├── proxy/              # LLM 代理和负载均衡
│   ├── quota/              # 配额检查
│   ├── static/             # 静态文件嵌入
│   ├── usage/              # 使用记录
│   └── user/               # 用户管理
├── web/                    # Web 前端（React + TS）
├── config.yaml             # 配置文件
├── Makefile                # 构建脚本
└── README.md
```

## 构建命令

```bash
make build      # 构建完整应用（前端 + Go）
make build-go   # 仅构建 Go（不构建前端）
make run        # 构建并运行
make dev        # 开发模式（前后端同时运行）
make release    # 构建多平台发布包
make clean      # 清理构建产物
make test       # 运行测试
```

## 测试

```bash
# 运行所有场景测试
go test ./test/scenarios/... -v

# 运行特定场景
go test ./test/scenarios/... -v -run TestScenario_UserEndToEndFlow

# 生成覆盖率报告
go test ./test/scenarios/... -cover
```

测试覆盖场景：
- ✅ 用户完整流程（注册→登录→创建Key→调用→扣减）
- ✅ 配额限制（日配额超限、多模型配额）
- ✅ API Key 生命周期（过期、禁用、用户禁用）
- ✅ 速率限制与并发控制

## 部署建议

### 单机部署

```bash
# 1. 复制两个文件到服务器
scp llmgate config.yaml user@server:/opt/llmgate/

# 2. 使用 systemd 管理服务
sudo systemctl enable --now llmgate
```

### 使用 Nginx 反向代理

```nginx
server {
    listen 80;
    server_name llm.company.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## 安全建议

1. **修改默认密码**：首次登录后立即修改管理员密码
2. **更换 JWT Secret**：生产环境务必使用强密钥
3. **启用 HTTPS**：使用 Nginx 或 Caddy 提供 HTTPS
4. **定期备份**：备份 SQLite 数据库文件
5. **日志审计**：定期检查日志目录的访问记录

## License

MIT License - 详见 [LICENSE](LICENSE) 文件

---

**注意**：本项目默认配置仅供开发测试使用，生产环境请务必修改默认密码和 JWT Secret。
