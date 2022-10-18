package redis

import (
	"github.com/eolinker/apinto/resources"
	"github.com/eolinker/eosc"
	"github.com/go-redis/redis/v8"
	"reflect"
)

type Worker struct {
	resources.ICache
	config    *Config
	client    *redis.ClusterClient
	id        string
	name      string
	isRunning bool
}

func (w *Worker) Id() string {
	return w.id
}

func (w *Worker) Start() error {
	if w.isRunning {
		return nil
	}
	if len(w.config.Addrs) == 0 {
		return eosc.ErrorConfigIsNil
	}
	client, err := w.config.connect()
	if err != nil {
		return err
	}
	w.client, w.ICache = client, &Cmdable{cmdable: client}
	w.isRunning = true
	return nil
}

func (w *Worker) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cfg, err := checkConfig(conf)
	if err != nil {
		return err
	}
	if w.config == nil || !reflect.DeepEqual(w.config, cfg) {

		w.config = cfg
		client, err := cfg.connect()
		if err != nil {
			return err
		}
		if w.isRunning {
			oc := w.client
			w.client, w.ICache = client, &Cmdable{cmdable: client}
			oc.Close()
		} else {
			client.Close()
		}

	}
	return nil

}

func (w *Worker) Stop() error {
	if !w.isRunning {
		return eosc.ErrorWorkerNotRunning
	}
	w.isRunning = false
	if w.client != nil {
		w.ICache = &Empty{}
		e := w.client.Close()
		w.client = nil

		return e
	}
	return nil
}

func (w *Worker) CheckSkill(skill string) bool {
	return skill == resources.CacheSkill
}
