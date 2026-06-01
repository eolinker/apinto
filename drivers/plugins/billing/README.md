# billing 通用计费拦截插件

## 概述

`billing` 是一个通用的请求级计费拦截插件，覆盖 AI 大模型与标准 API 两类调用场景。它在 HTTP 过滤链中拦截请求/响应，按 `provider/resource/phase` 三元组查找已注册的价格计算器（由 `drivers/pricing` 提供），完成两阶段余额扣减、防异步任务重复计费，并将计费结果写入上下文供日志/审计下游消费。

与旧的 `ai-price` 相比：

- 抽象层从 `ai-convert` 迁出，独立到顶层 `pricing` 包，去除 AI 专属命名
- `model` → `resource`，新增 `method`/`path` 条件类型，支持按 HTTP 状态码差异化阶梯计费
- Redis 键前缀切换到 `apinto:billing:*`，与旧版数据物理隔离
- 移除 `total_tokens/video_size/input_tokens` 等旧字段别名，仅识别新版命名

## 配置示例

```yaml
- name: billing-default
  driver: billing
  request_fields:
    provider: $.provider
    resource: $.model
    input_count: $.input_count
  response_fields:
    output_count: $.usage.output_count
    total_count: $.usage.total_count
    task_id: $.task_id
  cache: redis@resource              # 可选 Redis 资源
  task_ttl: 86400                    # 异步任务缓存秒
  enable_balance: true               # 是否启用余额扣减
  match_rules:
    method: POST
    path_prefix: /api/v1
  result_match:
    status_codes: [200]
    success_field: code
    success_value: 0
```

## 处理流程

```
matchRules → 提取请求字段 → 查找计算器 → PreDeduct → 转发下游
   ↓ (失败)                                              ↓
   放行                                              Rollback
                                                          ↓
                  提取响应字段 → 异步侦测 → 已计费幂等
                                                          ↓
                                          resultMatch → Calculate → MarkBilled → Settle
                                                          ↓ (不匹配)
                                                       Rollback
```

## 配套组件

- 计算器：`drivers/pricing`（profession=resource，driver=pricing）
- 契约层：顶层 `github.com/eolinker/apinto/pricing`
- 余额管理：基于 Redis Lua 脚本，无 Redis 时降级放行
- 任务执行：Redis 优先 + 本地内存回退（双层防重复计费）

## 上下文标签

| 标签 | 含义 |
| --- | --- |
| `pricing_result` | 计费结果 JSON |
| `pricing_fields` | 计费字段集合 JSON |
| `pricing_phase` | 计费阶段 |
| `pricing_freeze_id` | 预扣冻结 ID |
| `pricing_task_duplicate` | 任务已计费标记 |
| `pricing_result_unmatched` | 未通过 result_match |
