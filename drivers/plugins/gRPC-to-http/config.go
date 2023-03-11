package grpc_to_http

import "github.com/eolinker/eosc"

type Config struct {
	Path       string            `json:"path" label:"请求路径"`
	Method     string            `json:"method" label:"请求方式" enum:"POST,PUT,PATCH"`
	ProtobufID eosc.RequireId    `json:"protobuf_id" required:"true" label:"Protobuf ID" skill:"github.com/eolinker/apinto/grpc-transcode.transcode.IDescriptor"`
	Headers    map[string]string `json:"headers" label:"额外头部"`
	Query      map[string]string `json:"query" label:"query参数"`
}
