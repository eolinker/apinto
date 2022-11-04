package visit_strategy

import "github.com/eolinker/apinto/strategy"

type VisitRule string

const (
	// VisitRuleAllow 访问策略访问规则
	VisitRuleAllow  VisitRule = "allow"  //允许访问
	VisitRuleRefuse VisitRule = "refuse" //拒绝访问
)

type Config struct {
	Name        string                `json:"name" skip:"skip"`
	Description string                `json:"description" skip:"skip"`
	Stop        bool                  `json:"stop"`
	Priority    int                   `json:"priority" label:"优先级" description:"1-999"`
	Filters     strategy.FilterConfig `json:"filters" label:"过滤规则"`
	Rule        Rule                  `json:"visit" label:"规则"`
}

type Rule struct {
	VisitRule       VisitRule             `json:"visit_rule" label:"访问规则"`
	InfluenceSphere strategy.FilterConfig `json:"influence_sphere" label:"生效范围"`
	Continue        bool                  `json:"continue" label:"继续匹配其他访问策略"`
}
