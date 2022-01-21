package main

import (
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/extends"
	"github.com/eolinker/goku/drivers/auth/aksk"
	"github.com/eolinker/goku/drivers/auth/apikey"
	"github.com/eolinker/goku/drivers/auth/basic"
	"github.com/eolinker/goku/drivers/auth/jwt"
	"github.com/eolinker/goku/drivers/discovery/consul"
	"github.com/eolinker/goku/drivers/discovery/eureka"
	"github.com/eolinker/goku/drivers/discovery/nacos"
	"github.com/eolinker/goku/drivers/discovery/static"
	"github.com/eolinker/goku/drivers/output/fileoutput"
	"github.com/eolinker/goku/drivers/output/nsq"
	access_log "github.com/eolinker/goku/drivers/plugins/access-log"
	"github.com/eolinker/goku/drivers/plugins/auth"
	circuit_breaker "github.com/eolinker/goku/drivers/plugins/circuit-breaker"
	"github.com/eolinker/goku/drivers/plugins/cors"
	extra_params "github.com/eolinker/goku/drivers/plugins/extra-params"
	"github.com/eolinker/goku/drivers/plugins/gzip"
	ip_restriction "github.com/eolinker/goku/drivers/plugins/ip-restriction"
	params_transformer "github.com/eolinker/goku/drivers/plugins/params-transformer"
	proxy_rewrite "github.com/eolinker/goku/drivers/plugins/proxy-rewrite"
	rate_limiting "github.com/eolinker/goku/drivers/plugins/rate-limiting"
	response_rewrite "github.com/eolinker/goku/drivers/plugins/response-rewrite"
	"github.com/eolinker/goku/drivers/plugins/rewrite"
	http_router "github.com/eolinker/goku/drivers/router/http-router"
	service_http "github.com/eolinker/goku/drivers/service/service-http"
	upstream_http "github.com/eolinker/goku/drivers/upstream/upstream-http"
	plugin_manager "github.com/eolinker/goku/plugin-manager"
)

func registerInnerExtenders() {
	extends.AddInnerExtendProject("eolinker.com", "goku", Register)
}
func Register(extenderRegister eosc.IExtenderDriverRegister) {
	// router
	http_router.Register(extenderRegister)

	// service
	service_http.Register(extenderRegister)

	// upstream
	upstream_http.Register(extenderRegister)

	// discovery
	static.Register(extenderRegister)
	nacos.Register(extenderRegister)
	consul.Register(extenderRegister)
	eureka.Register(extenderRegister)
	//output
	fileoutput.Register(extenderRegister)
	nsq.Register(extenderRegister)

	// auth
	basic.Register(extenderRegister)
	apikey.Register(extenderRegister)
	aksk.Register(extenderRegister)
	jwt.Register(extenderRegister)

	//plugin
	plugin_manager.Register(extenderRegister)
	auth.Register(extenderRegister)
	rewrite.Register(extenderRegister)

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
}
