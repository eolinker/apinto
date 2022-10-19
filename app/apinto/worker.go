package main

import (
	"github.com/eolinker/apinto/application/auth"
	"github.com/eolinker/apinto/drivers/app"
	"github.com/eolinker/apinto/drivers/discovery/consul"
	"github.com/eolinker/apinto/drivers/discovery/eureka"
	"github.com/eolinker/apinto/drivers/discovery/nacos"
	"github.com/eolinker/apinto/drivers/discovery/static"
	"github.com/eolinker/apinto/drivers/output/fileoutput"
	"github.com/eolinker/apinto/drivers/output/httpoutput"
	"github.com/eolinker/apinto/drivers/output/kafka"
	"github.com/eolinker/apinto/drivers/output/nsq"
	"github.com/eolinker/apinto/drivers/output/syslog"
	plugin_manager "github.com/eolinker/apinto/drivers/plugin-manager"
	access_log "github.com/eolinker/apinto/drivers/plugins/access-log"
	plugin_app "github.com/eolinker/apinto/drivers/plugins/app"
	circuit_breaker "github.com/eolinker/apinto/drivers/plugins/circuit-breaker"
	"github.com/eolinker/apinto/drivers/plugins/cors"
	extra_params "github.com/eolinker/apinto/drivers/plugins/extra-params"
	"github.com/eolinker/apinto/drivers/plugins/gzip"
	ip_restriction "github.com/eolinker/apinto/drivers/plugins/ip-restriction"
	params_transformer "github.com/eolinker/apinto/drivers/plugins/params-transformer"
	proxy_rewrite "github.com/eolinker/apinto/drivers/plugins/proxy-rewrite"
	proxy_rewriteV2 "github.com/eolinker/apinto/drivers/plugins/proxy_rewrite_v2"
	rate_limiting "github.com/eolinker/apinto/drivers/plugins/rate-limiting"
	response_rewrite "github.com/eolinker/apinto/drivers/plugins/response-rewrite"
	"github.com/eolinker/apinto/drivers/plugins/strategy/cache"
	"github.com/eolinker/apinto/drivers/plugins/strategy/grey"
	"github.com/eolinker/apinto/drivers/plugins/strategy/limiting"
	"github.com/eolinker/apinto/drivers/plugins/strategy/visit"
	"github.com/eolinker/apinto/drivers/resources/redis"
	http_router "github.com/eolinker/apinto/drivers/router/http-router"
	service "github.com/eolinker/apinto/drivers/service"
	cache_strategy "github.com/eolinker/apinto/drivers/strategy/cache-strategy"
	grey_strategy "github.com/eolinker/apinto/drivers/strategy/grey-strategy"
	limiting_strategy "github.com/eolinker/apinto/drivers/strategy/limiting-strategy"
	visit_strategy "github.com/eolinker/apinto/drivers/strategy/visit-strategy"
	template "github.com/eolinker/apinto/drivers/template"

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

	//app
	app.Register(extenderRegister)
	auth.Register(extenderRegister)

	redis.Register(extenderRegister)
	//plugin
	plugin_manager.Register(extenderRegister)

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
	proxy_rewriteV2.Register(extenderRegister)

	limiting.Register(extenderRegister)
	limiting_strategy.Register(extenderRegister)

	cache.Register(extenderRegister)
	cache_strategy.Register(extenderRegister)

	grey.Register(extenderRegister)
	grey_strategy.Register(extenderRegister)

	visit.Register(extenderRegister)
	visit_strategy.Register(extenderRegister)
}
