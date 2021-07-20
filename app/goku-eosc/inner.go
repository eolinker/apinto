package main

import (
	store_memory_yaml "github.com/eolinker/eosc/modules/store-memory-yaml"
	"github.com/eolinker/eosc/modules/store-yaml"
	http_router "github.com/eolinker/goku-eosc/http-router"
)

func Register()  {
	store.Register()
	store_memory_yaml.Register()
	http_router.Register()
}