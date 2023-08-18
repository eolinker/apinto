package response_rewrite_v2

import (
	"fmt"
	"regexp"
	"strings"

	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var variableRegexp = regexp.MustCompile(`\#([^\$]+)\$`)

const (
	MatchTypeEqual   = "equal"
	MatchTypeContain = "contain"
	MatchTypePrefix  = "prefix"
	MatchTypeSuffix  = "suffix"
	MatchTypeRegex   = "regex"
)

type matcher struct {
	statusCode int
	body       *contentMatcher
	header     map[string]*contentMatcher
}

func newMatcher(statusCode int, headerMatches []*HeaderMatchRule, bodyMatch *MatchRule, needParseVariable bool) *matcher {
	header := make(map[string]*contentMatcher)
	for _, rule := range headerMatches {
		header[rule.HeaderKey] = newContentMatcher(rule.Content, rule.MatchType, false)
	}

	return &matcher{
		statusCode: statusCode,
		header:     header,
		body:       newContentMatcher(bodyMatch.Content, bodyMatch.MatchType, needParseVariable),
	}
}

func (m *matcher) Match(ctx http_service.IHttpContext) (map[string]string, bool) {
	if m.statusCode != ctx.Response().StatusCode() {
		return nil, false
	}

	for key, mm := range m.header {
		header := ctx.Response().Headers().Get(key)
		if header == "" {
			return nil, false
		}
		_, match := mm.Match(header)
		if !match {
			return nil, false
		}
	}

	if m.body != nil {
		body := ctx.Response().GetBody()
		return m.body.Match(string(body))
	}
	return nil, true
}

type contentMatcher struct {
	content   string
	variables []string
	matchType string
}

func newContentMatcher(content string, matchType string, allowVariable bool) *contentMatcher {
	if !allowVariable {
		return &contentMatcher{
			content:   content,
			variables: nil,
			matchType: matchType,
		}
	}
	// 构造带有变量的内容匹配器

	matches := variableRegexp.FindAllStringSubmatch(content, -1)
	variables := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			variables = append(variables, match[1])
		}
	}

	return &contentMatcher{
		content:   variableRegexp.ReplaceAllString(content, "(.+)"),
		variables: variables,
		matchType: matchType,
	}
}

func (c *contentMatcher) Match(content string) (map[string]string, bool) {
	if c.matchType == MatchTypeRegex {
		// 正则类型，直接匹配，不筛选变量
		re := regexp.MustCompile(c.content)
		return nil, re.MatchString(c.content)
	}
	regexpRule := ""
	switch c.matchType {
	case MatchTypeEqual:
		if len(c.variables) < 1 {
			// 内容中不包含变量
			return nil, c.content == content
		}
		regexpRule = fmt.Sprintf("^%s$", c.content)
	case MatchTypePrefix:
		if len(c.variables) < 1 {
			// 内容中不包含变量
			return nil, strings.HasPrefix(content, c.content)
		}
		regexpRule = fmt.Sprintf("^%s", c.content)
	case MatchTypeSuffix:
		if len(c.variables) < 1 {
			// 内容中不包含变量
			return nil, strings.HasSuffix(content, c.content)
		}
		regexpRule = fmt.Sprintf("%s$", c.content)
	case MatchTypeContain:
		if len(c.variables) < 1 {
			// 内容中不包含变量
			return nil, strings.Contains(content, c.content)
		}
		regexpRule = c.content
	}

	// 记录变量值，若变量重复，以最后一次的变量值为准
	re := regexp.MustCompile(regexpRule)
	matches := re.FindAllStringSubmatch(content, -1)
	if len(matches) < 1 {
		return nil, false
	}
	variables := make(map[string]string, len(c.variables))
	for _, match := range matches {
		if len(match) > 1 {
			for i, variable := range c.variables {
				variables[variable] = match[i+1]
			}
		}
	}
	return variables, true
}

func (c *contentMatcher) HasVariable() bool {
	return len(c.variables) > 0
}
