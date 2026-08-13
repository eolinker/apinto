package quota_limiting_strategy

import (
	"fmt"
	context_label "github.com/eolinker/apinto/common/context-label"
	"github.com/eolinker/apinto/drivers/strategy"
	
	"github.com/eolinker/apinto/utils/response"
)

var quotaKeyPre = context_label.NewKeyGenerator("{product}:quota-limiting:{strategy}:{target_type}:{application}")

// QuotaConfig 配额配置集合
type QuotaConfig struct {
	Request    QuotaRule `json:"request" label:"请求配额"`        // 请求次数配额
	TotalToken QuotaRule `json:"total_token" label:"总Token配额"` // 总Token配额
	Amount     QuotaRule `json:"amount" label:"总金额配额"`       // 总金额配额
}

type Filter struct {
	Type    string   `json:"type"`
	All     bool     `json:"all"`     // 是否全选
	Parents []string `json:"parents"` // 全选的服务/供应商 ID 列表
	Items   []string `json:"items"`   // 资源 ID 列表
}

// FiltersConfig 过滤条件
type FiltersConfig struct {
	Resource Filter `json:"resource"` // 资源过滤
	Target   Filter `json:"target"`   // 目标对象
	Tenant   string `json:"tenant"`   // 租户 ID
}

var _ strategy.IConfig = (*Config)(nil)

// Config 策略核心配置结构
type Config struct {
	Na       string             `json:"name"`
	Desc     string             `json:"desc"`
	Quota    QuotaConfig        `json:"quota"`
	Filters  FiltersConfig      `json:"filters"`
	Response *response.Response `json:"response"` // 拦截超额时的响应配置
	PreKey   string             `json:"pre_key"`
}

func (c *Config) Name() string {
	return c.Na
}

func (c *Config) Description() string {
	return c.Desc
}

func (c *Config) Check() error {
	return checkConfig(c)
}

func checkConfig(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}
	
	return nil
}
