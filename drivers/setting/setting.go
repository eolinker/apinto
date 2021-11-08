package setting

import (
	"errors"

	"github.com/eolinker/eosc"
)

type Worker struct {
	id      string
	conf    interface{}
	manager IManager
}

func (w *Worker) setManager(manager IManager) {
	w.manager = manager
}

func (w *Worker) Id() string {
	return w.id
}

func (w *Worker) Start() error {
	if w.manager != nil {
		return errors.New("")
	}
	err := w.manager.Check(w.conf)
	if err != nil {
		return err
	}

	return w.manager.Reset(w.conf)
}

func (w *Worker) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	if w.manager != nil {
		return errors.New("")
	}
	err := w.manager.Check(conf)
	if err != nil {
		return err
	}
	err = w.manager.Reset(conf)
	if err != nil {
		return err
	}
	w.conf = conf
	return nil
}

func (w *Worker) Stop() error {
	return nil
}

func (w *Worker) CheckSkill(skill string) bool {
	return true
}

type IManager interface {
	Reset(conf interface{}) error
	Check(conf interface{}) error
}
