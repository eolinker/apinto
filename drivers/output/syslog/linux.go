//go:build !windows && !plan9
// +build !windows,!plan9

package syslog

import (
	"github.com/eolinker/apinto/output"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/formatter"
	sys "log/syslog"
	"strings"
)

//CreateTransporter 创建syslog-Transporter
func CreateTransporter(conf *Config) (*SysWriter, error) {
	sysWriter, err := newSysWriter(conf, "")
	if err != nil {
		return nil, err
	}
	return &SysWriter{
		writer: sysWriter,
	}, nil
}

const defaultTag = "apinto"

type SysWriter struct {
	*Driver
	id        string
	writer    *sys.Writer
	formatter eosc.IFormatter
}

func (s *SysWriter) Output(entry eosc.IEntry) error {
	if s.formatter != nil {
		data := s.formatter.Format(entry)
		if s.writer != nil && len(data) > 0 {
			_, err := s.writer.Write(data)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *SysWriter) Id() string {
	return s.id
}
func (s *SysWriter) Stop() error {
	err := s.writer.Close()
	if err != nil {
		return err
	}
	s.writer = nil
	s.formatter = nil
	return nil
}

func (s *SysWriter) Start() error {
	return nil
}

func (s *SysWriter) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	cfg, err := s.Driver.check(conf)
	if err != nil {
		return err
	}
	// 新建formatter
	factory, has := formatter.GetFormatterFactory(cfg.Type)
	if !has {
		return errFormatterType
	}
	s.formatter, err = factory.Create(cfg.Formatter)
	// 关闭旧的
	if s.writer != nil {
		err = s.writer.Close()
		if err != nil {
			return err
		}
	}
	w, err := newSysWriter(cfg, "")
	if err != nil {
		return err
	}
	s.writer = w
	return nil
}

func (s *SysWriter) CheckSkill(skill string) bool {
	return output.CheckSkill(skill)
}

func newSysWriter(conf *Config, tag string) (*sys.Writer, error) {
	if tag == "" {
		tag = defaultTag
	}
	return sys.Dial(strings.ToLower(conf.Network), conf.Address, parseLevel(conf.Level), tag)
}

func parseLevel(level string) sys.Priority {
	switch strings.ToLower(level) {
	case "error":
		return sys.LOG_ERR
	case "warn", "warning":
		return sys.LOG_WARNING
	case "info":
		return sys.LOG_INFO
	case "debug", "trace":
		return sys.LOG_DEBUG
	}
	return sys.LOG_ERR
}
