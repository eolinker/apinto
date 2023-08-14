package counter

import (
	"fmt"
	"strings"

	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var _ IKeyGenerator = (*keyGenerate)(nil)

type IKeyGenerator interface {
	Key(ctx http_service.IHttpContext) string
}

func newKeyGenerate(key string) *keyGenerate {
	key = strings.TrimSpace(key)
	tmp := strings.Split(key, ":")

	keys := make([]string, 0, len(tmp))
	variables := make([]string, 0, len(tmp))
	for _, t := range tmp {
		t = strings.TrimSpace(t)
		tLen := len(t)
		if tLen > 0 {
			if tLen > 1 && t[0] == '$' {
				variables = append(variables, t[1:])
				keys = append(keys, "%s")
			} else {
				keys = append(keys, t)
			}
		}
	}
	return &keyGenerate{format: strings.Join(keys, ":"), variables: variables}
}

type keyGenerate struct {
	format string
	// 变量列表
	variables []string
}

func (k *keyGenerate) Key(ctx http_service.IHttpContext) string {
	variables := make([]interface{}, 0, len(k.variables))
	for _, v := range k.variables {
		variables = append(variables, ctx.GetLabel(v))
	}
	return fmt.Sprintf(k.format, variables...)
}
