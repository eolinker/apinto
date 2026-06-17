package influxdb_v2

import (
	"context"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

const name = "influxdb-v2"

// Register 注册 influxdbv2 驱动工厂
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(name, NewFactory())
}

// NewFactory 创建 influxdbv2 驱动工厂
func NewFactory() eosc.IExtenderDriverFactory {
	return drivers.NewFactory[Config](Create)
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	cfg, err := check(conf)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	w := &Output{
		WorkerBase: drivers.Worker(id, name),
		outputChan: make(chan *Point, 100),
		ctx:        ctx,
		cancel:     cancel,
	}
	err = w.reset(cfg)
	if err != nil {
		cancel()
		return nil, err
	}
	go w.doLoop()
	return w, nil
}
