package model_mapping

import (
	"sort"
	"sync"

	"github.com/eolinker/apinto/drivers/model-mapping/match_tree"
	"github.com/eolinker/apinto/strategy"
	"github.com/eolinker/eosc/eocontext"
)

var (
	defaultManager = NewManager("consumer", "model")
)

// DefaultManager 获取全局单例管理器
func DefaultManager() *Manager {
	return defaultManager
}

// SetStrategy 全局快捷注册或更新策略
func SetStrategy(id string, handler IHandler, filters strategy.FilterConfig) {
	defaultManager.Set(id, handler, filters)
}

// DelStrategy 全局快捷注销策略
func DelStrategy(id string) {
	defaultManager.Del(id)
}

// MatchBest 全局快捷匹配函数：纯精确无兜底组合匹配，按 Priority 裁决最佳策略
func MatchBest(ctx eocontext.EoContext) (IHandler, bool) {
	return defaultManager.MatchBest(ctx)
}

// IndexedStrategy 包含策略主体、叶子路径及 Handler
type IndexedStrategy struct {
	ID             string
	Priority       int
	Handler        IHandler
	DimensionPaths [][]string
}

// Manager 策略全局管理器
type Manager struct {
	mu         sync.RWMutex
	strategies map[string]*IndexedStrategy
	matchTree  *match_tree.GenericMatchTree[*IndexedStrategy]
}

// NewManager 创建策略管理器，按指定的 labels 顺序初始化通用精确决策树
func NewManager(labels ...string) *Manager {
	if len(labels) == 0 {
		labels = []string{"consumer", "model"}
	}
	return &Manager{
		strategies: make(map[string]*IndexedStrategy),
		matchTree:  match_tree.NewLabelMatchTree[*IndexedStrategy](labels...),
	}
}

// Set 注册或更新策略
func (m *Manager) Set(id string, handler IHandler, filter strategy.FilterConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 1. 若旧策略存在，精准清理旧路径
	if old, exists := m.strategies[id]; exists {
		for _, path := range old.DimensionPaths {
			m.matchTree.Tree().Remove(path, func(item *IndexedStrategy) bool {
				return item.ID == id
			})
		}
	}

	// 2. 生成纯精确笛卡尔积路径（若 filters 未同时包含合法 consumer 和 model，则生成为空）
	paths := m.matchTree.GenerateDimensionPaths(filter)
	if len(paths) == 0 {
		delete(m.strategies, id)
		return
	}

	indexed := &IndexedStrategy{
		ID:             id,
		Priority:       handler.Priority(),
		Handler:        handler,
		DimensionPaths: paths,
	}

	// 3. 存入主表并挂载至多维决策树对应叶子节点
	m.strategies[id] = indexed
	for _, path := range paths {
		m.matchTree.Tree().Insert(path, indexed)
	}
}

// Del 注销策略
func (m *Manager) Del(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if indexed, exists := m.strategies[id]; exists {
		for _, path := range indexed.DimensionPaths {
			m.matchTree.Tree().Remove(path, func(item *IndexedStrategy) bool {
				return item.ID == id
			})
		}
		delete(m.strategies, id)
	}
}

// MatchBest 纯精确匹配：快速沿 [consumer, model] 查找命中，并按 Priority 裁决（数值越小越优先）
func (m *Manager) MatchBest(ctx eocontext.EoContext) (IHandler, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	matchedList := m.matchTree.Match(ctx)
	if len(matchedList) == 0 {
		return nil, false
	}

	visited := make(map[string]bool)
	candidates := make([]*IndexedStrategy, 0, len(matchedList))

	for _, srs := range matchedList {
		for _, item := range srs.Items {
			if !visited[item.ID] {
				visited[item.ID] = true
				candidates = append(candidates, item)
			}
		}
	}

	if len(candidates) == 0 {
		return nil, false
	}

	// Priority 越小越优先
	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].Priority < candidates[j].Priority
	})

	return candidates[0].Handler, true
}
