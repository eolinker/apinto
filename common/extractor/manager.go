package extractor

import (
	"sync"

	"github.com/eolinker/eosc/eocontext"
)

// ExtractorResult 包装按顺序提取到的每个维度的结果
type ExtractorResult[T IEntity] struct {
	Dimension string
	Items     []T
}

// Manager 管理多个维度的提取器，支持按指定/默认顺序依次提取
type Manager[T IEntity] struct {
	locker     sync.RWMutex
	extractors map[string]IExtractor[T]
	sorted     []string
}

// NewManager 创建提取器管理器
func NewManager[T IEntity](defaultSorted ...string) *Manager[T] {
	return &Manager[T]{
		extractors: make(map[string]IExtractor[T]),
		sorted:     defaultSorted,
	}
}

// GetExtractor 获取指定维度名称的提取器
func (m *Manager[T]) GetExtractor(name string) (IExtractor[T], bool) {
	m.locker.RLock()
	defer m.locker.RUnlock()
	ext, ok := m.extractors[name]
	return ext, ok
}

// SetExtractor 设置指定维度名称的提取器
func (m *Manager[T]) SetExtractor(name string, extractor IExtractor[T]) {
	m.locker.Lock()
	defer m.locker.Unlock()
	m.extractors[name] = extractor
}

// DelExtractor 移除指定维度名称的提取器
func (m *Manager[T]) DelExtractor(name string) {
	m.locker.Lock()
	defer m.locker.Unlock()
	delete(m.extractors, name)
}

// List 按照优先级顺序列出所有提取器
func (m *Manager[T]) List(sorted []string) []IExtractor[T] {
	m.locker.RLock()
	defer m.locker.RUnlock()

	s := m.sorted
	if len(sorted) > 0 {
		s = sorted
	}

	seen := make(map[string]bool, len(s))
	result := make([]IExtractor[T], 0, len(s))

	// 1. 优先按指定/默认顺序添加
	for _, name := range s {
		seen[name] = true
		if ext, has := m.extractors[name]; has {
			result = append(result, ext)
		}
	}

	// 2. 补充其余未包含的非全局提取器
	for name, ext := range m.extractors {
		if !seen[name] && name != "" {
			result = append(result, ext)
		}
	}

	// 3. 通用兜底 extractor 永远排在最后
	if !seen[""] {
		if ext, has := m.extractors[""]; has {
			result = append(result, ext)
		}
	}

	return result
}

// GetByOrder 按照给定的维度顺序（若未指定则使用默认顺序），依次从各维度 Extractor 提取匹配结果
func (m *Manager[T]) GetByOrder(ctx eocontext.EoContext, dims ...string) []ExtractorResult[T] {
	extractors := m.List(dims)
	if len(extractors) == 0 {
		return nil
	}

	results := make([]ExtractorResult[T], 0, len(extractors))
	for _, ext := range extractors {
		if items, ok := ext.Get(ctx); ok && len(items) > 0 {
			results = append(results, ExtractorResult[T]{
				Dimension: ext.Name(),
				Items:     items,
			})
		}
	}
	return results
}

// GetByValuesOrder 按照给定的维度与其具体值映射，按指定维度顺序依次从各 Extractor 提取匹配项
func (m *Manager[T]) GetByValuesOrder(vals map[string]string, dims ...string) []ExtractorResult[T] {
	extractors := m.List(dims)
	if len(extractors) == 0 {
		return nil
	}

	results := make([]ExtractorResult[T], 0, len(extractors))
	for _, ext := range extractors {
		val := vals[ext.Name()]
		if items, ok := ext.GetByValue(val); ok && len(items) > 0 {
			results = append(results, ExtractorResult[T]{
				Dimension: ext.Name(),
				Items:     items,
			})
		}
	}
	return results
}
