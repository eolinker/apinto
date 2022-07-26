package main

import (
	"github.com/eolinker/apinto/drivers/auth/aksk"
	"github.com/eolinker/apinto/drivers/auth/apikey"
	"github.com/eolinker/apinto/drivers/auth/basic"
	"github.com/eolinker/apinto/drivers/auth/jwt"
	"github.com/eolinker/apinto/drivers/discovery/consul"
	"github.com/eolinker/apinto/drivers/discovery/eureka"
	"github.com/eolinker/apinto/drivers/discovery/nacos"
	"github.com/eolinker/apinto/drivers/discovery/static"
	"github.com/eolinker/apinto/drivers/output/fileoutput"
	"github.com/eolinker/apinto/drivers/output/httpoutput"
	"github.com/eolinker/apinto/drivers/output/kafka"
	"github.com/eolinker/apinto/drivers/output/nsq"
	"github.com/eolinker/apinto/drivers/output/syslog"
	access_log "github.com/eolinker/apinto/drivers/plugins/access-log"
	"github.com/eolinker/apinto/drivers/plugins/auth"
	circuit_breaker "github.com/eolinker/apinto/drivers/plugins/circuit-breaker"
	"github.com/eolinker/apinto/drivers/plugins/cors"
	extra_params "github.com/eolinker/apinto/drivers/plugins/extra-params"
	"github.com/eolinker/apinto/drivers/plugins/gzip"
	ip_restriction "github.com/eolinker/apinto/drivers/plugins/ip-restriction"
	params_transformer "github.com/eolinker/apinto/drivers/plugins/params-transformer"
	proxy_rewrite "github.com/eolinker/apinto/drivers/plugins/proxy-rewrite"
	proxy_rewrite2 "github.com/eolinker/apinto/drivers/plugins/proxy-rewrite2"
	rate_limiting "github.com/eolinker/apinto/drivers/plugins/rate-limiting"
	response_rewrite "github.com/eolinker/apinto/drivers/plugins/response-rewrite"
	http_router "github.com/eolinker/apinto/drivers/router/http-router"
	service_http "github.com/eolinker/apinto/drivers/service/service-http"
	//upstream_http "github.com/eolinker/apinto/drivers/upstream/upstream-http"
	plugin_manager "github.com/eolinker/apinto/plugin-manager"
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

	// service
	service_http.Register(extenderRegister)

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

	// auth
	basic.Register(extenderRegister)
	apikey.Register(extenderRegister)
	aksk.Register(extenderRegister)
	jwt.Register(extenderRegister)

	//plugin
	plugin_manager.Register(extenderRegister)

	auth.Register(extenderRegister)

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
	proxy_rewrite2.Register(extenderRegister)
}
