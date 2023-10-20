package main

import (
	"github.com/eolinker/apinto/drivers/plugins/app"
	"github.com/eolinker/apinto/drivers/plugins/cors"
	data_transform "github.com/eolinker/apinto/drivers/plugins/data-transform"
	dubbo2_proxy_rewrite "github.com/eolinker/apinto/drivers/plugins/dubbo2-proxy-rewrite"
	extra_params "github.com/eolinker/apinto/drivers/plugins/extra-params"
	grpc_proxy_rewrite "github.com/eolinker/apinto/drivers/plugins/grpc-proxy-rewrite"
	"github.com/eolinker/apinto/drivers/plugins/gzip"
	params_check "github.com/eolinker/apinto/drivers/plugins/params-check"
	"github.com/eolinker/apinto/drivers/plugins/prometheus"
	request_interception "github.com/eolinker/apinto/drivers/plugins/request-interception"
	response_rewrite_v2 "github.com/eolinker/apinto/drivers/plugins/response-rewrite_v2"

	access_log "github.com/eolinker/apinto/drivers/plugins/access-log"
	body_check "github.com/eolinker/apinto/drivers/plugins/body-check"
	circuit_breaker "github.com/eolinker/apinto/drivers/plugins/circuit-breaker"
	"github.com/eolinker/apinto/drivers/plugins/counter"
	dubbo2_to_http "github.com/eolinker/apinto/drivers/plugins/dubbo2-to-http"
	extra_params_v2 "github.com/eolinker/apinto/drivers/plugins/extra-params_v2"
	grpc_to_http "github.com/eolinker/apinto/drivers/plugins/gRPC-to-http"
	http_to_dubbo2 "github.com/eolinker/apinto/drivers/plugins/http-to-dubbo2"
	http_to_grpc "github.com/eolinker/apinto/drivers/plugins/http-to-gRPC"
	"github.com/eolinker/apinto/drivers/plugins/http_mocking"
	ip_restriction "github.com/eolinker/apinto/drivers/plugins/ip-restriction"
	"github.com/eolinker/apinto/drivers/plugins/monitor"
	params_transformer "github.com/eolinker/apinto/drivers/plugins/params-transformer"
	proxy_mirror "github.com/eolinker/apinto/drivers/plugins/proxy-mirror"
	proxy_rewrite "github.com/eolinker/apinto/drivers/plugins/proxy-rewrite"
	"github.com/eolinker/apinto/drivers/plugins/proxy_rewrite_v2"
	rate_limiting "github.com/eolinker/apinto/drivers/plugins/rate-limiting"
	response_rewrite "github.com/eolinker/apinto/drivers/plugins/response-rewrite"
	"github.com/eolinker/apinto/drivers/plugins/strategy/cache"
	"github.com/eolinker/apinto/drivers/plugins/strategy/fuse"
	"github.com/eolinker/apinto/drivers/plugins/strategy/grey"
	"github.com/eolinker/apinto/drivers/plugins/strategy/limiting"
	"github.com/eolinker/apinto/drivers/plugins/strategy/visit"

	"github.com/eolinker/eosc"
)

func pluginRegister(extenderRegister eosc.IExtenderDriverRegister) {

	// 服务治理-策略相关插件
	limiting.Register(extenderRegister)
	cache.Register(extenderRegister)
	grey.Register(extenderRegister)
	visit.Register(extenderRegister)
	fuse.Register(extenderRegister)

	// Dubbo协议相关插件
	dubbo2_proxy_rewrite.Register(extenderRegister)
	http_to_dubbo2.Register(extenderRegister)
	dubbo2_to_http.Register(extenderRegister)

	// gRPC协议相关插件
	http_to_grpc.Register(extenderRegister)
	grpc_to_http.Register(extenderRegister)
	grpc_proxy_rewrite.Register(extenderRegister)

	// 请求处理相关插件
	body_check.Register(extenderRegister)
	extra_params.Register(extenderRegister)
	extra_params_v2.Register(extenderRegister)
	params_transformer.Register(extenderRegister)
	proxy_rewrite.Register(extenderRegister)
	proxy_rewrite_v2.Register(extenderRegister)
	http_mocking.Register(extenderRegister)
	params_check.Register(extenderRegister)
	data_transform.Register(extenderRegister)
	request_interception.Register(extenderRegister)

	// 响应处理插件
	response_rewrite.Register(extenderRegister)
	response_rewrite_v2.Register(extenderRegister)
	gzip.Register(extenderRegister)

	// 安全相关插件
	ip_restriction.Register(extenderRegister)
	rate_limiting.Register(extenderRegister)
	cors.Register(extenderRegister)
	circuit_breaker.Register(extenderRegister)
	app.Register(extenderRegister)

	// 可观测性（输出内容到第三方）
	access_log.Register(extenderRegister)
	prometheus.Register(extenderRegister)
	monitor.Register(extenderRegister)
	proxy_mirror.Register(extenderRegister)

	// 计数插件
	counter.Register(extenderRegister)
}
