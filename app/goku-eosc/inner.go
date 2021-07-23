package main

import (
	store_memory_yaml "github.com/eolinker/eosc/modules/store-memory-yaml"
	"github.com/eolinker/eosc/modules/store-yaml"
	"github.com/eolinker/goku-eosc/discovery/consul"
	"github.com/eolinker/goku-eosc/discovery/eureka"
	"github.com/eolinker/goku-eosc/discovery/nacos"
	"github.com/eolinker/goku-eosc/discovery/static"
	http_router "github.com/eolinker/goku-eosc/router/http-router"
	service_http "github.com/eolinker/goku-eosc/service/service-http"
	upstream_http "github.com/eolinker/goku-eosc/upstream/upstream-http"
	upstream_http_anonymous "github.com/eolinker/goku-eosc/upstream/upstream-http_anonymous"
)

func Register()  {
	store.Register()
	store_memory_yaml.Register()

	consul.Register()
	eureka.Register()
	nacos.Register()
	static.Register()

	upstream_http.Register()
	upstream_http_anonymous.Register()

	service_http.Register()
	http_router.Register()
}