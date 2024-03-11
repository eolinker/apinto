package nsq

import (
	"reflect"

	scope_manager "github.com/eolinker/apinto/scope-manager"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/output"
	"github.com/eolinker/eosc"
)

var _ output.IEntryOutput = (*NsqOutput)(nil)
var _ eosc.IWorker = (*NsqOutput)(nil)

type NsqOutput struct {
	drivers.WorkerBase
	write     *Writer
	config    *Config
	isRunning bool
}

func (n *NsqOutput) Output(entry eosc.IEntry) error {
	w := n.write
	if w != nil {
		w.output(entry)
		return nil
	}
	return eosc.ErrorWorkerNotRunning
}

func (n *NsqOutput) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cfg, err := check(conf)
	if err != nil {
		return err
	}
	if reflect.DeepEqual(cfg, n.config) {
		return nil
	}
	n.config = cfg
	if n.isRunning {
		w := n.write
		if w == nil {
			w = NewWriter()
		}

		err = w.reset(cfg)
		if err != nil {
			return err
		}
		n.write = w
	}
	scope_manager.Set(n.Id(), n, n.config.Scopes...)
	return nil
}

func (n *NsqOutput) Stop() error {
	scope_manager.Del(n.Id())
	w := n.write
	if w != nil {
		return w.stop()
	}
	return nil
}

func (n *NsqOutput) Start() error {
	n.isRunning = true
	w := n.write
	if w == nil {
		w = NewWriter()
	}
	err := w.reset(n.config)
	if err != nil {
		return err
	}
	n.write = w
	scope_manager.Set(n.Id(), n, n.config.Scopes...)
	return nil
}

func (n *NsqOutput) CheckSkill(skill string) bool {
	return output.CheckSkill(skill)
}
