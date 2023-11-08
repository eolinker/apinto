package counter

import (
	"fmt"
	"strings"

	"github.com/eolinker/eosc"

	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var _ IKeyGenerator = (*keyGenerate)(nil)

type IKeyGenerator interface {
	Key(ctx http_service.IHttpContext) string
	Variables(ctx http_service.IHttpContext) (eosc.Untyped[string, string], bool)
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

func (k *keyGenerate) Variables(ctx http_service.IHttpContext) (eosc.Untyped[string, string], bool) {
	variables := eosc.BuildUntyped[string, string]()
	entry := ctx.GetEntry()
	for _, v := range k.variables {
		value := eosc.ReadStringFromEntry(entry, v)
		if value == "" {
			return nil, false
		}
		variables.Set(fmt.Sprintf("$%s", v), value)
	}
	return variables, true
}

func (k *keyGenerate) Key(ctx http_service.IHttpContext) string {
	variables := make([]interface{}, 0, len(k.variables))
	for _, v := range k.variables {
		variables = append(variables, ctx.GetLabel(v))
	}
	return fmt.Sprintf(k.format, variables...)
}
