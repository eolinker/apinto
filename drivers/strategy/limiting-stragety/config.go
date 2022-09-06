package limiting_stragety

import "github.com/eolinker/apinto/strategy"

type Threshold struct {
	Second int64 `json:"second" label`
	Minute int64 `json:"minute"`
	Hour   int64 `json:"hour"`
}
type Rule struct {
	Metrics []string  `json:"metrics" `
	Query   Threshold `json:"query" `
	Traffic Threshold `json:"traffic"`
}
type Config struct {
	Filters strategy.FilterConfig `json:"filters"`
	Rule    Rule                  `json:"limiting"`
}
