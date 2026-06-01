package billing

import (
	"regexp"
	"strings"

	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

// MatchRules 请求匹配规则，支持 HTTP method / path 前缀 / path 正则 / header 多维度组合。
// 全部已配置维度均命中才视为命中。
type MatchRules struct {
	Method      string            `json:"method" label:"HTTP方法" description:"匹配的 HTTP 方法（不区分大小写）"`
	PathPrefix  string            `json:"path_prefix" label:"路径前缀" description:"path 前缀匹配"`
	PathPattern string            `json:"path_pattern" label:"路径正则" description:"path 正则匹配"`
	Headers     map[string]string `json:"headers" label:"请求头匹配" description:"header 键值精确匹配"`
}

type matchRulesMatcher struct {
	method      string
	pathPrefix  string
	pathPattern *regexp.Regexp
	headers     map[string]string
}

func newMatchRulesMatcher(rules *MatchRules) (*matchRulesMatcher, error) {
	m := &matchRulesMatcher{
		method:     strings.ToUpper(rules.Method),
		pathPrefix: rules.PathPrefix,
		headers:    rules.Headers,
	}
	if rules.PathPattern != "" {
		pattern, err := regexp.Compile(rules.PathPattern)
		if err != nil {
			return nil, err
		}
		m.pathPattern = pattern
	}
	return m, nil
}

// Match 仅在所有配置维度均命中时返回 true。未配置的维度视为通配。
func (m *matchRulesMatcher) Match(ctx http_context.IHttpContext) bool {
	if m.method != "" {
		if strings.ToUpper(ctx.Proxy().Method()) != m.method {
			return false
		}
	}

	path := ctx.Proxy().URI().Path()
	if m.pathPrefix != "" {
		if !strings.HasPrefix(path, m.pathPrefix) {
			return false
		}
	}
	if m.pathPattern != nil {
		if !m.pathPattern.MatchString(path) {
			return false
		}
	}

	for key, value := range m.headers {
		headerValue := ctx.Proxy().Header().GetHeader(key)
		if headerValue != value {
			return false
		}
	}

	return true
}
