# 默认模型 Fallback 功能

## 功能说明

当客户端请求一个模型，但该模型没有配置可用的后端时，系统可以自动 fallback 到一个默认的模型，而不是直接返回 503 错误。

## 使用场景

1. **平滑迁移**：新旧模型切换时，不需要立即更新所有客户端配置
2. **负载均衡**：多个模型可以共用同一个后端集群
3. **降级处理**：某个模型后端故障时，自动使用备用模型
4. **实验测试**：新模型上线前，可以先配置为 fallback 到已有模型

## 配置方法

在 `config.yaml` 中添加 `default_model` 字段：

```yaml
# 配置默认模型
default_model: "kimi2.5"

models:
  - id: kimi2.5
    name: Kimi 2.5
    enabled: true
    backends:
      - id: kimi2.5-backend
        name: Kimi 后端
        base_url: https://api.kimi.com/v1
        api_key: sk-your-key
        enabled: true

  # 这个模型没有配置后端
  - id: gpt-4
    name: GPT-4
    enabled: true
    backends: []
```

## 工作原理

1. 客户端请求 `gpt-4` 模型
2. 系统检查 `gpt-4` 是否有可用的后端
3. 如果没有，检查是否配置了 `default_model`
4. 如果配置了，使用 `default_model` 的后端处理请求
5. 日志会记录：`"using fallback model kimi2.5 for request model gpt-4"`

## 配置注意事项

1. **default_model 为空**：不启用 fallback 功能，直接返回 503
2. **default_model 必须有效**：配置的模型必须有至少一个启用的后端
3. **避免循环依赖**：default_model 本身不需要 fallback
4. **配额限制**：fallback 后的模型受原模型的配额策略限制

## 日志示例

```
# 正常情况
Next: no backends found for model gpt-4, trying default model kimi2.5
Next: using fallback model kimi2.5 for request model gpt-4 (backend: kimi2.5-gz-01)

# default_model 也没有后端的情况
Next: no backends found for model gpt-4, trying default model kimi2.5
Next: no backends found for model kimi2.5 (not in map)
# 返回 503 错误
```

## API 响应

当发生 fallback 时，API 响应会显示实际使用的模型：

```json
{
  "id": "chatcmpl-xxx",
  "object": "chat.completion",
  "created": 1234567890,
  "model": "kimi-for-coding",
  "choices": [...]
}
```

注意：`model` 字段显示的是实际调用的后端模型名称（如 `kimi-for-coding`），而不是客户端请求的模型名称。
