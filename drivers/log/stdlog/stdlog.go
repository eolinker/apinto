package stdlog

import (
	"fmt"

	"github.com/eolinker/eosc"
	transporterManager "github.com/eolinker/eosc/log/transporter-manager"
)

type stdlog struct {
	id                 string
	name               string
	config             *stdlog_transporter.Config
	formatterName      string
	transporterReset   log_transport.TransporterReset
	transporterManager transporterManager.ITransporterManager
}

func (h *stdlog) Id() string {
	return h.id
}

func (h *stdlog) Start() error {
	formatter, err := stdlog_transporter.CreateFormatter(h.formatterName)
	if err != nil {
		return err
	}

	transporterReset, err := stdlog_transporter.CreateTransporter(h.config, formatter)
	if err != nil {
		return err
	}
	h.transporterReset = transporterReset
	return h.transporterManager.Set(h.id, transporterReset)
}

func (h *stdlog) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
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

	formatter, err := stdlog_transporter.CreateFormatter(h.formatterName)
	if err != nil {
		return err
	}
	err = h.transporterReset.Reset(c, formatter)
	if err != nil {
		return err
	}

	return nil
}

func (h *stdlog) Stop() error {
	err := h.transporterReset.Close()
	if err != nil {
		return err
	}
	return h.transporterManager.Del(h.id)
}

func (h *stdlog) CheckSkill(skill string) bool {
	return log_transport.CheckSkill(skill)
}
