//go:build windows
// +build windows

package syslog

import (
	"fmt"
	"github.com/eolinker/eosc"
)

//CreateTransporter 创建windows下的syslog，windows下不支持syslog，直接返回错误
func CreateTransporter(conf *SysConfig) (*SysWriter, error) {
	return nil, fmt.Errorf("can not create syslog transporterReset in windows system")
}

const defaultTag = "apinto"

type SysWriter struct {
	*Driver
	id        string
	formatter eosc.IFormatter
}

func (s *SysWriter) Id() string {
	return ""
}
func (s *SysWriter) Stop() error {
	return fmt.Errorf("can not create syslog transporterReset in windows system")
}

func (s *SysWriter) Start() error {
	return fmt.Errorf("can not create syslog transporterReset in windows system")
}

func (s *SysWriter) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	return fmt.Errorf("can not create syslog transporterReset in windows system")
}

func (s *SysWriter) CheckSkill(skill string) bool {
	return false
}
