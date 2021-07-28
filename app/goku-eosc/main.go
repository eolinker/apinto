package main

import (
	"net/http"
	"path/filepath"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/log"
	admin_html "github.com/eolinker/eosc/modules/admin-html"
	admin_open_api "github.com/eolinker/eosc/modules/admin-open-api"
)

func main() {
	log.InitDebug(true)
	Register()
	pluginPath, _ := filepath.Abs("./plugins")
	loadPlugins(pluginPath)
	storeName := "memory"

	driverFile := "profession.yml"

	storeDriver, has := eosc.GetStoreDriver(storeName)
	if !has {
		log.Panic("unkonw store driver:", storeName)
	}

	storeT, err := storeDriver.Create(nil)
	if err != nil {
		log.Panic(err)
	}

	err = storeT.Initialization()
	if err != nil {
		log.Panic(err)
	}
	driverCfg, err := readProfessionConfig(driverFile)
	if err != nil {
		log.Panic(err)
	}
	professions, err := eosc.ProfessionConfigs(driverCfg).Gen(eosc.DefaultProfessionDriverRegister, storeT)
	if err != nil {
		panic(err)
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
