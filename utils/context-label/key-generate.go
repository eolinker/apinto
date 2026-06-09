package context_label

import (
	"fmt"
	"strings"

	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

// parseVariables 提前解析出模板中所有的变量名占位符列表
func parseVariables(template string) map[string]struct{} {
	vars := make(map[string]struct{})
	runes := []rune(template)
	n := len(runes)
	for i := 0; i < n; {
		if runes[i] == '{' {
			j := i + 1
			for j < n && runes[j] != '}' {
				j++
			}
			if j < n && runes[j] == '}' {
				varName := string(runes[i+1 : j])
				vars[varName] = struct{}{}
				i = j + 1
			} else {
				i++
			}
		} else {
			i++
		}
	}
	return vars
}

type GetLabelFunc func(ctx http_context.IHttpContext) string

type IKeyGenerator interface {
	Key(ctx http_context.IHttpContext, fns ...GetLabelFunc) string
}

func NewKeyGenerator(org string) IKeyGenerator {
	return &keyGenerator{
		org:  org,
		vars: parseVariables(org),
	}
}

type keyGenerator struct {
	org  string
	vars map[string]struct{}
}

func (k *keyGenerator) Key(ctx http_context.IHttpContext, fns ...GetLabelFunc) string {
	target := k.org
	for key := range k.vars {
		value := ctx.GetLabel(key)
		if value == "" {
			for _, fn := range fns {
				value = fn(ctx)
				if value != "" {
					break
				}
			}
		}
		if value == "" {
			continue
		}
		target = strings.ReplaceAll(target, fmt.Sprintf("{%s}", key), value)
	}

	return target
}
