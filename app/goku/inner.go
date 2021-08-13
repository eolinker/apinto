package main

import (
	"github.com/eolinker/goku/auth/aksk"
	"github.com/eolinker/goku/auth/apikey"
	"github.com/eolinker/goku/auth/basic"
	"github.com/eolinker/goku/auth/jwt"
	"github.com/eolinker/goku/discovery/consul"
	"github.com/eolinker/goku/discovery/eureka"
	"github.com/eolinker/goku/discovery/nacos"
	"github.com/eolinker/goku/discovery/static"
	http_router "github.com/eolinker/goku/router/http-router"
	service_http "github.com/eolinker/goku/service/service-http"
	store_memory "github.com/eolinker/goku/store-memory"
	upstream_http "github.com/eolinker/goku/upstream/upstream-http"
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
