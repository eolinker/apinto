package fileoutput

import (
	"fmt"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/formatter"
	"github.com/eolinker/eosc/log/filelog"
	"github.com/eolinker/eosc/router"
	"net/http"

	"time"
)

type FileWriter struct {
	formatter eosc.IFormatter
	transport *filelog.FileWriterByPeriod
	//id        string

	fileHandler http.Handler
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

func (a *FileWriter) reset(cfg *Config, name string) (err error) {

	factory, has := formatter.GetFormatterFactory(cfg.Type)
	if !has {
		return errorFormatterType
	}

	fm, err := factory.Create(cfg.Formatter)
	if err != nil {
		return err
	}

	transport := a.transport
	c := filelog.Config{
		Dir:    cfg.Dir,
		File:   cfg.File,
		Expire: time.Duration(cfg.Expire) * 24 * time.Hour,
		Period: filelog.ParsePeriod(cfg.Period),
	}
	if transport == nil {
		transport = filelog.NewFileWriteByPeriod(c)
		a.fileHandler = transport.ServeHTTP(fmt.Sprintf("/%slog/access/%s", router.RouterPrefix, name))
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
