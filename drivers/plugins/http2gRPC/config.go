package grpc_proxy_rewrite

import "github.com/eolinker/eosc"

type Config struct {
	Service    string            `json:"service" label:"服务名称"`
	Method     string            `json:"method" label:"方法名称"`
	Authority  string            `json:"authority" label:"虚拟主机域名(Authority)"`
	Headers    map[string]string `json:"headers" label:"请求头部"`
	Reflect    bool              `json:"reflect"`
	ProtobufID eosc.RequireId    `json:"protobuf_id" required:"false" label:"服务名 or 配置" switch:"reflect === false"`
}
