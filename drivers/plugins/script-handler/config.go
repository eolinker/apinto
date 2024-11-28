package script_handler

import (
	"fmt"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

// 参数配置属性
type Config struct {
	Script  string `json:"script" label:"调用的脚本"`
	Package string `json:"package" label:"调用的包名"`
	Fname   string `json:"fname" label:"调用函数名,需定义时返回error"`
	Stage   string `json:"stage" label:"脚本执行阶段，request或response,默认request"`
}

// 初始化插件执行实例
func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	err := conf.doCheck()
	if err != nil {
		return nil, err
	}
	fn, err := getFunc(conf)
	if err != nil {
		return nil, err
	}

	return &Script{
		WorkerBase: drivers.Worker(id, name),
		stage:      conf.Stage,
		fn:         fn,
	}, nil

}

func (c *Config) doCheck() error {
	if c.Script == "" {
		return fmt.Errorf("[plugin script-handler config err] param Script must be not null")
	}
	if c.Package == "" {
		return fmt.Errorf("[plugin script-handler config err] param Package must be not null")
	}
	if c.Fname == "" {
		return fmt.Errorf("[plugin script-handler config err] param Fname must be not null")
	}
	return nil
}
