package main

import (
	access_relational "github.com/eolinker/apinto/drivers/plugins/access-relational"
	"github.com/eolinker/apinto/drivers/plugins/acl"
	"github.com/eolinker/apinto/drivers/plugins/aes"
	ai_formatter "github.com/eolinker/apinto/drivers/plugins/ai-formatter"
	ai_prompt "github.com/eolinker/apinto/drivers/plugins/ai-prompt"
	"github.com/eolinker/apinto/drivers/plugins/app"
	auto_redirect "github.com/eolinker/apinto/drivers/plugins/auto-redirect"
	"github.com/eolinker/apinto/drivers/plugins/cors"
	data_transform "github.com/eolinker/apinto/drivers/plugins/data-transform"
	dubbo2_proxy_rewrite "github.com/eolinker/apinto/drivers/plugins/dubbo2-proxy-rewrite"
	extra_params "github.com/eolinker/apinto/drivers/plugins/extra-params"
	grpc_proxy_rewrite "github.com/eolinker/apinto/drivers/plugins/grpc-proxy-rewrite"
	"github.com/eolinker/apinto/drivers/plugins/gzip"
	js_inject "github.com/eolinker/apinto/drivers/plugins/js-inject"
	"github.com/eolinker/apinto/drivers/plugins/oauth2"
	oauth2_introspection "github.com/eolinker/apinto/drivers/plugins/oauth2-introspection"
	params_check "github.com/eolinker/apinto/drivers/plugins/params-check"
	params_check_v2 "github.com/eolinker/apinto/drivers/plugins/params-check-v2"
	"github.com/eolinker/apinto/drivers/plugins/prometheus"
	request_file_parse "github.com/eolinker/apinto/drivers/plugins/request-file-parse"
	request_interception "github.com/eolinker/apinto/drivers/plugins/request-interception"
	response_file_parse "github.com/eolinker/apinto/drivers/plugins/response-file-parse"
	response_filter "github.com/eolinker/apinto/drivers/plugins/response-filter"
	response_rewrite_v2 "github.com/eolinker/apinto/drivers/plugins/response-rewrite_v2"
	rsa_filter "github.com/eolinker/apinto/drivers/plugins/rsa-filter"
	script_handler "github.com/eolinker/apinto/drivers/plugins/script-handler"
	data_mask "github.com/eolinker/apinto/drivers/plugins/strategy/data-mask"

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
	data_mask.Register(extenderRegister)

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
	params_check_v2.Register(extenderRegister)
	data_transform.Register(extenderRegister)
	request_interception.Register(extenderRegister)
	request_file_parse.Register(extenderRegister)

	// 响应处理插件
	response_rewrite.Register(extenderRegister)
	response_rewrite_v2.Register(extenderRegister)
	response_filter.Register(extenderRegister)
	gzip.Register(extenderRegister)
	response_file_parse.Register(extenderRegister)
	auto_redirect.Register(extenderRegister)

	// 安全相关插件
	ip_restriction.Register(extenderRegister)
	rate_limiting.Register(extenderRegister)
	cors.Register(extenderRegister)
	circuit_breaker.Register(extenderRegister)
	app.Register(extenderRegister)
	rsa_filter.Register(extenderRegister)
	aes.Register(extenderRegister)
	js_inject.Register(extenderRegister)
	acl.Register(extenderRegister)
	access_relational.Register(extenderRegister)
	// 可观测性（输出内容到第三方）
	access_log.Register(extenderRegister)
	prometheus.Register(extenderRegister)
	monitor.Register(extenderRegister)
	proxy_mirror.Register(extenderRegister)

	// 计数插件
	counter.Register(extenderRegister)

	// 鉴权插件
	oauth2.Register(extenderRegister)
	oauth2_introspection.Register(extenderRegister)

	// ai相关插件
	ai_prompt.Register(extenderRegister)
	ai_formatter.Register(extenderRegister)
	//ai_balance.Register(extenderRegister)

	script_handler.Register(extenderRegister)
}
