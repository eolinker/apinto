package syslog

import (
	"fmt"
	log_transport "github.com/eolinker/goku/log-transport"
	syslog_transporter "github.com/eolinker/goku/log-transport/syslog"

	"github.com/eolinker/eosc"
	transporterManager "github.com/eolinker/eosc/log/transporter-manager"
)

type syslog struct {
	id                 string
	name               string
	config             *syslog_transporter.Config
	formatterName      string
	transporterReset   log_transport.TransporterReset
	transporterManager transporterManager.ITransporterManager
}

func (s *syslog) Id() string {
	return s.id
}

func (s *syslog) Start() error {
	formatter, err := syslog_transporter.CreateFormatter(s.formatterName)
	if err != nil {
		return err
	}
	transporterReset, err := syslog_transporter.CreateTransporter(s.config, formatter)
	if err != nil {
		return err
	}

	s.transporterReset = transporterReset

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

	formatter, err := syslog_transporter.CreateFormatter(s.formatterName)
	if err != nil {
		return err
	}

	err = s.transporterReset.Reset(c, formatter)
	if err != nil {
		return err
	}

	return nil
}

func (s *syslog) Stop() error {
	err := s.transporterReset.Close()
	if err != nil {
		return err
	}
	return s.transporterManager.Del(s.id)
}

func (s *syslog) CheckSkill(skill string) bool {
	return false
}
