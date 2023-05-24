package syslog

import (
	scope_manager "github.com/eolinker/apinto/scope-manager"
	"reflect"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/output"
	"github.com/eolinker/eosc"
)

var _ output.IEntryOutput = (*Output)(nil)
var _ eosc.IWorker = (*Output)(nil)

type Output struct {
	drivers.WorkerBase

	config  *Config
	writer  *SysWriter
	running bool
}

func (s *Output) Output(entry eosc.IEntry) error {
	w := s.writer
	if w != nil {
		return w.output(entry)
	}
	return eosc.ErrorWorkerNotRunning
}

func (s *Output) Start() error {
	scope_manager.Del(s.Id())
	s.running = true
	w := s.writer
	if w == nil {
		writer, err := CreateTransporter(s.config)
		if err != nil {
			return err
		}
		s.writer = writer
	}
	scope_manager.Set(s.Id(), s, s.config.Scopes...)
	return nil
}

func (s *Output) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cfg, err := check(conf)
	if err != nil {
		return err
	}
	// 新建formatter
	if reflect.DeepEqual(cfg, s.config) {
		return nil
	}
	s.config = cfg

	if s.running {
		w := s.writer
		if w != nil {
			return w.reset(cfg)
		}
		writer, err := CreateTransporter(s.config)
		if err != nil {
			return err
		}
		s.writer = writer
	}
	scope_manager.Set(s.Id(), s, s.config.Scopes...)
	return nil
}

func (s *Output) Stop() error {
	w := s.writer
	scope_manager.Del(s.Id())
	if w != nil {
		return w.stop()
	}
	return nil
}
func (s *Output) CheckSkill(skill string) bool {
	return output.CheckSkill(skill)
}
