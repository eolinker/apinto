package fileoutput

import (
	"github.com/eolinker/apinto/output"
	file_transport "github.com/eolinker/apinto/output/file-transport"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/formatter"
	"sync"
)

type FileOutput struct {
	*Driver
	id        string
	cfg       *file_transport.Config
	formatter eosc.IFormatter
	transport *file_transport.FileWriterByPeriod
	rmu       sync.RWMutex
}

func (a *FileOutput) Output(entry eosc.IEntry) error {
	a.rmu.RLock()
	defer a.rmu.RUnlock()
	if a.formatter != nil {
		data := a.formatter.Format(entry)
		if a.transport != nil && len(data) > 0 {
			_, err := a.transport.Write(data)
			return err
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
	a.rmu.Lock()
	defer a.rmu.Unlock()
	if a.cfg == nil {
		a.transport = file_transport.NewFileWriteByPeriod(c)
	} else if a.cfg.IsUpdate(c) {
		a.transport.Reset(c)
	}
	a.cfg = c

	a.formatter, err = factory.Create(cfg.Formatter)
	return
}

func (a *FileOutput) Stop() error {
	a.rmu.Lock()
	defer a.rmu.Unlock()
	a.transport.Close()
	a.transport = nil
	a.formatter = nil
	return nil
}

func (a *FileOutput) CheckSkill(skill string) bool {
	return output.CheckSkill(skill)
}
