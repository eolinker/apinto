package syslog

import (
	"github.com/eolinker/apinto/output"
	"github.com/eolinker/eosc"
	"reflect"
)

type Syslog struct {
	id   string
	name string

	config *Config
	writer *SysWriter
}

func (s *Syslog) Output(entry eosc.IEntry) error {
	w := s.writer
	if w != nil {
		return w.output(entry)
	}
	return eosc.ErrorWorkerNotRunning
}

func (s *Syslog) Id() string {
	return s.id
}

func (s *Syslog) Start() error {
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

func (s *Syslog) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
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

func (s *Syslog) Stop() error {
	w := s.writer
	if w != nil {
		return w.stop()
	}
	return nil
}
func (s *Syslog) CheckSkill(skill string) bool {
	return output.CheckSkill(skill)
}
