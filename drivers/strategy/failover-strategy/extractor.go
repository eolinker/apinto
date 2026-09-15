package failover_strategy

import (
	"sync"

	"github.com/eolinker/eosc/eocontext"
)

type IExtractor interface {
	Name() string
	Get(ctx eocontext.EoContext) ([]IHandler, bool)
	Set(id string, bind []string, handlers []IHandler)
	Del(id string)
}

// Extractor 基于特定维度 label（如 resource、provider，或 "" 全局通用）管理灾备策略 Handler，仅进行精确值匹配
type Extractor struct {
	locker     sync.RWMutex
	name       string                // 维度 label 名，如 "resource"、"provider"，"" 表示全局通用维度
	values     map[string][]IHandler // 精确值索引: exactValue -> handlers
	globals    []IHandler            // 当 name == "" 时使用的全局通用 handlers
	idKeys     map[string][]string   // id -> 该策略添加到的 key 列表
	idHandlers map[string][]IHandler // id -> 该策略的 handlers
}

func NewExtractor(name string) *Extractor {
	return &Extractor{
		name:       name,
		values:     make(map[string][]IHandler),
		globals:    make([]IHandler, 0),
		idKeys:     make(map[string][]string),
		idHandlers: make(map[string][]IHandler),
	}
}

func (e *Extractor) Name() string {
	return e.name
}

// Get 从当前 Extractor 检索匹配的 Handler 列表
func (e *Extractor) Get(ctx eocontext.EoContext) ([]IHandler, bool) {
	e.locker.RLock()
	defer e.locker.RUnlock()

	if e.name == "" {
		if len(e.globals) == 0 {
			return nil, false
		}
		var matched []IHandler
		seen := make(map[string]bool, len(e.globals))
		for _, h := range e.globals {
			if h != nil && !seen[h.Name()] {
				seen[h.Name()] = true
				matched = append(matched, h)
			}
		}
		if len(matched) == 0 {
			return nil, false
		}
		return matched, true
	}

	if ctx == nil {
		return nil, false
	}

	val := ctx.GetLabel(e.name)
	if val == "" {
		return nil, false
	}

	list, ok := e.values[val]
	if !ok || len(list) == 0 {
		return nil, false
	}

	var matched []IHandler
	seen := make(map[string]bool, len(list))
	for _, h := range list {
		if h != nil && !seen[h.Name()] {
			seen[h.Name()] = true
			matched = append(matched, h)
		}
	}

	if len(matched) == 0 {
		return nil, false
	}

	return matched, true
}

// Set 注册策略及其绑定的规则（以策略 id 为主键）
func (e *Extractor) Set(id string, bind []string, handlers []IHandler) {
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
func (e *Extractor) Del(id string) {
	if id == "" {
		return
	}
	e.locker.Lock()
	defer e.locker.Unlock()
	e.delInternal(id)
}

func (e *Extractor) delInternal(id string) {
	oldHandlers, exists := e.idHandlers[id]
	if !exists {
		return
	}

	targetNames := make(map[string]bool, len(oldHandlers))
	for _, h := range oldHandlers {
		if h != nil {
			targetNames[h.Name()] = true
		}
	}

	if e.name == "" {
		filtered := make([]IHandler, 0, len(e.globals))
		for _, h := range e.globals {
			if h != nil && !targetNames[h.Name()] {
				filtered = append(filtered, h)
			}
		}
		e.globals = filtered
	} else {
		if keys, ok := e.idKeys[id]; ok {
			for _, key := range keys {
				if list, existsList := e.values[key]; existsList {
					filtered := make([]IHandler, 0, len(list))
					for _, h := range list {
						if h != nil && !targetNames[h.Name()] {
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
