package main

import (
	process_worker "github.com/eolinker/eosc/process-worker"
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
	http_router "github.com/eolinker/goku/drivers/router/http-router"
	service_http "github.com/eolinker/goku/drivers/service/service-http"
	upstream_http "github.com/eolinker/goku/drivers/upstream/upstream-http"
)

func ProcessWorker() {
	register()
	process_worker.Process()
}

func register() {
	// router
	http_router.Register()

	// service
	service_http.Register()

	// upstream
	upstream_http.Register()

	// discovery
	static.Register()
	nacos.Register()
	consul.Register()
	eureka.Register()

	// auth
	basic.Register()
	apikey.Register()
	aksk.Register()
	jwt.Register()

	// log
	filelog.Register()
	httplog.Register()
	//syslog.Register()
	stdlog.Register()
}
