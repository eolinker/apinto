package main

import (
	"github.com/eolinker/goku-eosc/auth/aksk"
	"github.com/eolinker/goku-eosc/auth/apikey"
	"github.com/eolinker/goku-eosc/auth/basic"
	"github.com/eolinker/goku-eosc/auth/jwt"
	"github.com/eolinker/goku-eosc/discovery/consul"
	"github.com/eolinker/goku-eosc/discovery/eureka"
	"github.com/eolinker/goku-eosc/discovery/nacos"
	"github.com/eolinker/goku-eosc/discovery/static"
	http_router "github.com/eolinker/goku-eosc/router/http-router"
	service_http "github.com/eolinker/goku-eosc/service/service-http"
	store_memory "github.com/eolinker/goku-eosc/store-memory"
	upstream_http "github.com/eolinker/goku-eosc/upstream/upstream-http"
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
