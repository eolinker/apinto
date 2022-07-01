package fileoutput

import (
	"github.com/eolinker/apinto/output"
	file_transport "github.com/eolinker/apinto/output/file-transport"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/formatter"
)

type FileOutput struct {
	*Driver
	id        string
	cfg       *file_transport.Config
	formatter eosc.IFormatter
	transport formatter.ITransport
}

func (a *FileOutput) Output(entry eosc.IEntry) error {
	if a.formatter != nil {
		data := a.formatter.Format(entry)
		if a.transport != nil && len(data) > 0 {
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
	cfg, err := a.Driver.Check(conf)
	if err != nil {
		return err
	}
	factory, has := formatter.GetFormatterFactory(cfg.Type)
	if !has {
		return errorFormatterType
	}
	c := &file_transport.Config{
		Dir:    cfg.Dir,
		File:   cfg.File,
		Expire: cfg.Expire,
		Period: file_transport.ParsePeriod(cfg.Period),
	}
	if a.cfg == nil || a.cfg.IsUpdate(c) {
		transport := file_transport.NewtTransporter(c)
		if a.transport != nil {
			a.transport.Close()
		}
		a.transport = transport
		a.cfg = c
	}

	a.formatter, err = factory.Create(cfg.Formatter)
	return
}

func (a *FileOutput) Stop() error {
	a.transport.Close()
	a.transport = nil
	a.formatter = nil
	return nil
}

func (a *FileOutput) CheckSkill(skill string) bool {
	return output.CheckSkill(skill)
}
