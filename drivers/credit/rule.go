package credit

import "strings"

// Wildcard 通配符常量
const Wildcard = "*"

// ProtocolRule 单个协议维度（AI 或 API）的预编译过滤规则
type ProtocolRule struct {
	Enabled   bool
	All       bool
	Providers []string
	Resources []string
}

// CompiledRule 预编译的高性能代金券规则
type CompiledRule struct {
	ID  string
	Nbf int64
	Exp int64
	AI  ProtocolRule
	API ProtocolRule
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

// compileProtocolRule 编译单个协议（AI / API）的过滤规则
func compileProtocolRule(providers, resources []string, defaultAll bool) ProtocolRule {
	if !defaultAll && len(providers) == 0 && len(resources) == 0 {
		return ProtocolRule{}
	}

	pr := ProtocolRule{Enabled: true}
	if defaultAll || isAll(providers) {
		pr.All = true
		pr.Providers = []string{Wildcard}
	} else {
		for _, p := range providers {
			if p = strings.TrimSpace(p); p != "" {
				pr.Providers = append(pr.Providers, p)
			}
		}
	}

	// 资源维度不使用通配符，只支持具体 ResourceID
	for _, r := range resources {
		if r = strings.TrimSpace(r); r != "" && r != Wildcard && strings.ToLower(r) != "all" {
			pr.Resources = append(pr.Resources, r)
		}
	}
	return pr
}

// CompileRule 将 Config 预编译为快速匹配的内存结构
func CompileRule(id string, conf *Config) *CompiledRule {
	hasAI := len(conf.Filters.Provider.AI) > 0 || len(conf.Filters.Resource.AI) > 0
	hasAPI := len(conf.Filters.Provider.API) > 0 || len(conf.Filters.Resource.API) > 0
	defaultAll := !hasAI && !hasAPI

	return &CompiledRule{
		ID:  id,
		Nbf: conf.Nbf,
		Exp: conf.Exp,
		AI:  compileProtocolRule(conf.Filters.Provider.AI, conf.Filters.Resource.AI, defaultAll),
		API: compileProtocolRule(conf.Filters.Provider.API, conf.Filters.Resource.API, defaultAll),
	}
}
