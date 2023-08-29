package response_rewrite_v2

import (
	"fmt"

	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

func newRewriteHandler(org *ResponseRewrite) *rewriteHandler {
	return &rewriteHandler{
		statusCode: org.StatusCode,
		body:       newBodyRewrite(org.Body),
		header:     org.Headers,
	}
}

type rewriteHandler struct {
	statusCode int
	body       *bodyRewrite
	header     map[string]string
}

func (r *rewriteHandler) Rewrite(ctx http_service.IHttpContext, variables map[string]string) {
	if r.statusCode > 99 {
		ctx.Response().SetStatus(r.statusCode, fmt.Sprintf("%d", r.statusCode))
	}

	if r.body != nil {
		body := r.body.replace(variables)
		ctx.Response().SetBody([]byte(body))
	}
	if r.header != nil {
		for key, value := range r.header {
			if value == "" {
				ctx.Response().Headers().Del(key)
				continue
			}
			ctx.Response().SetHeader(key, value)
		}
	}
}

func (r *rewriteHandler) HasVariable() bool {
	if r.body != nil {
		return len(r.body.variables) > 0
	}
	return false
}

func newBodyRewrite(content string) *bodyRewrite {
	// 构造带有变量的内容匹配器
	matches := variableRegexp.FindAllStringSubmatch(content, -1)
	variables := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			variables = append(variables, match[1])
		}
	}
	return &bodyRewrite{
		content:   variableRegexp.ReplaceAllString(content, "%s"),
		variables: variables,
	}
}

type bodyRewrite struct {
	content   string
	variables []string
}

func (b *bodyRewrite) replace(variables map[string]string) string {
	values := make([]interface{}, 0, len(b.variables))
	for _, variable := range b.variables {
		values = append(values, variables[variable])
	}
	return fmt.Sprintf(b.content, values...)
}
