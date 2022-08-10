package syslog

import (
	"github.com/eolinker/apinto/output"
	"github.com/eolinker/eosc"
	"reflect"
)

var _ output.IEntryOutput = (*Output)(nil)
var _ eosc.IWorker = (*Output)(nil)

type Output struct {
	id   string
	name string

	config *Config
	writer *SysWriter
}

func (s *Output) Output(entry eosc.IEntry) error {
	w := s.writer
	if w != nil {
		return w.output(entry)
	}
	return eosc.ErrorWorkerNotRunning
}

func (s *Output) Id() string {
	return s.id
}

func (s *Output) Start() error {
	w := s.writer
	if w != nil {
		return nil
	}
	writer, err := CreateTransporter(s.config)
	if err != nil {
		return err
	}
	s.writer = writer
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
	w := s.writer
	if w != nil {
		w.reset(cfg)
	}
	return nil
}

func (s *Output) Stop() error {
	w := s.writer
	if w != nil {
		return w.stop()
	}
	return nil
}
func (s *Output) CheckSkill(skill string) bool {
	return output.CheckSkill(skill)
}
