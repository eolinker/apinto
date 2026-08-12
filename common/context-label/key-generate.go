package context_label

import (
	"github.com/eolinker/eosc/eocontext"
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

type GetLabelFunc func(ctx eocontext.EoContext, label string) string

type IKeyGenerator interface {
	Key(ctx eocontext.EoContext, fns ...GetLabelFunc) string
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

type placeholder struct {
	start int
	end   int
	name  string
}

func findInnermostPlaceholders(runes []rune) []placeholder {
	n := len(runes)
	var list []placeholder
	lastOpen := -1
	for i := 0; i < n; i++ {
		if runes[i] == '{' {
			lastOpen = i
		} else if runes[i] == '}' {
			if lastOpen != -1 {
				list = append(list, placeholder{
					start: lastOpen,
					end:   i,
					name:  string(runes[lastOpen+1 : i]),
				})
				lastOpen = -1
			}
		}
	}
	return list
}

func (k *keyGenerator) Key(ctx eocontext.EoContext, fns ...GetLabelFunc) string {
	runes := []rune(k.org)
	for {
		placeholders := findInnermostPlaceholders(runes)
		if len(placeholders) == 0 {
			break
		}
		replaced := false
		// Process from right to left to keep left indices valid
		for i := len(placeholders) - 1; i >= 0; i-- {
			p := placeholders[i]
			value := ""
			for _, fn := range fns {
				value = fn(ctx, p.name)
				if value != "" {
					break
				}
			}
			
			if value == "" {
				value = ctx.GetLabel(p.name)
			}
			if value != "" {
				valueRunes := []rune(value)
				newRunes := make([]rune, 0, len(runes)-(p.end-p.start+1)+len(valueRunes))
				newRunes = append(newRunes, runes[:p.start]...)
				newRunes = append(newRunes, valueRunes...)
				newRunes = append(newRunes, runes[p.end+1:]...)
				runes = newRunes
				replaced = true
			}
		}
		if !replaced {
			break
		}
	}
	return string(runes)
}
