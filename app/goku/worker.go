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
	"github.com/eolinker/goku/drivers/log/filelog"
	"github.com/eolinker/goku/drivers/log/httplog"
	"github.com/eolinker/goku/drivers/log/stdlog"
	"github.com/eolinker/goku/drivers/log/syslog"
	http_router "github.com/eolinker/goku/drivers/router/http-router"
	service_http "github.com/eolinker/goku/drivers/service/service-http"
	upstream_http "github.com/eolinker/goku/drivers/upstream/upstream-http"
)

func registerInnerExtenders() {
	extends.AddInnerExtendProject("eolinker.com", "goku", Register)
}
func Register(extenderRegister eosc.IExtenderRegister) {
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

	// auth
	basic.Register(extenderRegister)
	apikey.Register(extenderRegister)
	aksk.Register(extenderRegister)
	jwt.Register(extenderRegister)

	// log
	filelog.Register(extenderRegister)
	httplog.Register(extenderRegister)
	syslog.Register(extenderRegister)
	stdlog.Register(extenderRegister)
}
