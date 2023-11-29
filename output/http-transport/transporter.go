package http_transport

import (
	"github.com/eolinker/eosc/formatter"
)

type Transporter struct {
	writer *_HttpWriter
}

func (t *Transporter) Close() error {
	return t.writer.Close()
}

func (t *Transporter) Write(data []byte) error {
	_, err := t.writer.Write(data)
	return err
}

func CreateTransporter(conf *Config) (formatter.ITransport, error) {
	httpWriter := newHttpWriter()
	transport := &Transporter{
		writer: httpWriter,
	}
	transport.writer.reset(conf)

	return transport, nil
}
