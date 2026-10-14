package extractor

import (
	"sync"

	"github.com/eolinker/eosc/eocontext"
)

// IEntity 提取器管理的数据实体接口约束，需提供唯一标识 Name
type IEntity interface {
	Name() string
}

// IExtractor 通用维度提取器接口
type IExtractor[T IEntity] interface {
	Name() string
	Get(ctx eocontext.EoContext) ([]T, bool)
	GetByValue(val string) ([]T, bool)
	Set(id string, bind []string, handlers []T)
	Del(id string)
}

// Extractor 基于特定维度 label（如 resource、provider，或 "" 全局通用）管理 Entity
type Extractor[T IEntity] struct {
	locker     sync.RWMutex
	name       string              // 维度 label 名，如 "resource"、"provider"，"" 表示全局通用维度
	values     map[string][]T      // 精确值索引: exactValue -> handlers
	globals    []T                 // 当 name == "" 时使用的全局通用 handlers
	idKeys     map[string][]string // id -> 该策略添加到的 key 列表
	idHandlers map[string][]T      // id -> 该策略的 handlers
}

// NewExtractor 创建指定维度的提取器实例
func NewExtractor[T IEntity](name string) *Extractor[T] {
	return &Extractor[T]{
		name:       name,
		values:     make(map[string][]T),
		globals:    make([]T, 0),
		idKeys:     make(map[string][]string),
		idHandlers: make(map[string][]T),
	}
}

// Name 返回维度名称
func (e *Extractor[T]) Name() string {
	return e.name
}

// Get 从当前 Extractor 根据上下文 label 检索匹配的列表
func (e *Extractor[T]) Get(ctx eocontext.EoContext) ([]T, bool) {
	if e.name == "" {
		return e.GetByValue("")
	}

	if ctx == nil {
		return nil, false
	}

	val := ctx.GetLabel(e.name)
	if val == "" {
		return nil, false
	}

	return e.GetByValue(val)
}

// GetByValue 根据具体维度值检索匹配的列表（支持通配合并与精确值匹配）
func (e *Extractor[T]) GetByValue(val string) ([]T, bool) {
	e.locker.RLock()
	defer e.locker.RUnlock()

	if e.name == "" {
		if len(e.globals) == 0 {
			return nil, false
		}
		var matched []T
		seen := make(map[string]bool, len(e.globals))
		for _, h := range e.globals {
			if any(h) != nil && !seen[h.Name()] {
				seen[h.Name()] = true
				matched = append(matched, h)
			}
		}
		if len(matched) == 0 {
			return nil, false
		}
		return matched, true
	}

	var candidates []T
	if val != "" && val != "*" {
		if list, ok := e.values[val]; ok && len(list) > 0 {
			candidates = append(candidates, list...)
		}
	}

	// 若存在通配符 "*"，一并合并
	if wildcards, ok := e.values["*"]; ok && len(wildcards) > 0 {
		candidates = append(candidates, wildcards...)
	}

	if len(candidates) == 0 {
		return nil, false
	}

	var matched []T
	seen := make(map[string]bool, len(candidates))
	for _, h := range candidates {
		if any(h) != nil && !seen[h.Name()] {
			seen[h.Name()] = true
			matched = append(matched, h)
		}
	}

	if len(matched) == 0 {
		return nil, false
	}

	return matched, true
}

// Set 注册策略及其绑定的规则（以策略/规则 id 为主键）
func (e *Extractor[T]) Set(id string, bind []string, handlers []T) {
	if id == "" {
		return
	}

	e.locker.Lock()
	defer e.locker.Unlock()

	// 1. 若旧的存在，先清理
	e.delInternal(id)

	if len(handlers) == 0 {
		return
	}

	e.idHandlers[id] = handlers

	if e.name == "" {
		e.globals = append(e.globals, handlers...)
		return
	}

	if len(bind) == 0 {
		return
	}

	keys := make([]string, 0, len(bind))
	for _, b := range bind {
		if b == "" {
			continue
		}
		e.values[b] = append(e.values[b], handlers...)
		keys = append(keys, b)
	}

	if len(keys) > 0 {
		e.idKeys[id] = keys
	}
}

// Del 删除指定策略 id
func (e *Extractor[T]) Del(id string) {
	if id == "" {
		return
	}
	e.locker.Lock()
	defer e.locker.Unlock()
	e.delInternal(id)
}

func (e *Extractor[T]) delInternal(id string) {
	oldHandlers, exists := e.idHandlers[id]
	if !exists {
		return
	}

	targetNames := make(map[string]bool, len(oldHandlers))
	for _, h := range oldHandlers {
		if any(h) != nil {
			targetNames[h.Name()] = true
		}
	}

	if e.name == "" {
		filtered := make([]T, 0, len(e.globals))
		for _, h := range e.globals {
			if any(h) != nil && !targetNames[h.Name()] {
				filtered = append(filtered, h)
			}
		}
		e.globals = filtered
	} else {
		if keys, ok := e.idKeys[id]; ok {
			for _, key := range keys {
				if list, existsList := e.values[key]; existsList {
					filtered := make([]T, 0, len(list))
					for _, h := range list {
						if any(h) != nil && !targetNames[h.Name()] {
							filtered = append(filtered, h)
						}
					}
					if len(filtered) == 0 {
						delete(e.values, key)
					} else {
						e.values[key] = filtered
					}
				}
			}
		}
	}

	delete(e.idKeys, id)
	delete(e.idHandlers, id)
}
