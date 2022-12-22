package monitor_manager

import (
	"github.com/eolinker/apinto/monitor-entry"
	"github.com/eolinker/apinto/scope-manager"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/common/bean"
	"github.com/eolinker/eosc/log"
	"github.com/eolinker/eosc/utils"
	"reflect"
	"time"
)

var _ IManager = (*MonitorManager)(nil)

type IManager interface {
	SetProxyOutput(id string, proxy scope_manager.IProxyOutput)
	ConcurrencyAdd(apiID string, count int32)
	RemoveCurrencyAPI(apiID string)
	Output(id string, ps []monitor_entry.IPoint)
}

var monitorManager = NewMonitorManager()

func init() {
	bean.Injection(&monitorManager)
}

type MonitorManager struct {
	outputs        eosc.Untyped[string, scope_manager.IProxyOutput]
	concurrentApis eosc.Untyped[string, *concurrency]
	pointChan      chan point
}

func (o *MonitorManager) RemoveCurrencyAPI(apiID string) {
	v, ok := o.concurrentApis.Del(apiID)
	if ok {
		now := time.Now()
		globalLabel := utils.GlobalLabelGet()
		tags := map[string]string{
			"api":     apiID,
			"cluster": globalLabel["cluster"],
			"node":    globalLabel["node"],
		}
		fields := map[string]interface{}{
			"value": v.Get(),
		}
		p := monitor_entry.NewPoint("node", tags, fields, now)
		for _, v := range o.outputs.List() {
			o.proxyOutput(v, []monitor_entry.IPoint{p})
		}
	}
}

func NewMonitorManager() IManager {
	o := &MonitorManager{
		outputs:        eosc.BuildUntyped[string, scope_manager.IProxyOutput](),
		concurrentApis: eosc.BuildUntyped[string, *concurrency](),
		pointChan:      make(chan point, 100),
	}
	go o.doLoop()
	return o
}

type point struct {
	id     string
	points []monitor_entry.IPoint
}

func (o *MonitorManager) SetProxyOutput(id string, proxy scope_manager.IProxyOutput) {
	o.outputs.Set(id, proxy)
}

func (o *MonitorManager) ConcurrencyAdd(id string, count int32) {
	v, has := o.concurrentApis.Get(id)
	if !has {
		v = &concurrency{count: 0}
		o.concurrentApis.Set(id, v)
	}
	v.Add(count)
}

func (o *MonitorManager) Output(id string, ps []monitor_entry.IPoint) {
	o.pointChan <- point{
		id:     id,
		points: ps,
	}
}

func (o *MonitorManager) doLoop() {
	ticket := time.NewTicker(1 * time.Second)
	defer ticket.Stop()
	for {
		select {
		case p, ok := <-o.pointChan:
			if !ok {
				return
			}
			v, has := o.outputs.Get(p.id)
			if !has {
				continue
			}
			o.proxyOutput(v, p.points)
		case <-ticket.C:
			ticket.Reset(1 * time.Second)
		}
	}
}

func (o *MonitorManager) proxyOutput(v scope_manager.IProxyOutput, ps []monitor_entry.IPoint) {
	for _, proxy := range v.List() {
		out, ok := proxy.(monitor_entry.IOutput)
		if !ok {
			log.Error("error output type: ", reflect.TypeOf(proxy))
			continue
		}
		out.Output(ps...)
	}
}

func (o *MonitorManager) genNodePoints() []monitor_entry.IPoint {
	now := time.Now()
	globalLabel := utils.GlobalLabelGet()
	cluster := globalLabel["cluster"]
	node := globalLabel["node"]
	points := make([]monitor_entry.IPoint, 0, o.concurrentApis.Count())
	for key, value := range o.concurrentApis.All() {
		tags := map[string]string{
			"api":     key,
			"cluster": cluster,
			"node":    node,
		}
		fields := map[string]interface{}{
			"value": value.Get(),
		}
		points = append(points, monitor_entry.NewPoint("node", tags, fields, now))
	}
	return points
}
