package main

import (
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/log"
	admin_html "github.com/eolinker/eosc/modules/admin-html"
	admin_open_api "github.com/eolinker/eosc/modules/admin-open-api"
	store_memory_yaml "github.com/eolinker/eosc/modules/store-memory-yaml"
	"github.com/eolinker/eosc/modules/store-yaml"
	http_router "github.com/eolinker/goku-eosc/http-router"
	"net/http"
)

func main() {
	store.Register()
	store_memory_yaml.Register()
	http_router.Register()
 	storeName := "memory-yaml"
	file := "profession.yml"

	storeDriver, has := eosc.GetStoreDriver(storeName)
	if !has {
		log.Panic("unkonw store driver:", storeName)
	}

	storeT, err := storeDriver.Create(map[string]string{
		"file": file,
	})
	if err != nil {
		log.Panic(err)
	}

	err = storeT.Initialization()
	if err != nil {
		log.Panic(err)
	}
	pcs := []eosc.ProfessionConfig{
		{
			Name:         "router",
			Label:        "路由",
			Desc:         "路由",
			Dependencies: []string{"service"},
			AppendLabel:  []string{"host", "service"},

			Drivers: []eosc.DriverConfig{
				{
					ID:     "eolinker:goku:example",
					Name:   "http",
					Label:  "http",
					Desc:   "http路由",
					Params: nil,
				},
				{
					ID:     "eolinker:goku:example",
					Name:   "https",
					Label:  "https",
					Desc:   "http路由",
					Params: nil,
				},
			},
		}, {
			Name:         "service",
			Label:        "服务",
			Desc:         "服务",
			Dependencies: []string{"upstream"},
			AppendLabel:  []string{"upstream"},
			Drivers: []eosc.DriverConfig{
				{
					ID:     "eolinker:goku:service_http",
					Name:   "http",
					Label:  "service",
					Desc:   "服务",
					Params: nil,
				},
			},
		},
		{
			Name:         "upstream",
			Label:        "上游/负载",
			Desc:         "上游/负载",
			Dependencies: []string{"discovery"},
			AppendLabel:  []string{"discovery"},
			Drivers: []eosc.DriverConfig{
				{
					ID:     "eolinker:goku:upstream_http_proxy",
					Name:   "http_proxy",
					Label:  "http转发负载",
					Desc:   "http转发负载",
					Params: nil,
				},
			},
		}, {
			Name:         "discovery",
			Label:        "注册中心",
			Desc:         "注册中心",
			Dependencies: []string{},
			AppendLabel:  []string{},
			Drivers: []eosc.DriverConfig{
				{
					ID:     "eolinker:goku:discovery_static",
					Name:   "static",
					Label:  "静态服务发现",
					Desc:   "静态服务发现",
					Params: nil,
				},
				{
					ID:     "eolinker:goku:discovery_nacos",
					Name:   "nacos",
					Label:  "nacos服务发现",
					Desc:   "nacos服务发现",
					Params: nil,
				},
				{
					ID:     "eolinker:goku:discovery_consul",
					Name:   "consul",
					Label:  "consul服务发现",
					Desc:   "consul服务发现",
					Params: nil,
				},
				{
					ID:     "eolinker:goku:discovery_eureka",
					Name:   "eureka",
					Label:  "eureka服务发现",
					Desc:   "eureka服务发现",
					Params: nil,
				},
			},
		},
	}

	professions, err := eosc.ProfessionConfigs(pcs).Gen(eosc.DefaultProfessionDriverRegister, storeT)
	if err != nil {
		panic(err)
	}
	_, err = eosc.NewWorkers(professions, storeT)
	if err != nil {
		log.Panic(err)
	}

	admin := admin_open_api.NewOpenAdmin("/api", professions)
	htmlAdmin := admin_html.NewHtmlAdmin("/", professions)
	handler, err := admin.GenHandler()
	if err != nil {
		panic(err)
	}
	hadlerHtml, err := htmlAdmin.GenHandler()
	if err != nil {
		panic(err)
	}

	httpServer := http.NewServeMux()
	httpServer.Handle("/api/", handler)
	httpServer.Handle("/", hadlerHtml)
	log.Fatal(http.ListenAndServe(":8088", httpServer))
}
