package fileoutput

import (
	"github.com/eolinker/apinto/output"
	"github.com/eolinker/eosc"
	"reflect"
)

var _ output.IEntryOutput = (*FileOutput)(nil)
var _ eosc.IWorker = (*FileOutput)(nil)

type FileOutput struct {
	id        string
	name      string
	config    *Config
	writer    *FileWriter
	isRunning bool
}

func (a *FileOutput) Output(entry eosc.IEntry) error {
	w := a.writer
	if w != nil {
		return w.output(entry)
	}
	return eosc.ErrorWorkerNotRunning
}

func (a *FileOutput) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) (err error) {

	cfg, err := Check(conf)

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
			a.writer = new(FileWriter)
			w = a.writer
		}
		return w.reset(cfg)
	}

	return nil
}

func (a *FileOutput) Stop() error {
	a.isRunning = false
	w := a.writer
	if w != nil {
		err := w.stop()
		a.writer = nil
		return err
	}
	return nil
}

func (a *FileOutput) Id() string {
	return a.id
}

func (a *FileOutput) Start() error {
	a.isRunning = true
	w := a.writer
	if w == nil {
		w = new(FileWriter)
		a.writer = w
	}

	return w.reset(a.config)

}

func (a *FileOutput) CheckSkill(skill string) bool {
	return output.CheckSkill(skill)
}
