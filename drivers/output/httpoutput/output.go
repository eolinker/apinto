package httpoutput

import (
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/formatter"
	http_transport "github.com/eolinker/goku/output/http-transport"
)

type HttpOutput struct {
	*Driver
	id        string
	config    *HttpConf
	formatter eosc.IFormatter
	transport formatter.ITransport
}

func (h *HttpOutput) Output(entry eosc.IEntry) error {
	if h.formatter != nil {
		data := h.formatter.Format(entry)
		if h.transport != nil && len(data) > 0 {
			err := h.transport.Write(data)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *HttpOutput) Id() string {
	return h.id
}

func (h *HttpOutput) Start() error {
	return nil
}

func (h *HttpOutput) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) (err error) {
	config, err := h.Driver.Check(conf)
	if err != nil {
		return err
	}

	if h.config == nil || h.config.isConfUpdate(config) {
		if h.transport != nil {
			h.transport.Close()
		}
		cfg := &http_transport.Config{
			Method:       config.Method,
			Url:          config.Url,
			Headers:      toHeader(config.Headers),
			HandlerCount: 5, // 默认值， 以后可能会改成配置
		}

		h.transport, err = http_transport.CreateTransporter(cfg)
		if err != nil {
			return err
		}
	}

	//创建formatter
	factory, has := formatter.GetFormatterFactory(config.Type)
	if !has {
		return errFormatterType
	}
	h.formatter, err = factory.Create(config.Formatter)
	if err != nil {
		return err
	}
	h.config = config
	return
}

func (h *HttpOutput) Stop() error {
	h.transport.Close()
	h.transport = nil
	h.formatter = nil
	h.config = nil
	return nil
}

func (h *HttpOutput) CheckSkill(skill string) bool {
	return false
}
