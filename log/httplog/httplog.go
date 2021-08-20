package httplog

import (
	"fmt"
	"github.com/eolinker/eosc"
	transporterManager "github.com/eolinker/eosc/log/transporter-manager"
	"github.com/eolinker/goku/log"
	"github.com/eolinker/goku/log/httplog/httplog-transporter"
)

type httplog struct {
	id                 string
	name               string
	config             *httplog_transporter.Config
	formatterName      string
	transporterReset   log.TransporterReset
	transporterManager transporterManager.ITransporterManager
}

func (h *httplog) Id() string {
	return h.id
}

func (h *httplog) Start() error {
	formatter, err := httplog_transporter.CreateFormatter(h.formatterName)
	if err != nil {
		return err
	}

	transporterReset, err := httplog_transporter.CreateTransporter(h.config, formatter)
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

	formatter, err := httplog_transporter.CreateFormatter(h.formatterName)
	if err != nil {
		return err
	}
	err = h.transporterReset.Reset(c, formatter)
	if err != nil {
		return err
	}

	return nil
}

func (h *httplog) Stop() error {
	err := h.transporterReset.Close()
	if err != nil {
		return err
	}
	return h.transporterManager.Del(h.id)
}

func (h *httplog) CheckSkill(skill string) bool {
	return log.CheckSkill(skill)
}
