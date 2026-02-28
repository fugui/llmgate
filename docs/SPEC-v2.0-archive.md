# LLMGATE - 企业内部 LLM 管理平台
## 功能规格说明书 V2.0

---

## 1. 系统定位

**目标**：为 1000 人规模的企业提供统一的内部 LLM 服务入口，实现用户管理、权限控制、配额计费和审计追踪。

**核心场景**：
- 企业自建 3 台 LLM 服务器（可能随时扩展和收缩）
- 员工通过统一平台访问
- 管理员控制谁能用什么、用多少

---

## 2. 核心功能模块

### 2.1 用户管理系统

#### 2.1.1 用户身份
| 功能 | 规格 |
|------|------|
| 登录方式 | 企业 SSO（OIDC/SAML/LDAP）或本地账号 |
| 用户角色 | 超级管理员、部门管理员、普通用户 |
| 用户组 | 支持部门/团队分组，批量授权 |

#### 2.1.2 API Key 生命周期
| 功能 | 规格 |
|------|------|
| 自建 Key | 用户登录后可自行创建/删除 API Key |
| Key 命名 | 支持自定义名称（如"开发测试"、"生产脚本"） |
| Key 状态 | 启用/禁用/过期 |
| 权限继承 | Key 继承用户权限，也可单独限制 |

### 2.2 模型权限管理

#### 2.2.1 模型定义
```yaml
models:
  - id: "llama3-70b"
    name: "Llama 3 70B"
    backend: "http://llm-server-1:8000"
    enabled: true
    
  - id: "qwen-72b"
    name: "Qwen 72B"
    backend: "http://llm-server-2:8000"
    enabled: true
    
  - id: "deepseek-67b"
    name: "DeepSeek 67B"
    backend: "http://llm-server-3:8000"
    enabled: true
```

#### 2.2.2 授权粒度
| 级别 | 权限控制 |
|------|----------|
| 全局默认 | 新用户默认可见模型 |
| 用户组级 | 按部门批量分配模型权限 |
| 用户级 | 单个用户的模型白名单 |
| Key 级 | 特定 API Key 的模型限制 |

### 2.3 配额与限流系统

#### 2.3.1 多维度配额
| 维度 | 说明 | 示例 |
|------|------|------|
| 速率限制 | 每分钟/每小时请求数 | 60 req/min |
| Token 配额 | 按天/周/月的 Token 上限 | 100K tokens/day |
| 并发限制 | 同时进行的请求数 | 5 concurrent |

#### 2.3.2 配额策略
```yaml
quota_policies:
  - name: "开发团队"
    rate_limit: 120/min
    token_quota: 500K/day
    models: ["llama3-70b", "qwen-72b"]
    
  - name: "普通员工"
    rate_limit: 30/min
    token_quota: 50K/day
    models: ["llama3-70b"]
    
  - name: "VIP 用户"
    rate_limit: 300/min
    token_quota: 2M/day
    models: ["*"]  # 全部模型
```

#### 2.3.3 配额行为
- **软限制**：超配额时发出警告但允许继续（可配置）
- **硬限制**：超配额时直接拒绝请求
- **自动重置**：按日/周/月自动重置配额计数

### 2.4 负载均衡与后端管理

#### 2.4.1 后端 LLM 服务器
| 功能 | 规格 |
|------|------|
| 健康检查 | 定期探测后端可用性 |
| 权重分配 | 支持按性能分配流量权重 |
| 故障转移 | 后端不可用时自动剔除 |
| 会话保持 | 同一对话路由到同一后端 |

#### 2.4.2 路由策略
- Round Robin（轮询）
- Least Connections（最少连接）
- IP Hash（基于用户）

### 2.5 审计与监控

#### 2.5.1 请求日志
| 字段 | 说明 |
|------|------|
| 时间戳 | 请求时间 |
| 用户 ID | 谁发起的 |
| API Key | 使用的 Key |
| 模型 | 调用的模型 |
| 输入 Token | 请求消耗的 Token |
| 输出 Token | 响应消耗的 Token |
| 响应时间 | 耗时 |
| 状态码 | 成功/失败 |

#### 2.5.2 统计分析
- 用户/部门使用排行
- 模型热度分析
- Token 消耗趋势
- 配额使用率预警

---

## 3. 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                        用户层                                │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐                     │
│  │ Web UI  │  │  CLI    │  │  App    │                     │
│  └────┬────┘  └────┬────┘  └────┬────┘                     │
└───────┼────────────┼────────────┼───────────────────────────┘
        │            │            │
        └────────────┴────────────┘
                     │
        ┌────────────▼────────────┐
        │      LLMGATE           │
        │  ┌─────────────────┐   │
        │  │  Auth Middleware │   │  ← API Key 验证
        │  └────────┬────────┘   │
        │           ▼            │
        │  ┌─────────────────┐   │
        │  │  Quota Check    │   │  ← 配额检查
        │  └────────┬────────┘   │
        │           ▼            │
        │  ┌─────────────────┐   │
        │  │  Model Router   │   │  ← 模型权限检查
        │  └────────┬────────┘   │
        │           ▼            │
        │  ┌─────────────────┐   │
        │  │ Load Balancer   │   │  ← 后端选择
        │  └────────┬────────┘   │
        └───────────┼────────────┘
                    │
        ┌───────────┼───────────┐
        ▼           ▼           ▼
   ┌─────────┐ ┌─────────┐ ┌─────────┐
   │ LLM-1   │ │ LLM-2   │ │ LLM-3   │
   │ :8000   │ │ :8001   │ │ :8002   │
   └─────────┘ └─────────┘ └─────────┘
```

---

## 4. 数据模型

### 4.1 用户 (User)
```go
type User struct {
    ID          string
    Email       string
    Name        string
    Role        Role          // admin, manager, user
    Department  string
    QuotaPolicy string        // 关联的配额策略
    Models      []string      // 允许访问的模型
    CreatedAt   time.Time
    LastLogin   time.Time
}
```

### 4.2 API Key
```go
type APIKey struct {
    ID          string
    UserID      string
    Name        string
    Key         string        // 实际 Key（加密存储）
    Models      []string      // Key 级别的模型限制
    RateLimit   int           // Key 级别的速率限制
    Enabled     bool
    ExpiresAt   *time.Time
    LastUsedAt  *time.Time
}
```

### 4.3 使用记录 (Usage)
```go
type UsageRecord struct {
    ID           string
    Timestamp    time.Time
    UserID       string
    APIKeyID     string
    ModelID      string
    BackendID    string
    InputTokens  int
    OutputTokens int
    LatencyMs    int
    StatusCode   int
    ErrorMsg     string
}
```

### 4.4 配额记录 (Quota)
```go
type QuotaRecord struct {
    UserID       string
    Date         string          // YYYY-MM-DD
    ModelID      string
    RequestCount int
    TokenCount   int
    LimitReached bool
}
```

---

## 5. API 接口设计

### 5.1 用户接口
```
POST   /api/v1/auth/login           # 登录
POST   /api/v1/auth/logout          # 登出
GET    /api/v1/user/profile         # 获取用户信息
GET    /api/v1/user/keys            # 获取我的 API Keys
POST   /api/v1/user/keys            # 创建 API Key
DELETE /api/v1/user/keys/:id        # 删除 API Key
GET    /api/v1/user/quota           # 获取配额使用情况
GET    /api/v1/user/usage           # 获取使用记录
```

### 5.2 管理接口
```
# 用户管理
GET    /api/v1/admin/users
POST   /api/v1/admin/users
PUT    /api/v1/admin/users/:id
DELETE /api/v1/admin/users/:id

# 模型管理
GET    /api/v1/admin/models
POST   /api/v1/admin/models
PUT    /api/v1/admin/models/:id

# 配额策略
GET    /api/v1/admin/policies
POST   /api/v1/admin/policies
PUT    /api/v1/admin/policies/:id

# 系统统计
GET    /api/v1/admin/stats/usage
GET    /api/v1/admin/stats/tokens
GET    /api/v1/admin/stats/models
```

### 5.3 LLM 代理接口
```
POST   /v1/chat/completions        # OpenAI 兼容格式
POST   /v1/completions             # OpenAI 兼容格式
GET    /v1/models                  # 列出可用模型
```

---

## 6. 技术栈建议

| 层级 | 技术选择 |
|------|----------|
| 后端 | Go + Gin/Echo |
| 数据库 | PostgreSQL（主库）+ Redis（缓存/限流） |
| 前端 | React/Vue.js（管理后台） |
| 认证 | JWT + 可选 SSO 集成 |
| 部署 | Docker + Docker Compose |

---

## 7. 确认事项（已确认）

| 项目 | 决策 |
|------|------|
| 认证方式 | 第一阶段：本地 JWT 认证；第二阶段：企业 SSO |
| Web 界面 | ✅ 需要管理后台（React） |
| 超额处理 | 直接拦截请求，无需预警通知 |
| 日志保留 | 7 天自动清理 |

## 8. 第一阶段开发范围（MVP）

### 后端（Go）
- [ ] JWT 本地认证
- [ ] 用户/Key/模型 CRUD
- [ ] 配额检查（速率 + Token）
- [ ] 3 台后端 LLM 负载均衡
- [ ] 审计日志（7天保留）

### 前端（React）
- [ ] 登录页
- [ ] 用户 Dashboard（配额查看、Key 管理）
- [ ] 管理员后台（用户/模型/策略管理）

### 数据库
- [ ] PostgreSQL 主库
- [ ] Redis（限流计数器 + 缓存）

---

请确认以上规格是否符合需求，或需要调整哪些部分？