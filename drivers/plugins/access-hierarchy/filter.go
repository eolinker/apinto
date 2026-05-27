package access_hierarchy

import (
	http_entry "github.com/eolinker/apinto/entries/http-entry"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

// DoHttpFilter 核心 HTTP 过滤方法。请求经过当前插件时，此方法会被自动调用。
func (w *AccessHierarchy) DoHttpFilter(ctx http_context.IHttpContext, next eocontext.IChain) (err error) {
	// 如果没有任何配置的校验规则，直接跳过并进入链的下一个过滤器
	if len(w.rules) == 0 {
		return next.DoChain(ctx)
	}

	// 1. 构建当前请求的 Entry 指标承载体
	entry := http_entry.NewEntry(ctx)

	// 2. 依次遍历每个校验规则处理器。只要有一个规则匹配且允许访问，就代表当前请求合法，直接放行
	for _, rule := range w.rules {
		if rule.Check(entry) {
			return next.DoChain(ctx) // 允许访问，放行至下一环
		}
	}

	// 3. 所有规则检查全部不通过，进行拦截阻断
	httpContext, err := http_context.Assert(ctx)
	if err != nil {
		return err
	}
	// 返回配置文件中自定义的 Response（或默认的 403 页面）给客户端
	w.response.Response(httpContext)
	return nil
}

// DoFilter 满足底层通用 eocontext.IFilter 接口，将上下文桥接转换为 HTTP Filter。
func (w *AccessHierarchy) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_context.DoHttpFilter(w, ctx, next)
}
