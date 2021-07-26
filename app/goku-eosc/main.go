package main

import (
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/log"
	admin_html "github.com/eolinker/eosc/modules/admin-html"
	admin_open_api "github.com/eolinker/eosc/modules/admin-open-api"
	"net/http"
	"path/filepath"
)


func main() {
	Register()
	pluginPath,_:= filepath.Abs("./plugins")
	loadPlugins(pluginPath)
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


	professions, err := eosc.ProfessionConfigs(professionConfig()).Gen(eosc.DefaultProfessionDriverRegister, storeT)
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
