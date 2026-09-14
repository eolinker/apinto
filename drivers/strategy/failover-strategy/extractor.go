package failover_strategy

import (
	"strings"
	"sync"

	"github.com/eolinker/eosc/eocontext"
)

type IExtractor interface {
	Name() string
	Get(ctx eocontext.EoContext) ([]IHandler, bool)
	Set(id string, bind []string, handlers []IHandler)
	Del(id string)
}

// Extractor 基于特定维度 label（如 resource、provider 或 "" 通用兜底）管理灾备策略 Handler
type Extractor struct {
	locker        sync.RWMutex
	name          string                // 维度 label 名，如 "resource"、"provider"，"" 表示全局通用维度
	exact         map[string][]IHandler // 精确值索引: exactValue -> handlers (按优先级降序)
	wildcards     []IHandler            // 通配符/兜底 handlers (如 "*"、"all" 或无过滤条件)
	ids           map[string][]string   // id -> 该策略绑定的 bind 规则列表
	idExactKeys   map[string][]string   // id -> 该策略添加到 exact 表中的 key 列表（用于精确删除）
	idHasWildcard map[string]bool       // id -> 是否登记到了 wildcards 表
	idHandlers    map[string][]IHandler // id -> 该策略的 handlers
}

func NewExtractor(name string) *Extractor {
	return &Extractor{
		name:          name,
		exact:         make(map[string][]IHandler),
		wildcards:     make([]IHandler, 0),
		ids:           make(map[string][]string),
		idExactKeys:   make(map[string][]string),
		idHasWildcard: make(map[string]bool),
		idHandlers:    make(map[string][]IHandler),
	}
}

func (e *Extractor) Name() string {
	return e.name
}

// isExactValue 判断 filter 规则是否为可建 O(1) 索引的精确字符串值（排除通配符、正则、区间表达式等）
func isExactValue(pattern string) bool {
	if pattern == "" || pattern == "*" || pattern == "all" {
		return false
	}
	if strings.ContainsAny(pattern, "*?[]{}()^$|\\") {
		return false
	}
	if strings.HasPrefix(pattern, "~") || strings.HasPrefix(pattern, "=") ||
		strings.HasPrefix(pattern, ">") || strings.HasPrefix(pattern, "<") ||
		strings.HasPrefix(pattern, "!") {
		return false
	}
	return true
}

// Get 从当前 Extractor 检索匹配的 Handler 列表
func (e *Extractor) Get(ctx eocontext.EoContext) ([]IHandler, bool) {
	e.locker.RLock()
	defer e.locker.RUnlock()

	var matched []IHandler
	seen := make(map[string]bool)

	// 1. 如果有指定的维度名（如 resource, provider），先尝试精确索引
	if e.name != "" && ctx != nil {
		val := ctx.GetLabel(e.name)
		if val != "" {
			if list, ok := e.exact[val]; ok {
				for _, h := range list {
					if h != nil && !seen[h.Name()] {
						seen[h.Name()] = true
						matched = append(matched, h)
					}
				}
			}
		}
	}

	// 2. 追加通配符兜底 handlers
	for _, h := range e.wildcards {
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

	e.ids[id] = bind
	e.idHandlers[id] = handlers

	var exactKeys []string
	hasWildcard := false

	if len(bind) == 0 {
		hasWildcard = true
	} else {
		for _, b := range bind {
			if !isExactValue(b) {
				hasWildcard = true
			} else {
				e.exact[b] = append(e.exact[b], handlers...)
				exactKeys = append(exactKeys, b)
			}
		}
	}

	if hasWildcard {
		e.wildcards = append(e.wildcards, handlers...)
		e.idHasWildcard[id] = true
	}

	if len(exactKeys) > 0 {
		e.idExactKeys[id] = exactKeys
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

	// 1. 从 exact 索引表中移除
	if exactKeys, ok := e.idExactKeys[id]; ok {
		for _, key := range exactKeys {
			if list, existsList := e.exact[key]; existsList {
				filtered := make([]IHandler, 0, len(list))
				for _, h := range list {
					if h != nil && !targetNames[h.Name()] {
						filtered = append(filtered, h)
					}
				}
				if len(filtered) == 0 {
					delete(e.exact, key)
				} else {
					e.exact[key] = filtered
				}
			}
		}
	}

	// 2. 从 wildcards 中移除
	if e.idHasWildcard[id] {
		filtered := make([]IHandler, 0, len(e.wildcards))
		for _, h := range e.wildcards {
			if h != nil && !targetNames[h.Name()] {
				filtered = append(filtered, h)
			}
		}
		e.wildcards = filtered
		delete(e.idHasWildcard, id)
	}

	delete(e.ids, id)
	delete(e.idExactKeys, id)
	delete(e.idHandlers, id)
}
