package health_check_http

import "github.com/eolinker/goku-eosc/discovery"

type Agent struct {
	agentId string
	*HttpCheck
}

func NewAgent(agentId string) *Agent {
	return &Agent{agentId: agentId}
}

func (a *Agent) AddToCheck(node discovery.INode) error {
	a.addToCheck(&checkNode{
		node:    node,
		agentId: a.agentId,
	})
	return nil
}

func (a *Agent) Stop() error {
	a.stop(a.agentId)
	return nil
}
