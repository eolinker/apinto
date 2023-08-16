package response_rewrite_v2

import (
	"github.com/eolinker/eosc"
)

type Config struct {
}

type Match struct {
	StatusCode int    `json:"status_code" label:"匹配状态码" minimum:"100" description:"最小值：100"`
	Rules      string `json:"body" label:"匹配内容" description:"正则表达式"`
}

type BodyMatch struct {
	Content  string `json:"content" label:"匹配内容"`
	Type     string `json:"type" label:"匹配类型" enum:"contain,regex"`
	Position string `json:"position" label:"匹配位置" enum:"header,body"`
}

// Rewrite 重写内容
type Rewrite struct {
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {

	return nil, nil
}
