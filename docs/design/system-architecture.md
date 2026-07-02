# Apinto 系统设计文档

> 状态：现状梳理（As-Is）
> 版本：基于当前 main 分支代码
> 目标读者：新加入的研发、二次开发者、SRE

本文从"是什么、为什么、怎么组织、怎么跑起来"四个维度，梳理 apinto 网关的整体架构。所有引用均指向仓库内文件，可直接跳转。

---

## 1. 定位与设计目标

Apinto 是一款基于 Go 的云原生 API 网关，同时承担传统 API 网关和 AI 网关两类流量入口的职责。核心目标：

1. **协议多样**：南北向支持 HTTP/HTTPS、WebSocket、gRPC、Dubbo2，并可协议互转（`http↔dubbo2`、`http↔gRPC`）。
2. **动态可编排**：配置以 profession/driver/worker 的形式在运行期热更新，无需重启进程。
3. **插件化**：请求处理流程完全由插件链驱动，业务能力（限流、鉴权、缓存、AI 转换、动态计费等）以插件形式挂载。
4. **多进程高可用**：基于 eosc 框架跑成 `master + worker + admin + helper` 的多进程模型，主从热切换、配置广播。
5. **AI 网关能力**：内置 30+ AI Provider 适配、Key 池、模型抽象、Prompt/Formatter/Proxy/动态计费等 AI 专属插件。

---

## 2. 技术栈与依赖

| 层次 | 依赖 |
|---|---|
| 进程/生命周期 | `github.com/eolinker/eosc`（master/worker/admin/helper、profession/driver/worker 抽象、配置广播） |
| HTTP/gRPC/Dubbo | `fasthttp` 客户端（[node/fasthttp-client](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/node/fasthttp-client)）、`grpc-go`、Dubbo Getty（[dubbo-getty](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/dubbo-getty)） |
| 服务发现 | Nacos / Consul / Eureka / Polaris / Kubernetes / Static（[drivers/discovery](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/drivers/discovery)） |
| 缓存/资源 | Redis（[drivers/resources/redis](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/drivers/resources/redis)）、本地 cache/vector（[resources](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/resources)） |
| 可观测 | Prometheus / InfluxDB v2 / Loki / Kafka / NSQ / file / syslog / http（[drivers/output](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/drivers/output)） |
| AI | 30+ Provider（[drivers/ai-provider](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/drivers/ai-provider)）、通用 convert 抽象（[ai-convert](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/ai-convert)） |
| 表达式/计价 | [price-calcular](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/price-calcular)（表达式、变量提取、动态计费） |

---

## 3. 进程模型（eosc 承载）

入口 [main.go](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/app/apinto/main.go#L25-L57) 使用 eosc 的多进程框架注册四类角色：

```
Master  ← 配置持久化、Raft 集群、变更广播
  ├─ Worker  ← 真正处理南北向流量的工作进程
  ├─ Admin   ← OpenAPI/管理面
  └─ Helper  ← 辅助进程（信号处理、日志转发等）
```

- Master：[master.go](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/app/apinto/master.go) 用 `process_master.MasterHandler` + `InitProfession` 启动，Profession 列表见 [profession.go](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/app/apinto/profession.go)。
- Worker：[worker.go](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/app/apinto/worker.go) 通过 `extends.AddInnerExtendProject("eolinker.com", "apinto", Register)` 把所有 driver/plugin 注入 eosc 扩展仓库，然后 `process_worker.Process()` 起流量面。
- Register：[register.go](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/app/apinto/register.go)、[driver.go](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/app/apinto/driver.go)、[plugin.go](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/app/apinto/plugin.go) 把 40+ driver 与 60+ plugin 汇总为一次注册。

关键点：**配置变更只在 Master，Master 向 Worker 广播差异；Worker 内的每个 profession/driver/worker 都实现 `Reset(conf, workers)`，热更新时不重建进程。**

---

## 4. 领域模型（Profession/Driver/Worker）

Apinto 用 eosc 的三层抽象来组织"网关配置对象"：

- **Profession**：一大类资源，如 `router / service / discovery / app / strategy / output / certificate / transcode / ai-resource / pricing`，声明依赖关系与允许的 driver。定义见 [profession.go](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/app/apinto/profession.go#L14-L334)。
- **Driver**：某个 profession 下的实现类型。例如 `router` profession 下有 `http_router / grpc_router / dubbo2_router`。
- **Worker**：driver 的一次配置化实例。用户在 dashboard 上创建的每一条"路由/服务/应用/策略"，都会实例化成一个 worker。

### 4.1 Profession 关系与依赖

```
              transcode
                 ↑
        router → service → discovery
          │
          ├─ template
          ├─ plugins（chain）
          └─ certificate

  strategy（限流/缓存/灰度/访问/熔断/脱敏）—— 独立注入到请求链
  app（消费者/渠道） + auth（AK/SK、Basic、JWT、API Key、OAuth2、OIDC）
  output（Prometheus/Kafka/Loki/InfluxDB…）
  ai-resource（ai-key / ai-provider / ai-model）
  pricing（pricing-policy）
```

Profession 显式声明 `Dependencies`（如 `router → service → discovery`），eosc 会按依赖顺序初始化 worker，避免"路由绑定的服务还没就绪"这类时序问题。

### 4.2 Driver 目录约定

所有 driver 都在 [drivers/](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/drivers) 下，标准骨架：

```
drivers/<name>/
  factory.go   ← 驱动工厂 + Register 函数
  driver.go    ← 实现 eosc.IExtenderDriver
  config.go    ← 配置结构体 + JSON tag
  <name>.go    ← Worker 实体
  worker.go    ← （可选）Worker 生命周期
```

代码里通过 `xxx.Register(extenderRegister)` 把工厂交给 eosc；见 [driver.go](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/app/apinto/driver.go#L46-L105) 与 [plugin.go](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/app/apinto/plugin.go#L68-L152)。

---

## 5. 请求处理主链路（以 HTTP 为例）

以最常见的 HTTP 流量为例，全链路如下：

```
┌──────────────────────────────────────────────────────────────────────┐
│  Listener（router.Listener，端口 + 协议）                             │
│    │                                                                 │
│    ▼                                                                 │
│  Router 匹配（http-router/matcher）                                   │
│    - method / host / path                                            │
│    - 追加规则：header/query/cookie 等（AppendRule）                    │
│    │                                                                 │
│    ▼                                                                 │
│  httpHandler.Serve                                                   │
│    - 打标签（api/api_id/service/ip/time…）                            │
│    - Set retry / timeout                                             │
│    - Set CompleteHandler、Balance、UpstreamHost、Finisher             │
│    │                                                                 │
│    ▼                                                                 │
│  插件链 filters.Chain(ctx, completeCaller)                            │
│    - Global 插件 + Router 插件 + Template 插件（合并优先级）           │
│    - 每个插件实现 eocontext.IFilter                                    │
│    │                                                                 │
│    ▼                                                                 │
│  CompleteHandler（http-complete / websocket / stream）                │
│    - 通过 Service 拿到上游节点（balance）                              │
│    - fasthttp-client 发起转发（含 retry / timeout）                    │
│    │                                                                 │
│    ▼                                                                 │
│  响应处理链（逆序）                                                    │
│    │                                                                 │
│    ▼                                                                 │
│  Finisher（access-log / metrics / monitor / proxy-mirror）             │
└──────────────────────────────────────────────────────────────────────┘
```

关键代码：

- 路由 reset + 组装 handler：[router.go](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/drivers/router/http-router/router.go#L46-L125)
- 请求分发：[http-handler.go](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/drivers/router/http-router/http-handler.go#L36-L85)
- 匹配算法：[router/http-router](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/router/http-router)（`append.go` / `matcher.go` / `rule.go`）
- 规则打分与排序：[router/rule.go](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/router/rule.go#L94-L115)

### 5.1 上下文（EoContext）

Apinto 的请求上下文由 `eosc/eocontext` 定义，本项目在 [node/http-context](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/node/http-context) 中实现具体的 HTTP/WebSocket 版本。上下文承载：

- 请求/响应对象（含 body 二次读取、header/uri 改写）
- 标签（`SetLabel`），全链路可读
- 值（`WithValue`），如 `CtxKeyRetry`、`CtxKeyTimeout`（[entries/ctx_key](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/entries/ctx_key/ctx_key.go)）
- Complete/Balance/UpstreamHost/Finish 四类回调

其它协议：[node/grpc-context](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/node/grpc-context)、[node/dubbo2-context](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/node/dubbo2-context)。

---

## 6. 插件系统

### 6.1 抽象

- 单个插件是 `eocontext.IFilter`，实现 `DoFilter(ctx, next)`。
- 一组插件组成链 `IChainPro`，由 [drivers/plugin-manager](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/drivers/plugin-manager) 统一创建。
- 顶层接口：[plugin/plugin.go](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/plugin/plugin.go#L14-L45)。
- Config 合并：路由级 > 模板级 > 全局级，用 `MergeConfig` 合并。

### 6.2 插件分类（当前已注册）

见 [plugin.go](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/app/apinto/plugin.go)：

| 分组 | 代表插件 |
|---|---|
| **服务治理策略** | limiting / cache / grey / visit / fuse / data_mask |
| **协议转换** | http↔dubbo2、http↔gRPC、`*-proxy-rewrite` |
| **请求处理** | body-check、extra-params(v2)、params-transformer、proxy-rewrite(v2)、params-check(v2)、data-transform、http_mocking、request-file-parse、request-interception |
| **响应处理** | response-rewrite(v2)、response-filter、response-file-parse、gzip、auto-redirect |
| **安全/鉴权** | ip-restriction、rate-limiting、cors、circuit-breaker、rsa-filter、aes、js-inject、acl、access-relational、replay-attack-defender、oauth2、oauth2-introspection、custom-oauth2-introspection |
| **观测/上报** | access-log、prometheus、monitor、proxy-mirror |
| **计数器** | counter（内置 Lua、Redis 支持） |
| **AI 网关** | ai-prompt、ai-formatter、ai-proxy |
| **计费** | pricing-policy-base、dynamic-billing |
| **脚本** | script-handler |

### 6.3 Plugin Manager

[plugin-manager/manager.go](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/drivers/plugin-manager/manager.go#L69-L145)：

- 单例（`SettingModeSingleton`），承载全局 `plugin@setting`。
- `CreateRequest(id, conf)` 为每个路由创建一份链，缓存到 `pluginObjs`。
- `Reset` 时对所有已有 chain 做全量 rebuild（旧的 Destroy）。
- 插件三态：`StatusGlobal`（强制加载） / `StatusDisable`（禁用） / 其他（按路由配置）。

---

## 7. 策略（Strategy）子系统

策略与普通插件的差异：

- 策略是"横切规则"，可以按 `filter`（应用/IP/API/上游/自定义 label 等）匹配一批请求。
- 每类策略是一个 profession 下的 driver（[drivers/strategy](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/drivers/strategy)），同时通过对应的**插件形态**（[drivers/plugins/strategy](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/drivers/plugins/strategy)）挂进请求链。
- 每个策略 driver 内部包含 `actuator + controller + handler + config` 四件套，见 [limiting-strategy](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/drivers/strategy/limiting-strategy) 的目录结构。

抽象在 [strategy/strategy.go](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/strategy/strategy.go)：

```go
type IStrategyHandler interface { Strategy(ctx, next) error }
type IFilter interface { Check(ctx) bool }
```

匹配器：[strategy/filter.go](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/strategy/filter.go)、[strategy/checker.go](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/strategy/checker.go) 与通用 [checker](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/checker) 包（equal / prefix / suffix / regex / sub / all / exist / notequal / none）。

---

## 8. 应用（App / Consumer）与鉴权

### 8.1 应用抽象

[application/app.go](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/application/app.go#L25-L57)：

- `IApp`：Id/Name/Labels/Disable + `IAppExecutor.Execute(ctx)`。
- `IAuth`：一个鉴权驱动的多用户管理器（对应一批 `ITransformConfig`）。
- `IAuthUser`：从请求中还原出 `UserInfo`。

### 8.2 App Driver（渠道 = 消费者）

[drivers/app](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/drivers/app)：

- 配置内含多种鉴权用户（AK/SK、API Key、Basic、JWT、OAuth2、OIDC-JWT、para-hmac）。
- `executor` 是一组 `IAppExecutor` 的顺序执行链（[executor.go](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/drivers/app/executor.go#L18-L38)），命中鉴权后为请求打上应用标签，后续策略/计费/日志均基于此标签。
- 鉴权驱动集中在 [application/auth](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/application/auth)，每一种鉴权是一个独立 driver。

### 8.3 请求鉴权触发点

在路由链中，`app` 插件（[drivers/plugins/app](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/drivers/plugins/app)）负责匹配应用；`auth-interceptor` 用于统一拦截；acl / access-relational 等控制授权关系。

---

## 9. 服务与服务发现

- Service 抽象：[service/service.go](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/service/service.go)。同时实现 `BalanceHandler`（负载均衡）和 `UpstreamHostHandler`（重写目标 Host）两大 eosc 接口。
- Service Driver：[drivers/service](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/drivers/service) 承载超时、重试、协议、负载均衡算法等。
- 负载均衡算法：[upstream/round-robin](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/upstream/round-robin)、[ip-hash](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/upstream/ip-hash)、[session-keep](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/upstream/session-keep)。
- 服务发现：`discovery` profession，接入 [nacos](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/drivers/discovery/nacos) / [consul](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/drivers/discovery/consul) / [eureka](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/drivers/discovery/eureka) / [polaris](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/drivers/discovery/polaris) / [kubernetes](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/drivers/discovery/kubernetes) / [static](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/drivers/discovery/static)，共享 [discovery](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/discovery) 里的 `app/node/nodes/check` 抽象。
- 健康检查：[health-check-http](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/health-check-http)。

---

## 10. AI 网关子系统

Apinto 的 AI 能力由三层构成：

### 10.1 资源层（ai-resource profession）

- **ai-provider**（[drivers/ai-provider](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/drivers/ai-provider)）：每个厂商一个子包（openAI、authropic、bedrock、tongyi、hunyuan、moonshot、deepseek…），核心是 `RequestConvert` / `ResponseConvert`，把 OpenAI 兼容请求或原生格式，翻译到具体 Provider 的协议。
- **ai-key**（[drivers/ai-key](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/drivers/ai-key)）：厂商 API Key 池，支持轮询/失效摘除。
- **ai-model**（[drivers/ai-model](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/drivers/ai-model)）：模型对象，串接 provider + key + 配置。

### 10.2 转换抽象层

[ai-convert](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/ai-convert)：定义 `IConverter / IConverterDriver / ModelType` 等抽象（见 [convert.go](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/ai-convert/convert.go#L20-L116)）。

支持的 ModelType：`chat / openai-chat / image-generation / image-edit / image-task-commit / image-task-query / video-task-commit / video-task-query`。

### 10.3 插件层

- `ai-proxy`（[drivers/plugins/ai-proxy](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/drivers/plugins/ai-proxy)）：核心转发插件，从 ctx 中取 provider/key/model，走对应 converter。
- `ai-prompt`：Prompt 模板注入。
- `ai-formatter`：请求/响应格式规范化。
- `dynamic-billing`（[drivers/plugins/dynamic-billing](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/drivers/plugins/dynamic-billing)）：面向 AI/API 通用的动态计费，含并发限制、余额前置校验、消费扣款；余额 key 存 Redis，负值即拒绝（[executor.go:240](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/drivers/plugins/dynamic-billing/executor.go#L240)）。
- `pricing-policy-base`：定价规则封装，共享 [price-calcular](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/price-calcular) 里的表达式引擎（`expression.go`、`variable-extractor.go`、`calculator.go`）。

### 10.4 上下文标签

AI 通道用 [common/context-label](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/common/context-label) 在 ctx 中打通：`billing-mode`、`consumer`、`pricing`、`stream`、`task`、`model`、`key-generate` 等，供计费/日志/统计插件消费。

---

## 11. 可观测性

### 11.1 输出器（output profession）

支持 file / http / nsq / kafka / syslog / prometheus / redis / influxdb-v2 / loki 等输出目标，见 [drivers/output](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/drivers/output)。

### 11.2 请求侧输出

- [entries/http-entry](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/entries/http-entry)：HTTP 请求可读字段抽象（reader-index/reader/proxy-reader）。
- [entries/metric-entry](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/entries/metric-entry)：Prometheus 指标字段。
- [entries/monitor-entry](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/entries/monitor-entry)：Monitor 插件字段。

### 11.3 输出链

- [output/output.go](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/output/output.go)、[output/metrics.go](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/output/metrics.go)。
- [monitor-manager](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/monitor-manager)：并发监控/聚合。

---

## 12. 资源与缓存

- **Redis**（[drivers/resources/redis](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/drivers/resources/redis)）：作为分布式缓存 + Lua 脚本能力（add/compare_and_add/get），是限流、计数、计费、去重的共享层。
- **本地 cache/vector**（[resources](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/resources)）：进程内高性能缓存。
- **证书**（[drivers/certs](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/drivers/certs)、[drivers/gm-certs](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/drivers/gm-certs)）：普通/国密证书统一装载到 SNI。

---

## 13. 关键设计模式

### 13.1 依赖注入（Autowired）

Plugin Manager 通过 `bean.Autowired(&pm.extenderDrivers)` 拿到 eosc 全局的驱动仓库（[manager.go:196-210](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/drivers/plugin-manager/manager.go#L196-L210)），避免显式全局单例。

### 13.2 Skill（能力嗅探）

Worker 之间靠 `CheckSkill(skill string) bool` 做能力协商，例如路由要求依赖对象是 Service（[router.go:98](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/drivers/router/http-router/router.go#L98)）或 Template（同文件 L78）。Skill 是一个字符串常量，如 `service.ServiceSkill`（[service.go:9](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/service/service.go#L9)）。

### 13.3 责任链

请求处理链 = `Global 插件链 + Router 插件链` 合并展开，最终以 `IChainPro.Chain(ctx, next)` 递归调用。

### 13.4 事实/派生分层（渠道-资源组场景）

参见 [channel-resource-group.md](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/docs/design/channel-resource-group.md)：关系层（fact）+ 物化层（derivation）+ 缓存层，写路径只碰关系层，读路径走缓存；reconciler 异步展开级联。

---

## 14. 配置与生命周期

- 配置文件：`/etc/apinto/apinto.yml`（进程） + `/etc/apinto/config.yml`（业务）。生成脚本 [build/resources](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/build/resources)。
- 变更路径：**Dashboard/OpenAPI → Admin → Master（Raft 持久化 + 广播） → Worker.Reset**。
- Worker 重建代价：只重建被影响的 driver worker（如某条路由的规则/插件），不影响其他路由。

---

## 15. 部署形态

| 形态 | 说明 |
|---|---|
| **单机** | 直接跑 `apinto start`，master+worker+admin+helper 各一份。 |
| **集群** | 多节点 Raft，Master 主从热切换；Worker 独立处理各自流量。 |
| **Docker** | [build/resources/Dockerfile](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/build/resources/Dockerfile)；发布仓库 `eolinker/apinto-gateway`。 |
| **K8s** | 通过 discovery-kubernetes 融入集群服务发现；配置注入见 [config.yml.k8s.tpl](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/build/resources/config.yml.k8s.tpl)。 |

---

## 16. 扩展点速查

| 我要新增… | 位置 | 需要实现 |
|---|---|---|
| 一种协议路由 | `drivers/router/<xxx>/` | `IExtenderDriver` + `router/*.matcher` |
| 一种服务发现 | `drivers/discovery/<xxx>/` | `discovery.IDiscovery` |
| 一种鉴权 | `application/auth/<xxx>/` | `IAuth` + `IAuthUser` |
| 一种策略 | `drivers/strategy/<xxx>/` + `drivers/plugins/strategy/<xxx>/` | `IStrategyHandler` + `IFilter` |
| 一种插件 | `drivers/plugins/<xxx>/` | `eocontext.IFilter` |
| 一种输出 | `drivers/output/<xxx>/` | `output.IOutput` |
| 一个 AI Provider | `drivers/ai-provider/<xxx>/` | `IConverterDriver` |
| 一种负载均衡算法 | `upstream/<xxx>/` | `balance.Balance` |

新增后统一在 [driver.go](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/app/apinto/driver.go) 或 [plugin.go](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/app/apinto/plugin.go) 中登记，并在 [profession.go](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/app/apinto/profession.go) 中声明 driver。

---

## 17. 目录索引

```
apinto/
├── app/apinto/           # 进程入口、profession、driver/plugin 注册
├── application/          # 应用（渠道）与鉴权抽象 + 各鉴权实现
├── ai-convert/           # AI Provider 通用转换抽象
├── common/context-label/ # AI/计费共享的 ctx 标签
├── checker/              # 通用匹配器（equal/prefix/regex…）
├── discovery/            # 服务发现通用抽象
├── docs/design/          # 设计文档（本文所在）
├── drivers/              # 全部 driver（router/service/discovery/plugin/output/strategy/ai/…）
│   ├── router/           # HTTP/gRPC/Dubbo2 路由
│   ├── service/          # 上游服务
│   ├── discovery/        # 6 种服务发现
│   ├── app/              # 应用（渠道）
│   ├── plugin-manager/   # 插件链管理
│   ├── plugins/          # 60+ 插件
│   ├── strategy/         # 6 类策略
│   ├── output/           # 9 种输出
│   ├── ai-provider/      # 30+ AI 厂商适配
│   ├── ai-key/, ai-model/# AI 资源
│   ├── pricing-policy/   # 定价策略
│   ├── certs/, gm-certs/ # 证书
│   ├── resources/        # Redis 等资源
│   ├── template/         # 插件模板
│   └── transcode/        # protobuf 编解码
├── entries/              # 请求可读字段（http/metric/monitor）
├── node/                 # 协议上下文实现（http/grpc/dubbo2 + fasthttp-client）
├── output/               # 输出通用抽象
├── plugin/               # 插件接口 & Config 合并
├── price-calcular/       # 计费表达式引擎
├── resources/            # 本地 cache/vector
├── router/               # 路由通用匹配器抽象
├── scope-manager/        # 作用域管理器
├── service/              # Service Skill 定义
├── strategy/             # 策略抽象
├── upstream/             # 负载均衡算法
├── template/             # 模板抽象
├── utils/                # 工具（json/aes/hmac/regex/version…）
├── health-check-http/    # HTTP 健康检查
├── monitor-manager/      # 监控聚合
├── grpc-descriptor/      # gRPC descriptor 转码
└── encoder/              # gzip/br 编码器
```

---

## 18. 参考

- [渠道-资源组多级绑定 设计](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/docs/design/channel-resource-group.md)
- [README](file:///Users/liujian/work/golang/src/github.com/eolinker/apinto/README.md)
- eosc 框架：`github.com/eolinker/eosc`（进程模型 / profession / driver / worker / 配置广播）
