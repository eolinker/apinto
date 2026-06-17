package variable_relation

import (
	"fmt"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

// DoFilter 满足 eocontext.IFilter 接口
func (w *VariableRelation) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_context.DoHttpFilter(w, ctx, next)
}

// DoHttpFilter 满足 http_context.HttpFilter 接口
func (w *VariableRelation) DoHttpFilter(ctx http_context.IHttpContext, next eocontext.IChain) (err error) {
	if len(w.rules) == 0 {
		return next.DoChain(ctx)
	}

	// 1. 遍历每个规则处理器
	for _, rule := range w.rules {
		// 获取并渲染出用于查找自定义变量的 Key
		key := rule.keyGenerator.Key(ctx)
		if key == "" {
			continue
		}

		// 从自定义变量拉取当前的映射关系
		all, has := customerVar.GetAll(key)
		if !has || len(all) == 0 {
			return fmt.Errorf("variable relation: no mapping found for key '%s'", key)
		}
		matched := false
		// 2. 遍历该 Redis Key 下所有的 Pattern-Value 对进行匹配
		for pattern, val := range all {
			if MatchRule(ctx, pattern, rule.innerKeyTpl) {
				matched = true
				if rule.valueLabel != "" {
					ctx.SetLabel(rule.valueLabel, val)
				}
				// 该规则已匹配成功，不需要在该 Key 下继续查找其它 pattern，直接开始处理下一个 Rule
				break
			}
		}
		if !matched {
			return fmt.Errorf("variable relation: no pattern matched for key '%s'", key)
		}
	}

	return next.DoChain(ctx)
}
