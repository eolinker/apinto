package certs

import (
	"crypto/tls"
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

var (
	_ eosc.IWorker        = (*Worker)(nil)
	_ eosc.IWorkerDestroy = (*Worker)(nil)
)

type Worker struct {
	drivers.WorkerBase
	config    *Config
	isRunning bool
	cert      *tls.Certificate
}

func (w *Worker) Destroy() error {
	controller.Del(w.Id(), w.cert)
	return nil
}

func (w *Worker) Start() error {
	w.isRunning = true

	cert, err := parseCert(w.config.Key, w.config.Pem)
	if err != nil {
		return err
	}

	w.cert = cert

	controller.Save(w.cert.Leaf.Subject.CommonName, w.cert)

	return nil
}

func (w *Worker) Reset(conf interface{}, _ map[eosc.RequireId]eosc.IWorker) error {

	config := conf.(*Config)

	cert, err := parseCert(config.Key, config.Pem)
	if err != nil {
		return err
	}

	if w.isRunning {
		controller.Save(cert.Leaf.Subject.CommonName, cert)
	}

	return nil
}

func (w *Worker) Stop() error {
	w.isRunning = false
	return nil
}

func (w *Worker) CheckSkill(string) bool {
	return false
}
