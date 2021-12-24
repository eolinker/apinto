package fileoutput

import (
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/formatter"
	file_transport "github.com/eolinker/goku/output/file-transport"
)

type FileOutput struct {
	id        string
	cfg       *file_transport.Config
	formatter eosc.IFormatter
	transport formatter.ITransport
}

func (a *FileOutput) Output(entry eosc.IEntry) error {
	if a.formatter != nil {
		data := a.formatter.Format(entry)
		if a.transport != nil && len(data) > 0{
			err := a.transport.Write(data)
			if err != nil {
				return err
			}
			return a.transport.Write([]byte("\n"))
		}
	}
	return nil
}

func (a *FileOutput) Id() string {
	return a.id
}

func (a *FileOutput) Start() error {
	return nil
}

func (a *FileOutput) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) (err error) {
	cfg, ok := conf.(*Config)
	if !ok {
		return errorConfigType
	}
	factory, has := formatter.GetFormatterFactory(cfg.Config.Type)
	if !has {
		return errorFormatterType
	}
	c := &file_transport.Config{
		Dir:    cfg.Config.Dir,
		File:   cfg.Config.File,
		Expire: cfg.Config.Expire,
		Period: file_transport.ParsePeriod(cfg.Config.Period),
	}
	if a.cfg == nil || a.cfg.IsUpdate(c) {
		transport := file_transport.NewtTransporter(c)
		if a.transport != nil {
			a.transport.Close()
		}
		a.transport = transport
		a.cfg = c
	}

	a.formatter, err = factory.Create(cfg.Config.Formatter)
	return
}

func (a *FileOutput) Stop() error {
	a.transport.Close()
	a.transport = nil
	a.formatter = nil
	return nil
}

func (a *FileOutput) CheckSkill(skill string) bool {
	return false
}
