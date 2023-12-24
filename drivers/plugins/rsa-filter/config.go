package rsa_filter

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

type Config struct {
	PrivateKey         string `json:"private_key" label:"私钥" description:"对请求体进行解密，对响应体进行签名"`
	PublicKey          string `json:"public_key" label:"公钥" description:"对请求体进行验签，对响应体进行加密"`
	RequestSignHeader  string `json:"request_sign_header" label:"请求签名头"`
	ResponseSignHeader string `json:"response_sign_header" label:"响应签名头"`
	Format             string `json:"format" label:"密钥格式" enum:"origin,base64"`
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	e := &executor{
		WorkerBase: drivers.Worker(id, name),
	}
	e.reset(conf)
	return e, nil
}
