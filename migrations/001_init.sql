-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'user' CHECK (role IN ('admin', 'manager', 'user')),
    department VARCHAR(100),
    quota_policy VARCHAR(50) NOT NULL DEFAULT 'default',
    models TEXT[],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP WITH TIME ZONE,
    enabled BOOLEAN DEFAULT true
);

-- API Key 表
CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    key_hash VARCHAR(255) UNIQUE NOT NULL,
    key_prefix VARCHAR(20) NOT NULL,
    models TEXT[],
    rate_limit INTEGER,
    rate_limit_window INTEGER DEFAULT 60,
    enabled BOOLEAN DEFAULT true,
    expires_at TIMESTAMP WITH TIME ZONE,
    last_used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 模型配置表
CREATE TABLE IF NOT EXISTS models (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    backend_url VARCHAR(255) NOT NULL,
    enabled BOOLEAN DEFAULT true,
    weight INTEGER DEFAULT 1,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 配额策略表
CREATE TABLE IF NOT EXISTS quota_policies (
    name VARCHAR(50) PRIMARY KEY,
    rate_limit INTEGER NOT NULL DEFAULT 60,
    rate_limit_window INTEGER NOT NULL DEFAULT 60,
    token_quota_daily BIGINT NOT NULL DEFAULT 100000,
    models TEXT[],
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 使用记录表（7天保留，使用分区表）
CREATE TABLE IF NOT EXISTS usage_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    api_key_id UUID REFERENCES api_keys(id) ON DELETE SET NULL,
    model_id VARCHAR(50) NOT NULL,
    backend_url VARCHAR(255) NOT NULL,
    input_tokens INTEGER DEFAULT 0,
    output_tokens INTEGER DEFAULT 0,
    latency_ms INTEGER DEFAULT 0,
    status_code INTEGER DEFAULT 200,
    error_msg TEXT,
    request_path VARCHAR(255),
    request_method VARCHAR(10)
) PARTITION BY RANGE (timestamp);

-- 创建分区（当前周和下周）
CREATE TABLE IF NOT EXISTS usage_records_2024_01 PARTITION OF usage_records
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

-- 配额使用统计表（按天汇总）
CREATE TABLE IF NOT EXISTS quota_usage_daily (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    model_id VARCHAR(50),
    request_count INTEGER DEFAULT 0,
    token_count BIGINT DEFAULT 0,
    input_tokens BIGINT DEFAULT 0,
    output_tokens BIGINT DEFAULT 0,
    UNIQUE(user_id, date, model_id)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_enabled ON users(enabled);
CREATE INDEX IF NOT EXISTS idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX IF NOT EXISTS idx_api_keys_enabled ON api_keys(enabled);
CREATE INDEX IF NOT EXISTS idx_usage_records_user_id ON usage_records(user_id);
CREATE INDEX IF NOT EXISTS idx_usage_records_timestamp ON usage_records(timestamp);
CREATE INDEX IF NOT EXISTS idx_usage_records_model_id ON usage_records(model_id);
CREATE INDEX IF NOT EXISTS idx_quota_usage_user_date ON quota_usage_daily(user_id, date);

-- 插入默认配额策略
INSERT INTO quota_policies (name, rate_limit, rate_limit_window, token_quota_daily, models, description)
VALUES ('default', 60, 60, 100000, ARRAY['llama3-70b', 'qwen-72b', 'deepseek-67b'], '默认用户配额')
ON CONFLICT (name) DO NOTHING;

INSERT INTO quota_policies (name, rate_limit, rate_limit_window, token_quota_daily, models, description)
VALUES ('vip', 300, 60, 1000000, ARRAY['*'], 'VIP用户配额')
ON CONFLICT (name) DO NOTHING;

-- 插入默认模型
INSERT INTO models (id, name, backend_url, enabled, weight, description)
VALUES ('llama3-70b', 'Llama 3 70B', 'http://localhost:8001', true, 1, 'Llama 3 70B 模型')
ON CONFLICT (id) DO NOTHING;

INSERT INTO models (id, name, backend_url, enabled, weight, description)
VALUES ('qwen-72b', 'Qwen 72B', 'http://localhost:8002', true, 1, 'Qwen 72B 模型')
ON CONFLICT (id) DO NOTHING;

INSERT INTO models (id, name, backend_url, enabled, weight, description)
VALUES ('deepseek-67b', 'DeepSeek 67B', 'http://localhost:8003', true, 1, 'DeepSeek 67B 模型')
ON CONFLICT (id) DO NOTHING;
