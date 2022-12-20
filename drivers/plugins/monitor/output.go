package monitor

import (
	monitor_entry "github.com/eolinker/apinto/monitor-entry"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/log"
	"reflect"

	scope_manager "github.com/eolinker/apinto/drivers/scope-manager"
)

var outputManager = NewOutputManager()

type OutputManager struct {
	outputs   eosc.Untyped[string, scope_manager.IProxyOutput]
	pointChan chan point
}

func NewOutputManager() *OutputManager {
	o := &OutputManager{
		outputs:   eosc.BuildUntyped[string, scope_manager.IProxyOutput](),
		pointChan: make(chan point, 100),
	}
	go o.doLoop()
	return o
}

type point struct {
	id     string
	points []monitor_entry.IPoint
}

func (o *OutputManager) Set(id string, proxy scope_manager.IProxyOutput) {
	o.outputs.Set(id, proxy)
}

func (o *OutputManager) Output(id string, ps []monitor_entry.IPoint) {
	o.pointChan <- point{
		id:     id,
		points: ps,
	}
}
func (o *OutputManager) doLoop() {
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
			for _, proxy := range v.List() {
				out, ok := proxy.(monitor_entry.IOutput)
				if !ok {
					log.Error("error output type: ", reflect.TypeOf(proxy))
					continue
				}
				out.Output(p.points...)
			}
		}
	}
}
