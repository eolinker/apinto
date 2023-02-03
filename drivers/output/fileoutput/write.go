package fileoutput

import (
	file_transport "github.com/eolinker/apinto/output/file-transport"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/formatter"
)

type FileWriter struct {
	formatter eosc.IFormatter
	transport *file_transport.FileWriterByPeriod
	//id        string
}

func (a *FileWriter) output(entry eosc.IEntry) error {

	if a.formatter != nil && a.transport != nil {
		data := a.formatter.Format(entry)
		if len(data) > 0 {
			_, err := a.transport.Write(data)
			return err
		}
	}
	return nil
}

func (a *FileWriter) reset(cfg *Config) (err error) {

	factory, has := formatter.GetFormatterFactory(cfg.Type)
	if !has {
		return errorFormatterType
	}

	fm, err := factory.Create(cfg.Formatter)
	if err != nil {
		return err
	}

	transport := a.transport
	c := &file_transport.Config{
		Dir:    cfg.Dir,
		File:   cfg.File,
		Expire: cfg.Expire,
		Period: file_transport.ParsePeriod(cfg.Period),
	}
	if transport == nil {
		transport = file_transport.NewFileWriteByPeriod(c)
	} else {
		transport.Reset(c)
	}

	a.transport = transport
	a.formatter = fm

	return
}
func (a *FileWriter) stop() error {

	a.transport.Close()
	a.transport = nil
	a.formatter = nil
	return nil
}
