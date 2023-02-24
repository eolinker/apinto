package main

import (
	"github.com/eolinker/apinto/application/auth"
	"github.com/eolinker/apinto/drivers/app"
	"github.com/eolinker/apinto/drivers/certs"
	"github.com/eolinker/apinto/drivers/discovery/consul"
	"github.com/eolinker/apinto/drivers/discovery/eureka"
	"github.com/eolinker/apinto/drivers/discovery/nacos"
	"github.com/eolinker/apinto/drivers/discovery/static"
	"github.com/eolinker/apinto/drivers/output/fileoutput"
	"github.com/eolinker/apinto/drivers/output/httpoutput"
	"github.com/eolinker/apinto/drivers/output/kafka"
	"github.com/eolinker/apinto/drivers/output/nsq"
	prometheus_output "github.com/eolinker/apinto/drivers/output/prometheus"
	"github.com/eolinker/apinto/drivers/output/syslog"
	plugin_manager "github.com/eolinker/apinto/drivers/plugin-manager"
	access_log "github.com/eolinker/apinto/drivers/plugins/access-log"
	plugin_app "github.com/eolinker/apinto/drivers/plugins/app"
	circuit_breaker "github.com/eolinker/apinto/drivers/plugins/circuit-breaker"
	"github.com/eolinker/apinto/drivers/plugins/cors"
	dubbo2_proxy_rewrite "github.com/eolinker/apinto/drivers/plugins/dubbo2-proxy-rewrite"
	dubbo2_to_http "github.com/eolinker/apinto/drivers/plugins/dubbo2-to-http"
	extra_params "github.com/eolinker/apinto/drivers/plugins/extra-params"
	grpc_to_http "github.com/eolinker/apinto/drivers/plugins/gRPC-to-http"
	grpc_proxy_rewrite "github.com/eolinker/apinto/drivers/plugins/grpc-proxy-rewrite"
	"github.com/eolinker/apinto/drivers/plugins/gzip"
	http_to_dubbo2 "github.com/eolinker/apinto/drivers/plugins/http-to-dubbo2"
	http_to_grpc "github.com/eolinker/apinto/drivers/plugins/http-to-gRPC"
	ip_restriction "github.com/eolinker/apinto/drivers/plugins/ip-restriction"
	"github.com/eolinker/apinto/drivers/plugins/monitor"
	params_transformer "github.com/eolinker/apinto/drivers/plugins/params-transformer"
	prometheus_plugin "github.com/eolinker/apinto/drivers/plugins/prometheus"
	proxy_rewrite "github.com/eolinker/apinto/drivers/plugins/proxy-rewrite"
	proxy_rewriteV2 "github.com/eolinker/apinto/drivers/plugins/proxy_rewrite_v2"
	rate_limiting "github.com/eolinker/apinto/drivers/plugins/rate-limiting"
	response_rewrite "github.com/eolinker/apinto/drivers/plugins/response-rewrite"
	"github.com/eolinker/apinto/drivers/plugins/strategy/cache"
	"github.com/eolinker/apinto/drivers/plugins/strategy/fuse"
	"github.com/eolinker/apinto/drivers/plugins/strategy/grey"
	"github.com/eolinker/apinto/drivers/plugins/strategy/limiting"
	"github.com/eolinker/apinto/drivers/plugins/strategy/visit"
	"github.com/eolinker/apinto/drivers/resources/datasource/influxdbv2"
	"github.com/eolinker/apinto/drivers/resources/redis"
	dubbo2_router "github.com/eolinker/apinto/drivers/router/dubbo2-router"
	grpc_router "github.com/eolinker/apinto/drivers/router/grpc-router"
	http_router "github.com/eolinker/apinto/drivers/router/http-router"
	service "github.com/eolinker/apinto/drivers/service"
	cache_strategy "github.com/eolinker/apinto/drivers/strategy/cache-strategy"
	fuse_strategy "github.com/eolinker/apinto/drivers/strategy/fuse-strategy"
	grey_strategy "github.com/eolinker/apinto/drivers/strategy/grey-strategy"
	limiting_strategy "github.com/eolinker/apinto/drivers/strategy/limiting-strategy"
	visit_strategy "github.com/eolinker/apinto/drivers/strategy/visit-strategy"
	"github.com/eolinker/apinto/drivers/template"

	"github.com/eolinker/apinto/drivers/transcode/protobuf"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/extends"
	process_worker "github.com/eolinker/eosc/process-worker"
)

func ProcessWorker() {
	registerInnerExtenders()
	process_worker.Process()
}
func registerInnerExtenders() {
	extends.AddInnerExtendProject("eolinker.com", "apinto", Register)
}
func Register(extenderRegister eosc.IExtenderDriverRegister) {
	// router
	http_router.Register(extenderRegister)
	grpc_router.Register(extenderRegister)
	dubbo2_router.Register(extenderRegister)

	template.Register(extenderRegister)

	// service
	service.Register(extenderRegister)

	////// upstream
	//upstream_http.Register(extenderRegister)

	// discovery
	static.Register(extenderRegister)
	nacos.Register(extenderRegister)
	consul.Register(extenderRegister)
	eureka.Register(extenderRegister)

	//output
	fileoutput.Register(extenderRegister)
	nsq.Register(extenderRegister)
	httpoutput.Register(extenderRegister)
	kafka.Register(extenderRegister)
	syslog.Register(extenderRegister)
	prometheus_output.Register(extenderRegister)

	//app
	app.Register(extenderRegister)
	auth.Register(extenderRegister)

	redis.Register(extenderRegister)
	influxdbv2.Register(extenderRegister)

	//plugin
	plugin_manager.Register(extenderRegister)

	certs.Register(extenderRegister)

	plugin_app.Register(extenderRegister)
	extra_params.Register(extenderRegister)
	params_transformer.Register(extenderRegister)
	proxy_rewrite.Register(extenderRegister)
	ip_restriction.Register(extenderRegister)
	rate_limiting.Register(extenderRegister)
	cors.Register(extenderRegister)
	gzip.Register(extenderRegister)
	response_rewrite.Register(extenderRegister)
	circuit_breaker.Register(extenderRegister)

	access_log.Register(extenderRegister)
	prometheus_plugin.Register(extenderRegister)
	monitor.Register(extenderRegister)
	proxy_rewriteV2.Register(extenderRegister)

	limiting.Register(extenderRegister)
	limiting_strategy.Register(extenderRegister)

	cache.Register(extenderRegister)
	cache_strategy.Register(extenderRegister)

	grey.Register(extenderRegister)
	grey_strategy.Register(extenderRegister)

	visit.Register(extenderRegister)
	visit_strategy.Register(extenderRegister)

	fuse.Register(extenderRegister)
	fuse_strategy.Register(extenderRegister)

	grpc_proxy_rewrite.Register(extenderRegister)

	dubbo2_proxy_rewrite.Register(extenderRegister)
	http_to_dubbo2.Register(extenderRegister)
	dubbo2_to_http.Register(extenderRegister)

	http_to_grpc.Register(extenderRegister)
	protocbuf.Register(extenderRegister)
	grpc_to_http.Register(extenderRegister)
}
