package healthcheckhttp

import "github.com/eolinker/goku-eosc/discovery"

//agent 从属于HTTPCheck,实现了IHealthChecker接口
type agent struct {
	agentID string
	*HTTPCheck
}

//NewAgent 创建agent
func NewAgent(agentID string) discovery.IHealthChecker {
	return &agent{agentID: agentID}
}

//AddToCheck 将节点添加进HTTPCheck的检查列表
func (a *agent) AddToCheck(node discovery.INode) error {
	a.addToCheck(&checkNode{
		node:    node,
		agentID: a.agentID,
	})
	return nil
}

//Stop 停止agent并且将HTTPCheck中属于该agent的正在检查的所有节点都移除
func (a *agent) Stop() error {
	a.stop(a.agentID)
	return nil
}
