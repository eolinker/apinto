package syslog

import (
	"fmt"
	"github.com/eolinker/eosc"
	"github.com/eolinker/goku/log"
	logFormatter "github.com/eolinker/goku/log/common/log-formatter"
)

type syslog struct {
	id                 string
	name               string
	config             *Config
	formatterName      string
	transporterManager transporterManager.ITransporterManager
}

func (s *syslog) Id() string {
	return s.id
}

func (s *syslog) Start() error {
	formatter := logFormatter.CreateFormatter(driverName, s.formatterName)
	transporterReset, err := createTransporter(s.config, formatter)
	if err != nil {
		return err
	}

	return s.transporterManager.Set(s.id, transporterReset)
}

func (s *syslog) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	config, ok := conf.(*DriverConfig)
	if !ok {
		return fmt.Errorf("need %s,now %s", eosc.TypeNameOf((*DriverConfig)(nil)), eosc.TypeNameOf(conf))
	}

	c, err := toConfig(config)
	if err != nil {
		return err
	}
	s.config = c
	s.formatterName = config.FormatterName

	formatter := logFormatter.CreateFormatter(driverName, s.formatterName)
	transporter, err := createTransporter(s.config, formatter)
	if err != nil {
		return err
	}

	return s.transporterManager.Set(s.id, transporter)
}

func (s *syslog) Stop() error {
	return s.transporterManager.Del(s.id)
}

func (s *syslog) CheckSkill(skill string) bool {
	return log.CheckSkill(skill)
}
