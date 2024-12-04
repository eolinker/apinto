package loki

import (
	"context"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

const name = "loki_output"

// Register 注册nsqd驱动工厂
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(name, NewFactory())
}
func NewFactory() eosc.IExtenderDriverFactory {

	return drivers.NewFactory[Config](Create)
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	ctx, cancel := context.WithCancel(context.Background())
	w := &Output{
		WorkerBase: drivers.Worker(id, name),
		outputChan: make(chan *Request, 100),
		ctx:        ctx,
		cancel:     cancel,
	}
	err := w.reset(conf)
	if err != nil {
		return nil, err
	}
	go w.doLoop()
	return w, nil
}
