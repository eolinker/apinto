package billing

import (
	"fmt"
	"time"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/pricing"
	"github.com/eolinker/apinto/resources"
	scope_manager "github.com/eolinker/apinto/scope-manager"
	"github.com/eolinker/eosc"
)

// Config 计费拦截插件配置。
//
// 字段语义：
//   - Provider:        默认供应商，请求字段 / 上下文标签均缺失时作为兜底
//   - Resource:        默认资源（AI=模型名，API=接口名）
//   - Phase:           计费阶段，留空表示同步/自动
//   - Cache:           Redis 缓存资源 ID，可选；缺失时余额扣减/任务防重失效自动降级
//   - TaskTTL:         异步任务缓存过期时间（秒），<=0 走默认值 24h
//   - ResultMatch:     响应结果匹配规则，匹配成功才计费
//   - MatchRules:      请求匹配规则，可基于 method/path/header 决定是否拦截
//   - RequestFields:   请求体字段提取规则（JSONPath）
//   - ResponseFields:  响应体字段提取规则（JSONPath）
//   - EnableBalance:   是否启用余额扣减（默认开启）
type Config struct {
	Provider       string               `json:"provider" label:"供应商" description:"默认供应商，请求字段/上下文均缺失时使用"`
	Resource       string               `json:"resource" label:"资源" description:"默认资源（AI=模型名，API=接口名）"`
	Phase          pricing.PricingPhase `json:"phase" label:"计费阶段" description:"空(同步/自动检测)、submit(提交预扣)、query(查询实扣)" enum:",submit,query"`
	Cache          eosc.RequireId       `json:"cache" label:"缓存资源" skill:"github.com/eolinker/apinto/resources.resources.ICache" required:"false" description:"Redis 缓存资源 ID（可选）"`
	TaskTTL        int                  `json:"task_ttl" label:"任务缓存过期时间(秒)" default:"86400" description:"异步任务缓存过期时间，<=0 走默认 24h"`
	ResultMatch    *pricing.ResultMatch `json:"result_match,omitempty" label:"响应结果匹配" description:"响应匹配规则，全部命中才计费"`
	MatchRules     *MatchRules          `json:"match_rules,omitempty" label:"请求匹配规则" description:"请求匹配规则，未命中则跳过计费"`
	RequestFields  map[string]string    `json:"request_fields" label:"请求字段提取" description:"请求体字段 JSONPath 映射"`
	ResponseFields map[string]string    `json:"response_fields" label:"响应字段提取" description:"响应体字段 JSONPath 映射"`
	EnableBalance  *bool                `json:"enable_balance,omitempty" label:"启用余额扣减" description:"是否启用余额扣减，nil 视为开启"`
}

// checkConfig 校验配置参数：
//  1. 类型校验
//  2. 至少配置 RequestFields 或 ResponseFields 之一
//  3. 修正 TaskTTL 负值
//  4. MatchRules.PathPattern 提前编译以暴露非法正则
func checkConfig(v interface{}) (*Config, error) {
	conf, ok := v.(*Config)
	if !ok {
		return nil, eosc.ErrorConfigType
	}

	if len(conf.ResponseFields) == 0 && len(conf.RequestFields) == 0 {
		return nil, fmt.Errorf("at least request_fields or response_fields must be defined")
	}

	if conf.TaskTTL < 0 {
		conf.TaskTTL = 0
	}

	if conf.MatchRules != nil && conf.MatchRules.PathPattern != "" {
		if _, err := newMatchRulesMatcher(conf.MatchRules); err != nil {
			return nil, fmt.Errorf("invalid match_rules: %w", err)
		}
	}

	return conf, nil
}

// Create 实例化 executor。
//
// 流程：校验 → 字段提取器 → 匹配规则 → 结果匹配器 → TaskTTL 兜底 → 缓存代理 → 任务/余额执行器 → 组装 worker。
func Create(id, name string, v *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	cfg, err := checkConfig(v)
	if err != nil {
		return nil, err
	}

	extractor, err := NewFieldExtractor(cfg.RequestFields, cfg.ResponseFields)
	if err != nil {
		return nil, fmt.Errorf("create field extractor: %w", err)
	}

	var matcher *matchRulesMatcher
	if cfg.MatchRules != nil {
		matcher, err = newMatchRulesMatcher(cfg.MatchRules)
		if err != nil {
			return nil, fmt.Errorf("create match rules: %w", err)
		}
	}

	var rMatcher *resultMatcher
	if cfg.ResultMatch != nil {
		// 当只设置了 SuccessField 而未配 ResponseFields 时，自动构造一个独立提取器
		resultExtractor := extractor
		if cfg.ResultMatch.SuccessField != "" && len(cfg.ResponseFields) == 0 {
			resultExtractor, err = NewFieldExtractor(nil, map[string]string{
				cfg.ResultMatch.SuccessField: cfg.ResultMatch.SuccessField,
			})
			if err != nil {
				return nil, fmt.Errorf("create result match extractor: %w", err)
			}
		}
		rMatcher, err = newResultMatcher(cfg.ResultMatch, resultExtractor)
		if err != nil {
			return nil, fmt.Errorf("create result matcher: %w", err)
		}
	}

	ttl := time.Duration(cfg.TaskTTL) * time.Second
	if ttl <= 0 {
		ttl = pricing.DefaultTaskTTL
	}

	cache := scope_manager.Auto[resources.ICache](string(cfg.Cache), "redis")
	taskExecutor := NewRedisTaskExecutor(cache, ttl)

	enableBalance := true
	if cfg.EnableBalance != nil {
		enableBalance = *cfg.EnableBalance
	}
	var balanceManager pricing.IBalanceManager
	if enableBalance {
		balanceManager = NewRedisBalanceManager(cache)
	}

	w := &executor{
		WorkerBase:     drivers.Worker(id, name),
		provider:       cfg.Provider,
		resource:       cfg.Resource,
		phase:          cfg.Phase,
		extractor:      extractor,
		matchRules:     matcher,
		resultMatcher:  rMatcher,
		taskExecutor:   taskExecutor,
		balanceManager: balanceManager,
	}
	return w, nil
}
