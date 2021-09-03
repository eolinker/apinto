package httplog

import (
	"fmt"
	log_transport "github.com/eolinker/goku/log-transport"
	httplog_transporter "github.com/eolinker/goku/log-transport/httplog"

	"github.com/eolinker/eosc"
	transporterManager "github.com/eolinker/eosc/log/transporter-manager"
)

type httplog struct {
	id                 string
	name               string
	config             *httplog_transporter.Config
	formatterName      string
	transporterReset   log_transport.TransporterReset
	transporterManager transporterManager.ITransporterManager
}

func (h *httplog) Id() string {
	return h.id
}

func (h *httplog) Start() error {
	formatter, err := CreateFormatter(h.formatterName)
	if err != nil {
		return err
	}

	transporterReset := httplog_transporter.CreateTransporter(h.config.Level)
	err = transporterReset.Reset(h.config,formatter)
	if err != nil {
		return err
	}
	h.transporterReset = transporterReset
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

	formatter, err := CreateFormatter(h.formatterName)
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
	return false
}
