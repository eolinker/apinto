package http_to_grpc

import "github.com/eolinker/eosc"

type Config struct {
	Service    string            `json:"service" label:"服务名称"`
	Method     string            `json:"method" label:"方法名称"`
	Authority  string            `json:"authority" label:"虚拟主机域名(Authority)"`
	Format     string            `json:"format" label:"数据格式" enum:"json"`
	Reflect    bool              `json:"reflect" label:"反射"`
	ProtobufID eosc.RequireId    `json:"protobuf_id" required:"false" label:"Protobuf ID" skill:"github.com/eolinker/apinto/grpc-transcode.transcode.IDescriptor" switch:"reflect === false"`
	Headers    map[string]string `json:"headers" label:"额外头部"`
}
