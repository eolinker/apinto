package influxdbv2

import (
	"context"
	"fmt"

	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/drivers"
)

func Create(id, name string, v *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	client := NewClient(v)
	_, err := client.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("connect influxdb error: %w", err)
	}
	client.Close()
	ctx, cancel := context.WithCancel(context.Background())
	return &output{
		WorkerBase: drivers.Worker(id, name),
		cfg:        v,
		ctx:        ctx,
		cancel:     cancel,
	}, nil
}
