package grpc_proxy_rewrite

type Config struct {
	Service string            `json:"service" label:"服务名称"`
	Method  string            `json:"method" label:"方法名称"`
	Headers map[string]string `json:"headers" label:"headers"`
}
