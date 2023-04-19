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
)

var (
	_ discovery.IHealthChecker = (*HTTPCheck)(nil)
)

// NewHTTPCheck 创建HTTPCheck
func NewHTTPCheck(config Config) *HTTPCheck {
	ctx, cancel := context.WithCancel(context.Background())

	checker := &HTTPCheck{
		config: &config,
		ctx:    ctx,
		cancel: cancel,

		client: &http.Client{},
		locker: sync.RWMutex{},
	}

	return checker
}

// HTTPCheck HTTP健康检查结构,实现了IHealthChecker接口
type HTTPCheck struct {
	config *Config
	nodes  discovery.INodes
	ctx    context.Context
	cancel context.CancelFunc

	client *http.Client
	locker sync.RWMutex
}

func (h *HTTPCheck) Check(nodes discovery.INodes) {
	go h.doCheckLoop(nodes)
}

// doCheckLoop 定时检查，维护了一个待检测节点集合
func (h *HTTPCheck) doCheckLoop(nodes discovery.INodes) {
	if h.config.Period < 1 {
		return
	}
	ticker := time.NewTicker(h.config.Period)

	defer ticker.Stop()
	for {
		select {
		case <-h.ctx.Done():
			return
		case <-ticker.C:
			{
				h.check(nodes.All())
			}
		}
	}
}

// Reset 重置HTTPCheck的配置
func (h *HTTPCheck) Reset(conf interface{}) error {
	cf, ok := conf.(Config)
	if !ok {
		return nil
	}
	h.reset(&cf)
	return nil
}
func (h *HTTPCheck) reset(conf *Config) {
	h.config = conf
}

// Stop 停止HTTPCheck，中止定时检查
func (h *HTTPCheck) Stop() {
	h.cancel()

}

// check 对待检查的节点集合进行检测，入参：nodes map[agentID][nodeID]*checkNode
func (h *HTTPCheck) check(nodes []discovery.INode) {

	/*对每个节点地址进行检测
	成功则将属于该地址的所有节点的状态都置于可运行，并从HTTPCheck维护的待检测节点列表中移除
	失败则下次定时检查再进行检测
	*/
	for _, ns := range nodes {
		if ns.Status() != discovery.Down {
			continue
		}
		uri := fmt.Sprintf("%s://%s/%s", h.config.Protocol, strings.TrimSuffix(ns.Addr(), "/"), strings.TrimPrefix(h.config.URL, "/"))
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
		ns.Up()
	}
}
