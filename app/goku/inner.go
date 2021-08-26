package main

import (
	"github.com/eolinker/goku/drivers/auth/aksk"
	"github.com/eolinker/goku/drivers/auth/apikey"
	"github.com/eolinker/goku/drivers/auth/basic"
	"github.com/eolinker/goku/drivers/auth/jwt"
	"github.com/eolinker/goku/drivers/discovery/consul"
	"github.com/eolinker/goku/drivers/discovery/eureka"
	"github.com/eolinker/goku/drivers/discovery/nacos"
	"github.com/eolinker/goku/drivers/discovery/static"
	http_router "github.com/eolinker/goku/drivers/router/http-router"
	service_http "github.com/eolinker/goku/drivers/service/service-http"
	upstream_http "github.com/eolinker/goku/drivers/upstream/upstream-http"
	store_memory "github.com/eolinker/goku/store-memory"
)

func Register() {
	storeRegister()

	routerRegister()

	serviceRegister()

	upstreamRegister()

	discoveryRegister()

	authRegister()
}

func authRegister() {
	basic.Register()
	apikey.Register()
	aksk.Register()
	jwt.Register()
}

func discoveryRegister() {
	consul.Register()
	eureka.Register()
	nacos.Register()
	static.Register()
}

func storeRegister() {
	store_memory.Register()
}

func upstreamRegister() {
	upstream_http.Register()
}

func serviceRegister() {
	service_http.Register()
}

func routerRegister() {
	http_router.Register()
}
