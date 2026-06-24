package dynamic_billing

import (
	"github.com/eolinker/eosc"
)

// Config 定义了 dynamic-billing 插件的配置结构
type Config struct {
	Cache            eosc.RequireId `json:"cache" label:"缓存资源" skill:"github.com/eolinker/apinto/resources.resources.ICache" required:"false" description:"Redis 缓存资源 ID"`
	ConcurrencyLimit int            `json:"concurrency_limit" label:"默认并发限制" description:"未从上下文获取到并发限制时的默认值" default:"100"`
	EnableBalance    bool           `json:"enable_balance" label:"启用余额扣减" description:"是否启用扣款逻辑" default:"true"`
	TaskKey          string         `json:"task_key" label:"异步任务 Key 模版" default:"dynamic-billing-task:{application}:{resource}" description:"支持{application}, {resource}等占位符"`
	ConcurrencyKey   string         `json:"concurrency_key" label:"并发 Key 模版" default:"dynamic-billing-concurrency:{application}:{resource}" description:"支持{application}, {resource}等占位符"`
	BalanceKey       string         `json:"balance_key" label:"租户余额扣减 Key 模版" default:"balance:{balance_target}" description:"支持{balance_target}等占位符"`
	PriceKey         string         `json:"price_key" label:"资源定价价格 Key 模版" default:"access-resource-price:{application}:{resource}" description:"支持{application}, {resource}等占位符"`
}

// checkConfig 校验并设置默认值
func checkConfig(cfg *Config, workers map[eosc.RequireId]eosc.IWorker) error {
	if cfg.BalanceKey == "" {
		cfg.BalanceKey = "balance:{balance_target}"
	}
	if cfg.PriceKey == "" {
		cfg.PriceKey = "access-resource-price:{application}:{resource}"
	}

	if cfg.TaskKey == "" {
		cfg.TaskKey = "resource-pricing-task:{application}:{resource}"
	}
	if cfg.ConcurrencyKey == "" {
		cfg.ConcurrencyKey = "resource-pricing-concurrency:{application}:{resource}"
	}
	return nil
}
