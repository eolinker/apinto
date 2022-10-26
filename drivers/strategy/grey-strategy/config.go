package grey_strategy

import (
	"fmt"
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
	Stop        bool                  `json:"stop"`
	Priority    int                   `json:"priority" label:"优先级" description:"1-999"`
	Filters     strategy.FilterConfig `json:"filters" label:"过滤规则"`
	Rule        Rule                  `json:"grey" label:"灰度规则"`
}

type Rule struct {
	KeepSession  bool        `json:"keep_session"`
	Nodes        []string    `json:"nodes"`        //
	Distribution string      `json:"distribution"` // percent   match
	Percent      int         `json:"percent"`      // 灰度的百分比 四位数
	Matching     []*Matching `json:"matching"`
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

		nodes = append(nodes, newGreyNode(fmt.Sprintf("%s:%d", ip, port), ip, port))
	}

	return nodes
}

type GreyNode struct {
	labels eocontext.Attrs
	id     string
	ip     string
	port   int
	status eocontext.NodeStatus
}

func newGreyNode(id string, ip string, port int) *GreyNode {
	return &GreyNode{labels: map[string]string{}, id: id, ip: ip, port: port, status: eocontext.Running}
}

func (g *GreyNode) GetAttrs() eocontext.Attrs {
	return g.labels
}

func (g *GreyNode) GetAttrByName(name string) (string, bool) {
	v, ok := g.labels[name]
	return v, ok
}

func (g *GreyNode) ID() string {
	return g.id
}

func (g *GreyNode) IP() string {
	return g.ip
}

func (g *GreyNode) Port() int {
	return g.port

}

func (g *GreyNode) Addr() string {
	if g.port == 0 {
		return g.ip
	}
	return fmt.Sprintf("%s:%d", g.ip, g.port)
}

func (g *GreyNode) Status() eocontext.NodeStatus {
	return g.status
}

func (g *GreyNode) Up() {
	g.status = eocontext.Running
}

func (g *GreyNode) Down() {
	g.status = eocontext.Down
}

func (g *GreyNode) Leave() {
	g.status = eocontext.Leave
}
