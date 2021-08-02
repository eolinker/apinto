package main

import (
	"fmt"
	"net/http"
	"path/filepath"

	reader_yaml "github.com/eolinker/goku-eosc/reader-yaml"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/log"
	admin_html "github.com/eolinker/eosc/modules/admin-html"
	admin_open_api "github.com/eolinker/eosc/modules/admin-open-api"
)

func main() {
	log.InitDebug(true)
	initFlag()
	Register()
	pluginPath, _ := filepath.Abs("./plugins")
	loadPlugins(pluginPath)
	//storeName := "memory"

	driverFile := "profession.yml"

	//storeDriver, has := eosc.GetStoreDriver(storeName)
	//if !has {
	//	log.Panic("unkonw store driver:", storeName)
	//}
	//
	//storeT, err := storeDriver.Create(nil)
	//if err != nil {
	//	log.Panic(err)
	//}
	storeT := initStore()

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

	if path != "" {
		yamlReader, err := reader_yaml.NewYaml(path)
		if err == nil {
			for _, p := range professions.Infos() {
				values := yamlReader.AllByProfession(p.Name)
				for _, v := range values {
					err = storeT.Set(v)
					if err != nil {
						log.Errorf("init data error	%s	%s	:%s", p.Name, v.Id, err.Error())
						continue
					}
					log.Infof("set data successful	%s	%s", p.Name, v.Id)
				}
			}
		}
	}

	httpServer := http.NewServeMux()
	httpServer.Handle("/api/", handler)
	httpServer.Handle("/", hadlerHtml)

	log.Info(fmt.Sprintf("Listen http port %d successfully", httpPort))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", httpPort), httpServer))
}
