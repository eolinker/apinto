package aes

import (
	"errors"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

type Config struct {
	Key       string `json:"key"`
	Salt      string `json:"salt"`
	Mode      string `json:"mode" enum:"ECB, CBC, CTR, OFB, CFB"`
	Algorithm string `json:"algorithm" enum:"AES-128, AES-192, AES-256"`
}

func check(conf *Config) error {
	if conf.Key == "" {
		return errors.New("key is empty")
	}
	if conf.Mode == "" {
		return errors.New("mode is empty")
	}
	if conf.Algorithm == "" {
		return errors.New("algorithm is empty")
	}
	if conf.Mode != "ECB" && conf.Mode != "CBC" && conf.Mode != "CTR" && conf.Mode != "OFB" && conf.Mode != "CFB" {
		return errors.New("mode error")
	}

	if conf.Algorithm != "AES-128" && conf.Algorithm != "AES-192" && conf.Algorithm != "AES-256" {
		return errors.New("algorithm error")
	}
	return nil
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	e := &executor{
		WorkerBase: drivers.Worker(id, name),
	}
	e.reset(conf)
	return e, nil
}
