package health_check_http

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/goku-eosc/discovery"
	"github.com/go-basic/uuid"
)

func NewHttpCheck(config Config) *HttpCheck {
	ctx, cancel := context.WithCancel(context.Background())

	checker := &HttpCheck{
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

type HttpCheck struct {
	config *Config
	ctx    context.Context
	cancel context.CancelFunc
	ch     chan *checkNode
	delCh  chan string
	client *http.Client
	locker sync.RWMutex
}

func (h *HttpCheck) doCheckLoop() {
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
		case node, ok := <-h.ch:
			{
				if ok {
					if _, ok := nodes[node.agentId]; !ok {
						nodes[node.agentId] = make(map[string]*checkNode)
					}
					nodes[node.agentId][node.node.ID()] = node
				}
			}
		case id, ok := <-h.delCh:
			{
				if ok {
					delete(nodes, id)
				}
			}
		}
	}
}

func (h *HttpCheck) Agent() (discovery.IHealthChecker, error) {
	return NewAgent(uuid.New()), nil
}

func (h *HttpCheck) Reset(conf Config) error {
	h.config = &conf
	return nil
}

func (h *HttpCheck) AddToCheck(node discovery.INode) error {
	h.addToCheck(&checkNode{
		node:    node,
		agentId: "",
	})
	return nil
}

func (h *HttpCheck) addToCheck(node *checkNode) error {
	h.ch <- node
	return nil
}

func (h *HttpCheck) Stop() error {
	h.cancel()
	return nil
}

func (h *HttpCheck) stop(id string) {
	h.delCh <- id
}

func (h *HttpCheck) check(nodes map[string]map[string]*checkNode) map[string]map[string]*checkNode {
	newNodes := make(map[string][]*checkNode)
	for _, ns := range nodes {
		for _, n := range ns {
			if n.node.Status() == discovery.Down {
				newNodes[n.node.Addr()] = append(newNodes[n.node.Addr()], n)
			}
		}
	}
	for addr, ns := range newNodes {
		uri := fmt.Sprintf("%s://%s/%s", h.config.Protocol, strings.TrimSuffix(addr, "/"), strings.TrimPrefix(h.config.Url, "/"))
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
			delete(nodes[n.agentId], n.node.ID())
		}
	}
	return nodes
}

type checkNode struct {
	node    discovery.INode
	agentId string
}
