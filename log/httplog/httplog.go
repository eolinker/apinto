package httplog

import (
	"fmt"
	"github.com/eolinker/eosc"
	"github.com/eolinker/goku/log"
	logFormatter "github.com/eolinker/goku/log/common/log-formatter"
)

type httplog struct {
	id                 string
	name               string
	config             *Config
	formatterName      string
	transporterManager transporterManager.ITransporterManager
}

func (h *httplog) Id() string {
	return h.id
}

func (h *httplog) Start() error {
	formatter := logFormatter.CreateFormatter(driverName, h.formatterName)
	transporterReset, err := createTransporter(h.config, formatter)
	if err != nil {
		return err
	}

	return h.transporterManager.Set(h.id, transporterReset)
}

func (h *httplog) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	config, ok := conf.(*DriverConfig)
	if !ok {
		return fmt.Errorf("need %s,now %s", eosc.TypeNameOf((*DriverConfig)(nil)), eosc.TypeNameOf(conf))
	}

	c, err := toConfig(config)
	if err != nil {
		return err
	}
	h.config = c
	h.formatterName = config.FormatterName

	formatter := logFormatter.CreateFormatter(driverName, h.formatterName)
	transporter, err := createTransporter(h.config, formatter)
	if err != nil {
		return err
	}

	return h.transporterManager.Set(h.id, transporter)
}

func (h *httplog) Stop() error {
	return h.transporterManager.Del(h.id)
}

func (h *httplog) CheckSkill(skill string) bool {
	return log.CheckSkill(skill)
}
