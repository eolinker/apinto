package main

import (
	"fmt"
	"net/http"
	"path/filepath"
	"runtime"

	reader_yaml "github.com/eolinker/goku/reader-yaml"

	_ "net/http/pprof"

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

	if dataPath != "" {
		yamlReader, err := reader_yaml.NewYaml(dataPath)
		if err != nil {
			log.Warnf("load %s:%s", dataPath, err.Error())
			log.Panic(err)
		}

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

	runtime.GOMAXPROCS(4)              // 限制 CPU 使用数，避免过载
	runtime.SetMutexProfileFraction(1) // 开启对锁调用的跟踪
	runtime.SetBlockProfileRate(1)     // 开启对阻塞操作的跟踪
	go http.ListenAndServe("0.0.0.0:6060", nil)
	httpServer := http.NewServeMux()
	httpServer.Handle("/api/", handler)
	httpServer.Handle("/", hadlerHtml)

	go func() {
		log.Info(fmt.Sprintf("Listen http port %d successfully", httpPort))
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", httpPort), httpServer))
	}()

	if httpsPemPath != "" && httpsKeyPath != "" {
		go func() {
			log.Info(fmt.Sprintf("Listen https port %d successfully", httpsPort))
			log.Fatal(http.ListenAndServeTLS(fmt.Sprintf(":%d", httpsPort), httpsPemPath, httpsKeyPath, httpServer))
		}()
	}
	select {}
}
