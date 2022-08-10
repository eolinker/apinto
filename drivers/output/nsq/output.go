package nsq

import (
	"github.com/eolinker/apinto/output"
	"github.com/eolinker/eosc"
	"reflect"
)

type NsqOutput struct {
	id     string
	write  *Writer
	config *Config
}

func (n *NsqOutput) Output(entry eosc.IEntry) error {
	w := n.write
	if w != nil {
		return w.output(entry)
	}
	return eosc.ErrorWorkerNotRunning
}

func (n *NsqOutput) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	cfg, err := Check(conf)
	if err != nil {
		return err
	}
	if reflect.DeepEqual(cfg, n.config) {
		return nil
	}
	n.config = cfg
	w := n.write
	if w != nil {
		return w.reset(cfg)
	}
	return nil
}

func (n *NsqOutput) Stop() error {
	w := n.write
	if w != nil {
		return w.stop()
	}
	return nil
}

func (n *NsqOutput) Id() string {
	return n.id
}

func (n *NsqOutput) Start() error {
	w := n.write
	if w != nil {
		return nil
	}
	w = NewWriter(n.config)
	n.write = w
	return nil
}

func (n *NsqOutput) CheckSkill(skill string) bool {
	return output.CheckSkill(skill)
}
