package main

import (
	store_memory_yaml "github.com/eolinker/eosc/modules/store-memory-yaml"
	"github.com/eolinker/eosc/modules/store-yaml"
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
	upstream_http "github.com/eolinker/goku-eosc/upstream/upstream-http"
	upstream_http_anonymous "github.com/eolinker/goku-eosc/upstream/upstream-http_anonymous"
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
	store.Register()
	store_memory_yaml.Register()
}

func upstreamRegister() {
	upstream_http.Register()
	upstream_http_anonymous.Register()
}

func serviceRegister() {
	service_http.Register()
}

func routerRegister() {
	http_router.Register()
}
