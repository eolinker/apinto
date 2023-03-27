package grey_strategy

import (
	"github.com/eolinker/apinto/discovery"
	"github.com/eolinker/apinto/strategy"
	"github.com/eolinker/eosc/eocontext"
	"strconv"
	"strings"
)

const (
	percent = "percent"
	match   = "match"
	grey    = "grey"
	normal  = "normal"
)

type Config struct {
	Name        string                `json:"name" skip:"skip"`
	Description string                `json:"description" skip:"skip"`
	Stop        bool                  `json:"stop" label:"禁用"`
	Priority    int                   `json:"priority" label:"优先级" description:"1-999"`
	Filters     strategy.FilterConfig `json:"filters" label:"过滤规则"`
	Rule        Rule                  `json:"grey" label:"灰度规则"`
}

type Rule struct {
	KeepSession  bool        `json:"keep_session" label:"会话保持规则"`
	Nodes        []string    `json:"nodes" label:"灰度节点"`
	Distribution string      `json:"distribution" label:"流量分配方式" enum:"percent,match"` // percent   match
	Percent      int         `json:"percent" label:"灰度节点流量占比" description:"1-9999"`    // 灰度的百分比 四位数
	Matching     []*Matching `json:"matching" label:"高级匹配"`
}

type Matching struct {
	Type  string `json:"type"  label:"类型" enum:"header,query,cookie"`
	Name  string `json:"name"  label:"参数名"`
	Value string `json:"value"  label:"值规" `
}

func (r *Rule) GetNodes() []eocontext.INode {
	nodes := make([]eocontext.INode, 0)
	for _, node := range r.Nodes {
		addrSlide := strings.Split(node, ":")

		ip := addrSlide[0]
		port := 0
		if len(addrSlide) > 1 {
			port, _ = strconv.Atoi(addrSlide[1])
		}

		container := discovery.NewAppContainer()

		nodes = append(nodes, discovery.NewNode(container.Get(ip, port), nil))
	}

	return nodes
}
