package certs

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

var (
	_ eosc.IWorker        = (*Worker)(nil)
	_ eosc.IWorkerDestroy = (*Worker)(nil)
)

type Worker struct {
	drivers.WorkerBase
	config *Config
}

func (w *Worker) Destroy() error {
	controller.Del(w.Id())
	return nil
}

func (w *Worker) Start() error {
	return nil
}

func (w *Worker) Reset(conf interface{}, _ map[eosc.RequireId]eosc.IWorker) error {
	config := conf.(*Config)
	w.config = config

	cert, err := parseCert(config.Key, config.Pem)
	if err != nil {
		return err
	}
	controller.Save(cert.Leaf.Subject.CommonName, cert)
	return nil
}

func (w *Worker) Stop() error {
	return nil
}

func (w *Worker) CheckSkill(string) bool {
	return false
}
