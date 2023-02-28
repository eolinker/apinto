package mocking

import "github.com/eolinker/eosc"

type Config struct {
	ResponseStatus  int            `json:"response_status" label:"返回响应的 HTTP 状态码（仅http路由有效）"`
	ContentType     string         `json:"content_type" label:"返回响应的 Header Content-Type" enum:"application/json,application/grpc"`
	ProtobufID      eosc.RequireId `json:"protobuf_id"  label:"Protobuf ID" switch:"content_type==='application/grpc'" skill:"github.com/eolinker/apinto/grpc-transcode.transcode.IDescriptor"`
	ResponseExample string         `json:"response_example"  switch:"content_type==='application/json'" label:"返回响应的Body，与jsonschema字段二选一"`
	ResponseSchema  string         `json:"response_schema"  switch:"content_type==='application/json'" label:"指定响应的jsonschema对象"`
}

const (
	contentTypeJson = "application/json"
	contentTypeGrpc = "application/grpc"
)
