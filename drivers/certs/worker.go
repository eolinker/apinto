package certs

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
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

	controller.Save(w.cert)

	return nil
}

func (w *Worker) Reset(conf interface{}, _ map[eosc.RequireId]eosc.IWorker) error {

	config := conf.(*Config)

	cert, err := parseCert(config.Key, config.Pem)
	if err != nil {
		return err
	}

	w.config = config

	if w.isRunning {
		controller.Save(cert)
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

func parseCert(privateKey, pemValue string) (*tls.Certificate, error) {
	var err error
	privateKey, err = utils.B64Decode(privateKey)
	if err != nil {
		return nil, err
	}
	pemValue, err = utils.B64Decode(pemValue)
	if err != nil {
		return nil, err
	}

	var cert tls.Certificate
	//获取下一个pem格式证书数据 -----BEGIN CERTIFICATE-----   -----END CERTIFICATE-----
	certDERBlock, restPEMBlock := pem.Decode([]byte(pemValue))
	if certDERBlock == nil {
		return nil, errors.New("证书解析失败")
	}
	//附加数字证书到返回
	cert.Certificate = append(cert.Certificate, certDERBlock.Bytes)
	//继续解析Certificate Chan,这里要明白证书链的概念
	certDERBlockChain, _ := pem.Decode(restPEMBlock)
	if certDERBlockChain != nil {
		//追加证书链证书到返回
		cert.Certificate = append(cert.Certificate, certDERBlockChain.Bytes)
	}

	//解码pem格式的私钥------BEGIN RSA PRIVATE KEY-----   -----END RSA PRIVATE KEY-----
	keyDERBlock, _ := pem.Decode([]byte(privateKey))
	if keyDERBlock == nil {
		return nil, errors.New("证书解析失败")
	}
	var key interface{}
	var errParsePK error
	if keyDERBlock.Type == "RSA PRIVATE KEY" {
		//RSA PKCS1
		key, errParsePK = x509.ParsePKCS1PrivateKey(keyDERBlock.Bytes)
	} else if keyDERBlock.Type == "PRIVATE KEY" {
		//pkcs8格式的私钥解析
		key, errParsePK = x509.ParsePKCS8PrivateKey(keyDERBlock.Bytes)
	}

	if errParsePK != nil {
		return nil, errors.New("证书解析失败")
	} else {
		cert.PrivateKey = key
	}
	//第一个叶子证书就是我们https中使用的证书
	x509Cert, err := x509.ParseCertificate(certDERBlock.Bytes)
	if err != nil {
		return nil, err
	}
	cert.Leaf = x509Cert
	return &cert, nil
}
