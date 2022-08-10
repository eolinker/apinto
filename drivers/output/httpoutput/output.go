package httpoutput

import (
	"github.com/eolinker/apinto/output"
	"github.com/eolinker/eosc"
)

type HttpOutput struct {
	id      string
	config  *Config
	handler *Handler
}

func (h *HttpOutput) Output(entry eosc.IEntry) error {
	hd := h.handler
	if hd != nil {
		return hd.Output(entry)
	}

	return eosc.ErrorWorkerNotRunning
}

func (h *HttpOutput) Id() string {
	return h.id
}

func (h *HttpOutput) Start() error {
	hd := h.handler
	if hd != nil {
		return nil
	}
	handler, err := NewHandler(h.config)
	if err != nil {
		return err
	}
	h.handler = handler
	return nil
}

func (h *HttpOutput) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) (err error) {
	config, err := Check(conf)
	if err != nil {
		return err
	}
	if h.config != nil && !h.config.isConfUpdate(config) {
		return nil
	}
	h.config = config
	hd := h.handler

	if hd != nil {
		return hd.reset(config)
	}

	return nil
}

func (h *HttpOutput) Stop() error {
	hd := h.handler
	if hd != nil {
		h.handler = nil
		hd.Close()
	}
	h.config = nil
	return nil
}

func (h *HttpOutput) CheckSkill(skill string) bool {
	return output.CheckSkill(skill)
}
