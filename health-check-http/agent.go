package health_check_http

import "github.com/eolinker/goku/discovery"

//agent 从属于HTTPCheck,实现了IHealthChecker接口
type agent struct {
	agentID string
	checker *HTTPCheck
}

//NewAgent 创建agent
func NewAgent(agentID string, checker *HTTPCheck) discovery.IHealthChecker {
	return &agent{agentID: agentID, checker: checker}
}

//AddToCheck 将节点添加进HTTPCheck的检查列表
func (a *agent) AddToCheck(node discovery.INode) error {
	a.checker.addToCheck(&checkNode{
		node:    node,
		agentID: a.agentID,
	})
	return nil
}

//Stop 停止agent并且将HTTPCheck中属于该agent的正在检查的所有节点都移除
func (a *agent) Stop() error {
	a.checker.stop(a.agentID)
	return nil
}
