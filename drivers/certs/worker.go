package certs

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/eolinker/apinto/certs"
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/utils"
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
	certs.DelCert(w.Id())
	return nil
}

func (w *Worker) Start() error {

	return nil
}

func (w *Worker) Reset(conf interface{}, _ map[eosc.RequireId]eosc.IWorker) error {

	config := conf.(*Config)

	cert, err := parseCert(config.Key, config.Pem)
	if err != nil {
		return err
	}

	w.config = config
	certs.SaveCert(w.Id(), cert)

	return nil
}

func (w *Worker) Stop() error {
	return nil
}

func (w *Worker) CheckSkill(string) bool {
	return false
}

func parseCert(privateKey, pemValue string) (*tls.Certificate, error) {

	keydata, err := utils.B64Decode(privateKey)
	if err != nil {
		return nil, err
	}
	pem, err := utils.B64Decode(pemValue)
	if err != nil {
		return nil, err
	}
	certificate, err := tls.X509KeyPair(pem, keydata)
	if err != nil {
		return nil, err
	}

	x509Cert, err := x509.ParseCertificate(certificate.Certificate[0])
	if err != nil {
		return nil, err
	}
	certificate.Leaf = x509Cert
	return &certificate, nil
}
