//+build !windows,!plan9

package syslog

import (
	"log/syslog"
	"strings"

	"github.com/eolinker/eosc/log"
)

const defaultTag = "goku"

type _SysWriter struct {
	network string
	url     string
	level   syslog.Priority
	writer  *syslog.Writer
}

func newSysWriter(network string, url string, level log.Level, tag string) (*_SysWriter, error) {
	if tag == "" {
		tag = defaultTag
	}
	writer, err := syslog.Dial(strings.ToLower(network), url, parseLevel(level), tag)
	if err != nil {
		return nil, err
	}
	return &_SysWriter{network: network, url: url, writer: writer}, nil
}

func (s *_SysWriter) Write(p []byte) (n int, err error) {
	//不需要加锁，syslog.Writer的write方法已加上锁
	s.writer.Write(p)
	return len(p), nil
}

func (s *_SysWriter) Close() error {
	//不需要加锁，syslog.Writer的close方法已加上锁
	return s.writer.Close()
}

func parseLevel(level log.Level) syslog.Priority {
	switch level {
	case log.ErrorLevel:
		{
			return syslog.LOG_ERR
		}
	case log.WarnLevel:
		{
			return syslog.LOG_WARNING
		}
	case log.InfoLevel:
		{
			return syslog.LOG_INFO
		}
	case log.DebugLevel, log.TraceLevel:
		{
			return syslog.LOG_DEBUG
		}
	}
	return syslog.LOG_ERR
}
