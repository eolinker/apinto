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

	setProtocolRule(m.aiManager, id, rule, rule.AI)
	setProtocolRule(m.apiManager, id, rule, rule.API)
}

func setProtocolRule(m *extractor.Manager[*CompiledRule], id string, rule *CompiledRule, pr ProtocolRule) {
	if pExt, ok := m.GetExtractor("provider"); ok {
		if pr.Enabled && len(pr.Providers) > 0 {
			pExt.Set(id, pr.Providers, []*CompiledRule{rule})
		} else {
			pExt.Del(id)
		}
	}
	if rExt, ok := m.GetExtractor("resource"); ok {
		if pr.Enabled && len(pr.Resources) > 0 {
			rExt.Set(id, pr.Resources, []*CompiledRule{rule})
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

	delRule(m.aiManager, id, defaultSorted)
	delRule(m.apiManager, id, defaultSorted)
}

func delRule(m *extractor.Manager[*CompiledRule], id string, keys []string) {
	for _, key := range keys {
		if ext, ok := m.GetExtractor(key); ok {
			ext.Del(id)
		}
	}
}

// MatchCandidateRules 按照自定义维度顺序（先找 provider，再找 resource）获取候选规则列表（符合任一维度即可），按 Exp ASC 正序返回
func (m *Manager) MatchCandidateRules(targetType, provider, resource string, now int64) []*CompiledRule {
	if (provider == "" || provider == Wildcard) && (resource == "" || resource == Wildcard) {
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
	if len(dimResults) == 0 {
		return nil
	}

	// 2. 合并各维度命中的规则（满足其一即可，按规则 ID 去重），并校验时效（Nbf 与 Exp）
	seen := make(map[string]struct{})
	var matched []*CompiledRule
	for _, res := range dimResults {
		for _, r := range res.Items {
			if _, ok := seen[r.ID]; ok {
				continue
			}
			seen[r.ID] = struct{}{}
			if r.Nbf > 0 && now < r.Nbf {
				continue
			}
			if r.Exp > 0 && now >= r.Exp {
				continue
			}
			matched = append(matched, r)
		}
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
