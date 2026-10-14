package credit

import "strings"

// Wildcard 通配符常量
const Wildcard = "*"

// CompiledRule 预编译的高性能代金券规则
type CompiledRule struct {
	ID          string
	Nbf         int64
	Exp         int64
	HasAI       bool
	AllAI       bool
	ProviderAI  map[string]struct{}
	AIProviders []string
	ResourceAI  map[string]struct{}
	AIResources []string

	HasAPI       bool
	AllAPI       bool
	ProviderAPI  map[string]struct{}
	APIProviders []string
	ResourceAPI  map[string]struct{}
	APIResources []string
}

// Name 实现 extractor.IEntity 接口
func (r *CompiledRule) Name() string {
	if r == nil {
		return ""
	}
	return r.ID
}

func isAll(list []string) bool {
	for _, item := range list {
		s := strings.ToLower(strings.TrimSpace(item))
		if s == "all" || s == "*" {
			return true
		}
	}
	return false
}

// CompileRule 将 Config 预编译为快速匹配的内存结构
func CompileRule(id string, conf *Config) *CompiledRule {
	rule := &CompiledRule{
		ID:          id,
		Nbf:         conf.Nbf,
		Exp:         conf.Exp,
		ProviderAI:  make(map[string]struct{}),
		ProviderAPI: make(map[string]struct{}),
		ResourceAI:  make(map[string]struct{}),
		ResourceAPI: make(map[string]struct{}),
	}

	hasAIExplicit := len(conf.Filters.Provider.AI) > 0 || len(conf.Filters.Resource.AI) > 0
	hasAPIExplicit := len(conf.Filters.Provider.API) > 0 || len(conf.Filters.Resource.API) > 0

	// 若两者皆未显式指定，则默认全局通用规则（同时生效于 AI 与 API）
	if !hasAIExplicit && !hasAPIExplicit {
		hasAIExplicit = true
		hasAPIExplicit = true
	}

	if hasAIExplicit {
		rule.HasAI = true
		if len(conf.Filters.Provider.AI) == 0 || isAll(conf.Filters.Provider.AI) {
			rule.AllAI = true
			rule.AIProviders = []string{Wildcard}
		} else {
			for _, p := range conf.Filters.Provider.AI {
				rule.ProviderAI[p] = struct{}{}
				rule.AIProviders = append(rule.AIProviders, p)
			}
		}

		// 资源维度不使用通配符，只支持具体 ResourceID
		for _, r := range conf.Filters.Resource.AI {
			r = strings.TrimSpace(r)
			if r != "" && r != Wildcard && strings.ToLower(r) != "all" {
				rule.ResourceAI[r] = struct{}{}
				rule.AIResources = append(rule.AIResources, r)
			}
		}
	}

	if hasAPIExplicit {
		rule.HasAPI = true
		if len(conf.Filters.Provider.API) == 0 || isAll(conf.Filters.Provider.API) {
			rule.AllAPI = true
			rule.APIProviders = []string{Wildcard}
		} else {
			for _, p := range conf.Filters.Provider.API {
				rule.ProviderAPI[p] = struct{}{}
				rule.APIProviders = append(rule.APIProviders, p)
			}
		}

		// 资源维度不使用通配符，只支持具体 ResourceID
		for _, r := range conf.Filters.Resource.API {
			r = strings.TrimSpace(r)
			if r != "" && r != Wildcard && strings.ToLower(r) != "all" {
				rule.ResourceAPI[r] = struct{}{}
				rule.APIResources = append(rule.APIResources, r)
			}
		}
	}

	return rule
}

// MatchZeroIO 判断规则是否满足当前请求上下文条件
func (r *CompiledRule) MatchZeroIO(targetType, provider, resource string, now int64) bool {
	// 1. 时间窗口有效性：nbf <= now < exp
	if r.Nbf > 0 && now < r.Nbf {
		return false
	}
	if r.Exp > 0 && now >= r.Exp {
		return false
	}

	targetType = strings.ToLower(targetType)
	switch targetType {
	case "ai":
		if !r.HasAI {
			return false
		}
		// 供应商匹配: AllAI 或包含当前 provider
		if !r.AllAI {
			if _, ok := r.ProviderAI[provider]; !ok {
				return false
			}
		}
		// 资源匹配: 必须精准匹配具体 resource
		if _, ok := r.ResourceAI[resource]; !ok {
			return false
		}
		return true

	case "api":
		if !r.HasAPI {
			return false
		}
		// 服务匹配: AllAPI 或包含当前 service
		if !r.AllAPI {
			if _, ok := r.ProviderAPI[provider]; !ok {
				return false
			}
		}
		// 资源匹配: 必须精准匹配具体 resource
		if _, ok := r.ResourceAPI[resource]; !ok {
			return false
		}
		return true

	default:
		return false
	}
}
