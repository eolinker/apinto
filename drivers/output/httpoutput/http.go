package httpoutput

import (
	http_transport "github.com/eolinker/apinto/output/http-transport"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/formatter"
)

type Handler struct {
	formatter eosc.IFormatter
	transport formatter.ITransport
}

func NewHandler(config *Config) (*Handler, error) {
	transport, fm, err := create(config)
	if err != nil {
		return nil, err
	}
	h := &Handler{
		formatter: fm,
		transport: transport,
	}
	return h, nil
}

func (h *Handler) Close() error {
	h.transport.Close()
	h.transport = nil
	h.formatter = nil
	return nil
}

func (h *Handler) Output(entry eosc.IEntry) error {
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
func (h *Handler) reset(config *Config) error {

	if h.transport != nil {
		h.transport.Close()
	}
	transport, fm, err := create(config)
	if err != nil {
		return err
	}
	h.transport = transport
	h.formatter = fm
	return nil
}

func create(config *Config) (formatter.ITransport, eosc.IFormatter, error) {
	cfg := &http_transport.Config{
		Method:       config.Method,
		Url:          config.Url,
		Headers:      toHeader(config.Headers),
		HandlerCount: 5, // 默认值， 以后可能会改成配置
	}
	transport, err := http_transport.CreateTransporter(cfg)
	if err != nil {
		return nil, nil, err
	}

	//创建formatter
	factory, has := formatter.GetFormatterFactory(config.Type)
	if !has {
		return nil, nil, errFormatterType
	}
	fm, err := factory.Create(config.Formatter)
	if err != nil {
		return nil, nil, err
	}
	return transport, fm, nil
}
