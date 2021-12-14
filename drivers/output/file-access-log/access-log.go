package file_access_log

import (
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/formatter"
	http_service "github.com/eolinker/eosc/http-service"
	file_transport "github.com/eolinker/goku/transport/file-transport"
)

type accessLog struct {
	id        string
	cfg       *file_transport.Config
	formatter formatter.Config
	transport formatter.ITransport
}

func (a *accessLog) Output(context http_service.IHttpContext) error {
	return nil
}

func (a *accessLog) Id() string {
	return a.id
}

func (a *accessLog) Start() error {
	return nil
}

func (a *accessLog) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	cfg, ok := conf.(*Config)
	if !ok {
		return errorConfigType
	}
	c := &file_transport.Config{
		Dir:    cfg.Dir,
		File:   cfg.File,
		Expire: cfg.Expire,
		Period: file_transport.ParsePeriod(cfg.Period),
	}
	if a.cfg.IsUpdate(c) {
		transport := file_transport.NewtTransporter(c)
		a.transport.Close()
		a.transport = transport
		a.cfg = c
	}
	a.formatter = cfg.Formatter
	return nil
}

func (a *accessLog) Stop() error {
	a.transport.Close()
	a.transport = nil
	return nil
}

func (a *accessLog) CheckSkill(skill string) bool {
	return false
}
