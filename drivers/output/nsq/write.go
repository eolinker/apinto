package nsq

import (
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/formatter"
	"sync"
)

type Writer struct {
	pool      *producerPool
	topic     string
	formatter eosc.IFormatter

	lock sync.Mutex
}

func NewWriter(conf *Config) *Writer {

	w := &Writer{}
	w.reset(conf)
	return w
}

func (n *Writer) reset(config *Config) error {

	//创建生产者pool
	pool, err := CreateProducerPool(config.Address, config.AuthSecret, config.ClientConf)
	if err != nil {
		return err
	}

	//创建formatter
	factory, has := formatter.GetFormatterFactory(config.Type)
	if !has {
		return errFormatterType
	}
	fm, err := factory.Create(config.Formatter)
	if err != nil {
		return err
	}
	n.lock.Lock()
	defer n.lock.Unlock()

	op := n.pool
	n.pool = pool
	n.topic = config.Topic
	n.formatter = fm
	if op != nil {
		op.Close()
	}
	return nil
}
func (n *Writer) stop() error {
	n.lock.Lock()
	defer n.lock.Unlock()

	n.pool.Close()
	n.formatter = nil
	n.pool = nil

	return nil
}

func (n *Writer) output(entry eosc.IEntry) error {
	if n.formatter != nil {
		data := n.formatter.Format(entry)
		if n.pool != nil && len(data) > 0 {
			err := n.pool.PublishAsync(n.topic, data)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
