package response_rewrite_v2

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

type Config struct {
	Matches []*Match `json:"matches" label:"响应匹配规则"`
}

type Match struct {
	StatusCode      int                `json:"match_status_code" label:"匹配状态码" minimum:"100" default:"200" description:"最小值：100"`
	HeaderMatch     []*HeaderMatchRule `json:"match_headers" label:"响应头匹配规则"`
	BodyMatch       *MatchRule         `json:"match_body" label:"响应体匹配规则"`
	ResponseRewrite *ResponseRewrite   `json:"response_rewrite" label:"重写响应内容"`
}

type HeaderMatchRule struct {
	MatchRule
	HeaderKey string `json:"header_key" label:"响应头Key"`
}

type MatchRule struct {
	Content   string `json:"content" label:"匹配内容"`
	MatchType string `json:"match_type" label:"匹配类型" enum:"equal,contain,prefix,suffix,regex"`
}

// ResponseRewrite 重写内容
type ResponseRewrite struct {
	Body       string            `json:"body" label:"响应体"`
	StatusCode int               `json:"status_code" label:"响应状态码" default:"200"`
	Headers    map[string]string `json:"headers"`
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	handlers := make([]*responseRewrite, 0, len(conf.Matches))
	for _, match := range conf.Matches {
		rh := newRewriteHandler(match.ResponseRewrite)
		needParseVariable := rh.HasVariable()
		handlers = append(handlers, &responseRewrite{
			matcher:        newMatcher(match.StatusCode, match.HeaderMatch, match.BodyMatch, needParseVariable),
			rewriteHandler: rh,
		})
	}
	return &executor{
		WorkerBase: drivers.Worker(id, name),
		handlers:   handlers,
	}, nil
}
