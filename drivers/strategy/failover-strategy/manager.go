package failover_strategy

import "github.com/eolinker/eosc"

var (
	manager = NewManager()
)

// defaultSorted 定义默认的维度匹配顺序：
// 1. resource 维度（精确到 provider/model）
// 2. provider 维度
var defaultSorted = []string{
	"resource",
	"provider",
}

type Manager struct {
	extractors eosc.Untyped[string, IExtractor]
}

func NewManager() *Manager {
	return &Manager{
		extractors: eosc.BuildUntyped[string, IExtractor](),
	}
}

func (m *Manager) GetExtractor(name string) (IExtractor, bool) {
	return m.extractors.Get(name)
}

func (m *Manager) setExtractor(name string, extractor IExtractor) {
	m.extractors.Set(name, extractor)
}

func (m *Manager) List(sorted []string) []IExtractor {
	s := defaultSorted
	if len(sorted) > 0 {
		s = sorted
	}

	seen := make(map[string]bool, len(s))
	result := make([]IExtractor, 0, len(s))

	// 1. 优先按指定/默认顺序添加
	for _, name := range s {
		seen[name] = true
		if ext, has := m.extractors.Get(name); has {
			result = append(result, ext)
		}
	}

	// 2. 将其余可能注册的 extractor（除通用兜底 "" 之外）补充进列表
	for _, name := range m.extractors.Keys() {
		if !seen[name] && name != "" {
			if ext, has := m.extractors.Get(name); has {
				result = append(result, ext)
			}
		}
	}

	// 3. 通用兜底 extractor 永远排在最后
	if !seen[""] {
		if ext, has := m.extractors.Get(""); has {
			result = append(result, ext)
		}
	}

	return result
}

func GetExtractor(name string) (IExtractor, bool) {
	return manager.GetExtractor(name)
}

func ListExtractor(sorted []string) []IExtractor {
	return manager.List(sorted)
}

func setExtractor(name string, extractor IExtractor) {
	manager.setExtractor(name, extractor)
}

// RegisterStrategy 注册策略至对应维度 Extractor（可用于测试或动态调度）
func RegisterStrategy(id string, handler IHandler, filters map[string][]string) {
	if handler == nil {
		return
	}
	hasFilters := false
	for dim, vals := range filters {
		if len(vals) == 0 {
			continue
		}
		hasFilters = true
		ex, has := GetExtractor(dim)
		if !has {
			ex = NewExtractor(dim)
			setExtractor(dim, ex)
		}
		ex.Set(id, vals, []IHandler{handler})
	}
	if !hasFilters {
		ex, has := GetExtractor("")
		if !has {
			ex = NewExtractor("")
			setExtractor("", ex)
		}
		ex.Set(id, nil, []IHandler{handler})
	}
}

// UnregisterStrategy 从所有 Extractor 中注销策略
func UnregisterStrategy(id string) {
	for _, dim := range manager.extractors.Keys() {
		if ext, has := manager.extractors.Get(dim); has {
			ext.Del(id)
		}
	}
}

// SetStrategy 注册策略（兼容快捷方式）
func SetStrategy(id string, handler IHandler, filters ...map[string][]string) {
	var f map[string][]string
	if len(filters) > 0 {
		f = filters[0]
	}
	RegisterStrategy(id, handler, f)
}

// DelStrategy 注销策略（兼容快捷方式）
func DelStrategy(id string) {
	UnregisterStrategy(id)
}
