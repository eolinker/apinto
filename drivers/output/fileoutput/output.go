package fileoutput

import (
	"fmt"
	scope_manager "github.com/eolinker/apinto/scope-manager"
	"github.com/eolinker/eosc/router"
	"net/http"
	"reflect"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/output"
	"github.com/eolinker/eosc"
)

var _ output.IEntryOutput = (*FileOutput)(nil)
var _ eosc.IWorker = (*FileOutput)(nil)

type FileOutput struct {
	drivers.WorkerBase
	config    *Config
	writer    *FileWriter
	isRunning bool
}

func (a *FileOutput) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if a.writer == nil || a.writer.fileHandler == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	a.writer.fileHandler.ServeHTTP(w, r)
}

func (a *FileOutput) Output(entry eosc.IEntry) error {
	w := a.writer
	if w != nil {
		return w.output(entry)
	}
	return eosc.ErrorWorkerNotRunning
}

func (a *FileOutput) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) (err error) {

	cfg, err := check(conf)

	if err != nil {
		return err
	}
	if reflect.DeepEqual(cfg, a.config) {
		return nil
	}
	a.config = cfg

	if a.isRunning {
		w := a.writer
		if w == nil {
			w = new(FileWriter)
		}

		err = w.reset(cfg, a.Name())
		if err != nil {
			return err
		}
		a.writer = w
		scope_manager.Set(a.Id(), a, cfg.Scopes...)
	}

	return nil
}

func (a *FileOutput) Stop() error {
	scope_manager.Del(a.Id())
	router.DeletePath(a.Id())
	a.isRunning = false

	w := a.writer
	if w != nil {
		err := w.stop()
		a.writer = nil
		return err
	}
	return nil
}

func (a *FileOutput) Start() error {
	a.isRunning = true
	w := a.writer
	if w == nil {
		w = new(FileWriter)
	}
	err := router.SetPath(a.Id(), fmt.Sprintf("/log/access-log/%s", a.Name()), a)
	if err != nil {
		return err
	}
	err = w.reset(a.config, a.Name())
	if err != nil {
		return err
	}

	a.writer = w
	scope_manager.Set(a.Id(), a, a.config.Scopes...)

	return nil

}

func (a *FileOutput) CheckSkill(skill string) bool {
	return output.CheckSkill(skill)
}
