package model_mapping

import (
	"github.com/eolinker/apinto/checker"
	"github.com/eolinker/apinto/strategy"
	"github.com/eolinker/eosc/eocontext"
)

// IFilterTree 过滤匹配树接口
type IFilterTree interface {
	Match(ctx eocontext.EoContext, keys ...string) bool
	SubTrees() []*KeySubTree
	GetSubTree(key string) (*KeySubTree, bool)
	Keys() []string
}

// FilterTree 是由多个 KeySubTree 构成的规则匹配整树
type FilterTree struct {
	subTrees   []*KeySubTree
	subTreeMap map[string]*KeySubTree
	keys       []string
}

// NewFilterTree 根据配置的 filters 规则构建匹配树
// 每个 key 将作为一个独立的 KeySubTree
func NewFilterTree(config strategy.FilterConfig) (*FilterTree, error) {
	subTrees := make([]*KeySubTree, 0, len(config))
	subTreeMap := make(map[string]*KeySubTree, len(config))
	keys := make([]string, 0, len(config))

	for key, patterns := range config {
		if len(patterns) == 0 {
			continue
		}
		subTree, err := NewKeySubTree(key, patterns)
		if err != nil {
			return nil, err
		}
		subTrees = append(subTrees, subTree)
		subTreeMap[key] = subTree
		keys = append(keys, key)
	}

	return &FilterTree{
		subTrees:   subTrees,
		subTreeMap: subTreeMap,
		keys:       keys,
	}, nil
}

// SubTrees 返回所有的 KeySubTree 子树切片
func (t *FilterTree) SubTrees() []*KeySubTree {
	return t.subTrees
}

// Keys 返回所有已构建子树的 key 列表
func (t *FilterTree) Keys() []string {
	return t.keys
}

// GetSubTree 根据 key 获取对应的子树
func (t *FilterTree) GetSubTree(key string) (*KeySubTree, bool) {
	st, ok := t.subTreeMap[key]
	return st, ok
}

// Match 从上下文中寻找对应 key 的 label 值并在对应的子树中匹配。
// 参数 keys 用于指定本次匹配只检查哪几颗子树，避免无差别匹配所有 label 带来额外的检索开销：
// 1. 若指定了 keys，则仅在指定的子树中匹配（未配置该 key 子树的视为不设限跳过，不在 keys 列表中的子树不参与匹配）；
// 2. 若未指定 keys（len(keys) == 0），则按整树全匹配原则，要求整树中所有已配置的子树全部匹配成功才算通过。
func (t *FilterTree) Match(ctx eocontext.EoContext, keys ...string) bool {
	if ctx == nil {
		return false
	}

	// 1. 指定了匹配哪几颗子树
	if len(keys) > 0 {
		var (
			labels       map[string]string
			labelsLoaded bool
		)

		for _, key := range keys {
			subTree, ok := t.subTreeMap[key]
			if !ok {
				// 策略树未配置该 key 对应的子树，表示该维度不设限，直接通过
				continue
			}

			// 延迟加载 context 的 labels 集合
			if !labelsLoaded {
				labels = ctx.Labels()
				labelsLoaded = true
			}

			var (
				val string
				has bool
			)
			if labels != nil {
				val, has = labels[key]
			}
			if !has {
				val = ctx.GetLabel(key)
				has = val != ""
			}

			// 只要指定匹配的任意一颗子树不匹配，立即短路失败
			if !subTree.Match(val, has) {
				return false
			}
		}

		return true
	}

	// 2. 未指定 keys，默认整树所有已配置子树全匹配
	if len(t.subTrees) == 0 {
		return true
	}

	labels := ctx.Labels()
	for _, subTree := range t.subTrees {
		key := subTree.Key()
		var (
			val string
			has bool
		)
		if labels != nil {
			val, has = labels[key]
		}
		if !has {
			val = ctx.GetLabel(key)
			has = val != ""
		}

		if !subTree.Match(val, has) {
			return false
		}
	}

	return true
}

// KeySubTree 以特定的上下文 key 为根的子树
type KeySubTree struct {
	key         string
	allowAll    bool
	exactValues map[string]struct{}
	checkers    []checker.Checker
	rawPatterns []string
}

// NewKeySubTree 构建单个 key 的匹配子树
func NewKeySubTree(key string, patterns []string) (*KeySubTree, error) {
	subTree := &KeySubTree{
		key:         key,
		exactValues: make(map[string]struct{}),
		checkers:    make([]checker.Checker, 0, len(patterns)),
		rawPatterns: patterns,
	}

	for _, p := range patterns {
		chk, err := checker.Parse(p)
		if err != nil {
			return nil, err
		}

		switch chk.CheckType() {
		case checker.CheckTypeAll:
			subTree.allowAll = true
		case checker.CheckTypeEqual:
			subTree.exactValues[chk.Value()] = struct{}{}
		default:
			subTree.checkers = append(subTree.checkers, chk)
		}
	}

	return subTree, nil
}

// Key 返回该子树绑定的上下文 key 名称
func (k *KeySubTree) Key() string {
	return k.key
}

// Patterns 返回该子树包含的原始模式列表
func (k *KeySubTree) Patterns() []string {
	return k.rawPatterns
}

// Match 在该 key 的子树内匹配传入的 label 值
// 只要满足精确匹配、任意通配符或任意一条 Checker 规则即为该子树匹配成功
func (k *KeySubTree) Match(val string, has bool) bool {
	// 1. 如果包含全匹配通配符，只要存在该 label（或 has 为 true）即命中
	if k.allowAll {
		return has
	}

	if !has {
		return false
	}

	// 2. 优先进行 O(1) 的精确值哈希比对
	if _, ok := k.exactValues[val]; ok {
		return true
	}

	// 3. 遍历高级模式检查器（前缀、后缀、正则等）
	for _, chk := range k.checkers {
		if chk.Check(val, has) {
			return true
		}
	}

	return false
}
