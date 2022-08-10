package httpoutput

import (
	http_transport "github.com/eolinker/apinto/output/http-transport"
	"github.com/eolinker/eosc/formatter"
	"github.com/eolinker/eosc/utils/schema"
	"reflect"

	"github.com/eolinker/eosc"
)

type Driver struct {
	configType reflect.Type
}

func (d *Driver) ConfigType() reflect.Type {
	return d.configType
}

func (d *Driver) Render() interface{} {
	render, err := schema.Generate(reflect.TypeOf((*Config)(nil)), nil)
	if err != nil {
		return nil
	}
	return render
}

func (d *Driver) Check(v interface{}) (*Config, error) {
	conf, ok := v.(*Config)
	if !ok {
		return nil, errConfigType
	}

	httpConf := conf
	if httpConf.Method == "" {
		return nil, errMethod
	}
	switch httpConf.Method {
	case "GET", "POST", "HEAD", "PUT", "DELETE", "CONNECT", "OPTIONS", "TRACE":
	default:
		return nil, errMethod
	}

	if httpConf.Url == "" {
		return nil, errUrlNull
	}

	if httpConf.Type == "" {
		httpConf.Type = "line"
	}

	switch httpConf.Type {
	case "line", "json":
	default:
		return nil, errFormatterType
	}

	if len(httpConf.Formatter) == 0 {
		return nil, errFormatterConf
	}

	return httpConf, nil
}

func (d *Driver) Create(id, name string, v interface{}, workers map[eosc.RequireId]interface{}) (eosc.IWorker, error) {
	worker := &HttpOutput{
		Driver: d,
		id:     id,
	}

	conf, err := d.Check(v)
	if err != nil {
		return nil, err
	}

	worker.config = conf

	cfg := &http_transport.Config{
		Method:       conf.Method,
		Url:          conf.Url,
		Headers:      toHeader(conf.Headers),
		HandlerCount: 5, // 默认值， 以后可能会改成配置
	}

	worker.transport, err = http_transport.CreateTransporter(cfg)
	if err != nil {
		return nil, err
	}

	//创建formatter
	factory, has := formatter.GetFormatterFactory(conf.Type)
	if !has {
		return nil, errFormatterType
	}
	worker.formatter, err = factory.Create(conf.Formatter)

	return worker, err
}
