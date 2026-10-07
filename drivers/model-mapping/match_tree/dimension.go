package match_tree

import (
	"strings"

	"github.com/eolinker/apinto/strategy"
	"github.com/eolinker/eosc/eocontext"
)

// IDimension 维度抽象接口
type IDimension interface {
	Label() string
	ExtractConfigKeys(filter strategy.FilterConfig) []string
	ExtractCandidateKeys(ctx eocontext.EoContext) []string
}

// LabelDimension 通用精确 Label 维度（基于上下文 Label，不用兜底策略）
type LabelDimension struct {
	label        string
	valExtractor func(ctx eocontext.EoContext, label string) string
}

// NewLabelDimension 创建通用的精确 Label 维度
func NewLabelDimension(label string) *LabelDimension {
	return &LabelDimension{
		label: strings.TrimSpace(label),
	}
}

// NewCustomLabelDimension 支持自定义上下文提取逻辑的精确维度
func NewCustomLabelDimension(label string, extractor func(ctx eocontext.EoContext, label string) string) *LabelDimension {
	return &LabelDimension{
		label:        strings.TrimSpace(label),
		valExtractor: extractor,
	}
}

func (d *LabelDimension) Label() string {
	return d.label
}

// ExtractConfigKeys 仅提取配置中明确设置的非空 keys，无配置则返回 nil（不作任何通配兜底）
func (d *LabelDimension) ExtractConfigKeys(filter strategy.FilterConfig) []string {
	if filter == nil {
		return nil
	}
	items, ok := filter[d.label]
	if !ok || len(items) == 0 {
		return nil
	}
	res := make([]string, 0, len(items))
	for _, it := range items {
		it = strings.TrimSpace(it)
		if it != "" {
			res = append(res, it)
		}
	}
	if len(res) == 0 {
		return nil
	}
	return res
}

// ExtractCandidateKeys 仅提取上下文中实际存在的非空值，不增加任何通配兜底！
func (d *LabelDimension) ExtractCandidateKeys(ctx eocontext.EoContext) []string {
	if ctx == nil {
		return nil
	}

	var val string
	if d.valExtractor != nil {
		val = d.valExtractor(ctx, d.label)
	} else {
		val = ctx.GetLabel(d.label)
	}

	val = strings.TrimSpace(val)
	if val == "" {
		return nil // 无该 label 值，直接返回空，匹配时直接短路中断
	}

	// 纯精确值：绝无通配兜底
	return []string{val}
}

// GenericMatchTree 通用多维决策树引擎
type GenericMatchTree[T any] struct {
	dimensions []IDimension
	tree       *GenericDimensionTree[T]
}

// NewGenericMatchTree 创建多维决策树
func NewGenericMatchTree[T any](dimensions ...IDimension) *GenericMatchTree[T] {
	return &GenericMatchTree[T]{
		dimensions: dimensions,
		tree:       NewGenericDimensionTree[T](),
	}
}

// NewLabelMatchTree 根据给定的 label 顺序快速创建通用多维决策树
func NewLabelMatchTree[T any](labels ...string) *GenericMatchTree[T] {
	dims := make([]IDimension, 0, len(labels))
	for _, l := range labels {
		dims = append(dims, NewLabelDimension(l))
	}
	return NewGenericMatchTree[T](dims...)
}

func (m *GenericMatchTree[T]) Dimensions() []IDimension {
	return m.dimensions
}

func (m *GenericMatchTree[T]) Tree() *GenericDimensionTree[T] {
	return m.tree
}

// GenerateDimensionPaths 计算多维笛卡尔积所有组合路径；任一维度缺失则返回 nil
func (m *GenericMatchTree[T]) GenerateDimensionPaths(filter strategy.FilterConfig) [][]string {
	if len(m.dimensions) == 0 {
		return nil
	}

	dimKeys := make([][]string, len(m.dimensions))
	for i, dim := range m.dimensions {
		keys := dim.ExtractConfigKeys(filter)
		if len(keys) == 0 {
			// 该维度未配置有效值，不满足强契约全维度要求，直接返回空路径
			return nil
		}
		dimKeys[i] = keys
	}

	// 组合 N 维笛卡尔积路径
	paths := [][]string{{}}
	for _, keys := range dimKeys {
		var nextPaths [][]string
		for _, p := range paths {
			for _, k := range keys {
				newP := make([]string, len(p)+1)
				copy(newP, p)
				newP[len(p)] = k
				nextPaths = append(nextPaths, newP)
			}
		}
		paths = nextPaths
	}
	return paths
}

// Match 执行精确组合搜索；若任一维度上下文中无值，直接短路返回 nil
func (m *GenericMatchTree[T]) Match(ctx eocontext.EoContext) []SearchResult[T] {
	if len(m.dimensions) == 0 {
		return nil
	}

	queryKeys := make([][]string, len(m.dimensions))
	for i, dim := range m.dimensions {
		keys := dim.ExtractCandidateKeys(ctx)
		if len(keys) == 0 {
			// 当前维度上下文中无有效值，直接短路退出
			return nil
		}
		queryKeys[i] = keys
	}

	return m.tree.Search(queryKeys)
}
