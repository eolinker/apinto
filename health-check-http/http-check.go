package health_check_http

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/apinto/discovery"
	"github.com/go-basic/uuid"
)

//NewHTTPCheck 创建HTTPCheck
func NewHTTPCheck(config Config) *HTTPCheck {
	ctx, cancel := context.WithCancel(context.Background())

	checker := &HTTPCheck{
		config: &config,
		ctx:    ctx,
		cancel: cancel,
		ch:     make(chan *checkNode, 10),
		client: &http.Client{},
		locker: sync.RWMutex{},
	}
	go checker.doCheckLoop()
	return checker
}

//HTTPCheck HTTP健康检查结构,实现了IHealthChecker接口
type HTTPCheck struct {
	config *Config
	ctx    context.Context
	cancel context.CancelFunc
	ch     chan *checkNode
	delCh  chan string
	client *http.Client
	locker sync.RWMutex
}

//doCheckLoop 定时检查，维护了一个待检测节点集合
func (h *HTTPCheck) doCheckLoop() {
	if h.config.Period < 1 {
		return
	}
	ticker := time.NewTicker(h.config.Period)
	nodes := map[string]map[string]*checkNode{}
	defer ticker.Stop()
	for {
		select {
		case <-h.ctx.Done():
			return
		case <-ticker.C:
			{
				nodes = h.check(nodes)
			}
		//接收待检测节点并存入待检测节点集合
		case node, ok := <-h.ch:
			{
				if ok {
					if _, ok := nodes[node.agentID]; !ok {
						nodes[node.agentID] = make(map[string]*checkNode)
					}
					nodes[node.agentID][node.node.ID()] = node
				}
			}
		//接收agentID,并将待检测集合中属于该agent的所有节点移除
		case agentID, ok := <-h.delCh:
			{
				if ok {
					delete(nodes, agentID)
				}
			}
		}
	}
}

//Agent 生成一个agent
func (h *HTTPCheck) Agent() (discovery.IHealthChecker, error) {
	return NewAgent(uuid.New(), h), nil
}

//Reset 重置HTTPCheck的配置
func (h *HTTPCheck) Reset(conf Config) error {
	h.config = &conf
	return nil
}

//AddToCheck 将节点添加进HTTPCheck的检查列表
func (h *HTTPCheck) AddToCheck(node discovery.INode) error {
	h.addToCheck(&checkNode{
		node:    node,
		agentID: "",
	})
	return nil
}

//addToCheck 将节点传入HTTPCheck的检测Channel
func (h *HTTPCheck) addToCheck(node *checkNode) error {
	h.ch <- node
	return nil
}

//Stop 停止HTTPCheck，中止定时检查
func (h *HTTPCheck) Stop() error {
	h.cancel()
	return nil
}

//stop 停止从属该agentID的所有节点的健康检查
func (h *HTTPCheck) stop(agentID string) {
	h.delCh <- agentID
}

//check 对待检查的节点集合进行检测，入参：nodes map[agentID][nodeID]*checkNode
func (h *HTTPCheck) check(nodes map[string]map[string]*checkNode) map[string]map[string]*checkNode {
	//将待检测节点集合中地址相同的节点整合在一起，结构为：map[node.Addr][]*checkNode
	newNodes := make(map[string][]*checkNode)
	for _, ns := range nodes {
		for _, n := range ns {
			if n.node.Status() == discovery.Down {
				newNodes[n.node.Addr()] = append(newNodes[n.node.Addr()], n)
			}
		}
	}

	/*对每个节点地址进行检测
	成功则将属于该地址的所有节点的状态都置于可运行，并从HTTPCheck维护的待检测节点列表中移除
	失败则下次定时检查再进行检测
	*/
	for addr, ns := range newNodes {
		uri := fmt.Sprintf("%s://%s/%s", h.config.Protocol, strings.TrimSuffix(addr, "/"), strings.TrimPrefix(h.config.URL, "/"))
		h.client.Timeout = h.config.Timeout
		request, err := http.NewRequest(h.config.Method, uri, nil)
		if err != nil {
			log.Error(err)
			continue
		}
		resp, err := h.client.Do(request)
		if err != nil {
			log.Error(err)
			continue
		}
		resp.Body.Close()
		if h.config.SuccessCode != resp.StatusCode {
			log.Error(err)
			continue
		}
		for _, n := range ns {
			n.node.Up()
			delete(nodes[n.agentID], n.node.ID())
		}
	}
	return nodes
}

//checkNode 进入检查channel的节点结构
type checkNode struct {
	node    discovery.INode
	agentID string
}
