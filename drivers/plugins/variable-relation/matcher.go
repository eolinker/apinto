package variable_relation

import (
	"fmt"
	http_entry "github.com/eolinker/apinto/entries/http-entry"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
	"strings"
)

type IMatcher interface {
	Match(ctx http_context.IHttpContext, pattern string) bool
}

func NewPath() *Path {
	return &Path{}
}

type Path struct {
}

func (p *Path) Match(ctx http_context.IHttpContext, pattern string) bool {
	pattern = normalizePath(pattern)
	path := normalizePath(ctx.Request().URI().Path())

	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")

	if len(patternParts) != len(pathParts) {
		return false
	}

	for i, pPart := range patternParts {
		if isParam(pPart) {
			// 提取参数
			paramName := extractParamName(pPart)
			ctx.SetLabel(fmt.Sprintf("path_param_%s", paramName), pathParts[i])
		} else if pPart != pathParts[i] {
			// 静态部分不匹配
			return false
		}
	}

	return true
}

// normalizePath 规范化路径：去除首尾斜杠，并处理多余斜杠
func normalizePath(p string) string {
	p = strings.Trim(p, "/")             // 去除首尾 /
	p = strings.ReplaceAll(p, "//", "/") // 防止出现双斜杠
	return p
}

// isParam 判断是否为参数占位符
func isParam(part string) bool {
	return len(part) > 2 && (part[0] == '{' && part[len(part)-1] == '}' || part[0] == ':')
}

// extractParamName 提取参数名称
func extractParamName(part string) string {
	if part[0] == '{' {
		return part[1 : len(part)-1]
	}
	return part[1:] // :task_id 风格
}

func NewMatcher(label string) *Matcher {
	return &Matcher{label: label}
}

type Matcher struct {
	label string
}

func (m *Matcher) Match(ctx http_context.IHttpContext, pattern string) bool {
	return ctx.GetLabel(m.label) == pattern
}

// Segment 模板中的片段
type Segment struct {
	IsVar bool
	Name  string
	Value string
}

// ParseTemplate 解析模板，提取变量和静态分隔符
func ParseTemplate(tpl string) []Segment {
	var segments []Segment
	runes := []rune(tpl)
	n := len(runes)
	var lastStatic strings.Builder
	for i := 0; i < n; {
		if runes[i] == '{' {
			j := i + 1
			for j < n && runes[j] != '}' {
				j++
			}
			if j < n && runes[j] == '}' {
				if lastStatic.Len() > 0 {
					segments = append(segments, Segment{IsVar: false, Value: lastStatic.String()})
					lastStatic.Reset()
				}
				segments = append(segments, Segment{IsVar: true, Name: string(runes[i+1 : j])})
				i = j + 1
				continue
			}
		}
		lastStatic.WriteRune(runes[i])
		i++
	}
	if lastStatic.Len() > 0 {
		segments = append(segments, Segment{IsVar: false, Value: lastStatic.String()})
	}
	return segments
}

// ExtractValues 根据模板片段从具体的 pattern 中提取出变量值
func ExtractValues(pattern string, segments []Segment) (map[string]string, bool) {
	values := make(map[string]string)
	curr := pattern

	for i, seg := range segments {
		if !seg.IsVar {
			// 静态分隔符必须精确匹配
			if !strings.HasPrefix(curr, seg.Value) {
				return nil, false
			}
			curr = curr[len(seg.Value):]
		} else {
			// 变量部分，需要确定消费多少字符
			if i == len(segments)-1 {
				// 最后一个片段，消费剩余所有字符
				values[seg.Name] = curr
				curr = ""
			} else {
				// 下一个片段必须是静态分隔符
				nextSeg := segments[i+1]
				if nextSeg.IsVar {
					return nil, false // 不支持连续两个变量，模板必须由分隔符隔开
				}
				idx := strings.Index(curr, nextSeg.Value)
				if idx == -1 {
					return nil, false
				}
				values[seg.Name] = curr[:idx]
				curr = curr[idx:]
			}
		}
	}
	if curr != "" {
		return nil, false
	}
	return values, true
}

// MatchRule 判断请求上下文是否匹配特定规则的变量解析值
func MatchRule(ctx http_context.IHttpContext, pattern string, segments []Segment) bool {
	entry := http_entry.NewEntry(ctx)
	values, ok := ExtractValues(pattern, segments)
	if !ok {
		return false
	}

	for name, patternVal := range values {
		switch name {
		case "method":
			actualMethod := ctx.Request().Method()
			if patternVal != "*" && !strings.EqualFold(actualMethod, patternVal) {
				return false
			}
		case "path":
			pathMatcher := NewPath()
			if !pathMatcher.Match(ctx, patternVal) {
				return false
			}
		default:
			actualVal := entry.ReadLabel(name)
			if patternVal != "*" && actualVal != patternVal {
				return false
			}
		}
	}

	return true
}
