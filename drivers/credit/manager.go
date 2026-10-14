package credit

import (
	"sort"
	"strings"
	"sync"

	"github.com/eolinker/apinto/common/extractor"
)

var defaultManager = NewManager()

// Manager 本地内存代金券规则管理器与匹配引擎
type Manager struct {
	mu         sync.Mutex
	rules      map[string]*CompiledRule
	aiManager  *extractor.Manager[*CompiledRule]
	apiManager *extractor.Manager[*CompiledRule]
}

var defaultSorted = []string{
	"provider",
	"resource",
}

// NewManager 创建规则管理器实例，按顺序支持 provider 与 resource 自定义维度
func NewManager() *Manager {
	m := &Manager{
		rules:      make(map[string]*CompiledRule),
		aiManager:  extractor.NewManager[*CompiledRule](defaultSorted...),
		apiManager: extractor.NewManager[*CompiledRule](defaultSorted...),
	}

	setExtractors(m.aiManager, defaultSorted)
	setExtractors(m.apiManager, defaultSorted)
	return m
}

func setExtractors(manager *extractor.Manager[*CompiledRule], names []string) {
	for _, name := range names {
		manager.SetExtractor(name, extractor.NewExtractor[*CompiledRule](name))
	}
}

// SetRule 注册或更新一条规则，按多维度分别注入 Extractor
func (m *Manager) SetRule(id string, rule *CompiledRule) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.rules[id] = rule

	// 1. 同步更新 AI 场景多维度 Extractor
	if pExt, ok := m.aiManager.GetExtractor("provider"); ok {
		if rule.HasAI {
			provs := rule.AIProviders
			if rule.AllAI || len(provs) == 0 {
				provs = []string{Wildcard}
			}
			pExt.Set(id, provs, []*CompiledRule{rule})
		} else {
			pExt.Del(id)
		}
	}
	if rExt, ok := m.aiManager.GetExtractor("resource"); ok {
		if rule.HasAI && len(rule.AIResources) > 0 {
			rExt.Set(id, rule.AIResources, []*CompiledRule{rule})
		} else {
			rExt.Del(id)
		}
	}

	// 2. 同步更新 API 场景多维度 Extractor
	if pExt, ok := m.apiManager.GetExtractor("provider"); ok {
		if rule.HasAPI {
			provs := rule.APIProviders
			if rule.AllAPI || len(provs) == 0 {
				provs = []string{Wildcard}
			}
			pExt.Set(id, provs, []*CompiledRule{rule})
		} else {
			pExt.Del(id)
		}
	}
	if rExt, ok := m.apiManager.GetExtractor("resource"); ok {
		if rule.HasAPI && len(rule.APIResources) > 0 {
			rExt.Set(id, rule.APIResources, []*CompiledRule{rule})
		} else {
			rExt.Del(id)
		}
	}
}

// DelRule 移除一条规则，从所有 Extractor 级联删除
func (m *Manager) DelRule(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.rules, id)

	if pExt, ok := m.aiManager.GetExtractor("provider"); ok {
		pExt.Del(id)
	}
	if rExt, ok := m.aiManager.GetExtractor("resource"); ok {
		rExt.Del(id)
	}

	if pExt, ok := m.apiManager.GetExtractor("provider"); ok {
		pExt.Del(id)
	}
	if rExt, ok := m.apiManager.GetExtractor("resource"); ok {
		rExt.Del(id)
	}
}

// MatchCandidateRules 按照自定义维度顺序（先找 provider，再找 resource）获取候选规则列表并求交集，按 Exp ASC 正序返回
func (m *Manager) MatchCandidateRules(targetType, provider, resource string, now int64) []*CompiledRule {
	if resource == "" || resource == Wildcard {
		return nil
	}

	targetType = strings.ToLower(targetType)
	var mgr *extractor.Manager[*CompiledRule]
	switch targetType {
	case "ai":
		mgr = m.aiManager
	case "api":
		mgr = m.apiManager
	default:
		return nil
	}

	// 1. 按照自定义维度顺序依次提取候选列表：先找 provider，再找 resource
	vals := map[string]string{
		"provider": provider,
		"resource": resource,
	}
	dimResults := mgr.GetByValuesOrder(vals, defaultSorted...)
	if len(dimResults) < 2 {
		return nil
	}

	var providerRules, resourceRules []*CompiledRule
	for _, res := range dimResults {
		switch res.Dimension {
		case "provider":
			providerRules = res.Items
		case "resource":
			resourceRules = res.Items
		}
	}

	if len(providerRules) == 0 || len(resourceRules) == 0 {
		return nil
	}

	// 2. 取两个维度的交集，并校验时效（Nbf 与 Exp）
	provMap := make(map[string]struct{}, len(providerRules))
	for _, r := range providerRules {
		provMap[r.ID] = struct{}{}
	}

	var matched []*CompiledRule
	for _, r := range resourceRules {
		if _, ok := provMap[r.ID]; !ok {
			continue
		}
		if r.Nbf > 0 && now < r.Nbf {
			continue
		}
		if r.Exp > 0 && now >= r.Exp {
			continue
		}
		matched = append(matched, r)
	}

	if len(matched) <= 1 {
		return matched
	}

	// 3. 排序：按 Exp ASC 正序排列
	sort.Slice(matched, func(i, j int) bool {
		if matched[i].Exp == matched[j].Exp {
			return matched[i].ID < matched[j].ID
		}
		if matched[i].Exp == 0 {
			return false
		}
		if matched[j].Exp == 0 {
			return true
		}
		return matched[i].Exp < matched[j].Exp
	})
	return matched
}

// SetRule 向默认规则管理器注册规则
func SetRule(id string, rule *CompiledRule) {
	defaultManager.SetRule(id, rule)
}

// DelRule 从默认规则管理器注销规则
func DelRule(id string) {
	defaultManager.DelRule(id)
}

// MatchCandidateRules 快捷匹配默认规则管理器中的规则
func MatchCandidateRules(targetType, provider, resource string, now int64) []*CompiledRule {
	return defaultManager.MatchCandidateRules(targetType, provider, resource, now)
}
