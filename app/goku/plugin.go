package main

import (
	"fmt"
	"github.com/eolinker/eosc/log"
	"path/filepath"
	"plugin"
)

//RegisterFunc 注册函数
type RegisterFunc func()

func loadPlugins(dir string) error {

	files, err := filepath.Glob(fmt.Sprintf("%s/*.so", dir))
	if err != nil {
		return err
	}

	for _, f := range files {

		p, err := plugin.Open(f)
		if err != nil {
			log.Errorf("error to open plugin %s:%s", f, err.Error())
			continue
		}

		r, err := p.Lookup("Register")
		if err != nil {
			log.Errorf("call register from  plugin : %s : %s", f, err.Error())
			continue
		}

		r.(RegisterFunc)()
	}
	return nil
}
