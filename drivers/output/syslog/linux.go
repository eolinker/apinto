//go:build !windows && !plan9
// +build !windows,!plan9

package syslog

import (
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/formatter"
	sys "log/syslog"
	"strings"
)

//CreateTransporter 创建syslog-Transporter
func CreateTransporter(conf *Config) (*SysWriter, error) {
	fm, w, err := create(conf)
	if err != nil {
		return nil, err
	}

	return &SysWriter{
		writer:    w,
		formatter: fm,
	}, nil
}

const defaultTag = "apinto"

type SysWriter struct {
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

func (s *SysWriter) Stop() error {
	err := s.writer.Close()
	if err != nil {
		return err
	}
	s.writer = nil
	s.formatter = nil
	return nil
}
func create(cfg *Config) (eosc.IFormatter, *sys.Writer, error) {
	factory, has := formatter.GetFormatterFactory(cfg.Type)
	if !has {
		return nil, nil, errFormatterType
	}
	fm, err := factory.Create(cfg.Formatter)
	if err != nil {
		return nil, nil, err
	}
	w, err := newSysWriter(cfg, "")
	if err != nil {
		return nil, nil, err
	}
	return fm, w, nil
}
func (s *SysWriter) Reset(cfg *Config) error {

	fm, w, err := create(cfg)
	if err != nil {
		return err
	}
	o := s.writer
	s.formatter, s.writer = fm, w
	if o != nil {
		o.Close()
	}
	return nil
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
