package response_filter

import (
	"fmt"
	"github.com/eolinker/apinto/utils"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"strings"
)

type CompiledRule struct {
	IsArray   bool
	BasePath  string // 如 d.data
	SubPath   string // 如 id / a.b
	FieldName string // 最终字段名
}

func SafeCompile(paths []string) ([]CompiledRule, error) {
	valid := make([]string, 0)

	for _, p := range paths {
		if err := utils.ValidateJSONPath(p); err != nil {
			return nil, fmt.Errorf("invalid jsonpath %s: %w", p, err)
		}
		valid = append(valid, p)
	}

	return CompileRules(valid)
}

func lastKey(path string) string {
	if !strings.Contains(path, ".") {
		return path
	}
	parts := strings.Split(path, ".")
	return parts[len(parts)-1]
}

func CompileRules(paths []string) ([]CompiledRule, error) {
	rules := make([]CompiledRule, 0)

	for _, p := range paths {
		p = strings.TrimSpace(p)
		if !strings.HasPrefix(p, "$.") {
			return nil, fmt.Errorf("invalid path: %s", p)
		}

		p = strings.TrimPrefix(p, "$.")

		if strings.Contains(p, "[*]") {
			parts := strings.SplitN(p, "[*].", 2)
			rules = append(rules, CompiledRule{
				IsArray:   true,
				BasePath:  parts[0],
				SubPath:   parts[1],
				FieldName: lastKey(parts[1]),
			})
		} else {
			rules = append(rules, CompiledRule{
				IsArray:  false,
				BasePath: p,
			})
		}
	}

	return rules, nil
}

func applyNormal(dst *string, src string, rule CompiledRule) error {
	val := gjson.Get(src, rule.BasePath)
	if val.Exists() {
		var err error
		*dst, err = sjson.Set(*dst, rule.BasePath, val.Value())
		return err
	}
	return nil
}

func applyArray(dst *string, src string, base string, rules []CompiledRule) error {
	arr := gjson.Get(src, base)
	if !arr.IsArray() {
		return nil
	}

	result := make([]map[string]interface{}, 0)

	arr.ForEach(func(_, item gjson.Result) bool {
		obj := map[string]interface{}{}
		for _, r := range rules {
			v := item.Get(r.SubPath)
			if v.Exists() {
				obj[r.FieldName] = v.Value()
			}
		}
		if len(obj) > 0 {
			result = append(result, obj)
		}
		return true
	})

	var err error
	*dst, err = sjson.Set(*dst, base, result)
	return err
}

func Extract(src string, rules []CompiledRule) (string, error) {
	dst := "{}"

	// 1️⃣ 普通字段
	for _, r := range rules {
		if !r.IsArray {
			if err := applyNormal(&dst, src, r); err != nil {
				return "", err
			}
		}
	}

	// 2️⃣ 数组字段（按 basePath 分组）
	group := map[string][]CompiledRule{}
	for _, r := range rules {
		if r.IsArray {
			group[r.BasePath] = append(group[r.BasePath], r)
		}
	}

	for base, rs := range group {
		if err := applyArray(&dst, src, base, rs); err != nil {
			return "", err
		}
	}

	return dst, nil
}
