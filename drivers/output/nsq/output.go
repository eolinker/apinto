package nsq

import (
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/formatter"
	"sync"
)

type NsqOutput struct {
	*Driver
	id        string
	pool      *producerPool
	topic     string
	formatter eosc.IFormatter

	lock sync.Mutex
}

func (n *NsqOutput) Id() string {
	return n.id
}

func (n *NsqOutput) Start() error {
	return nil
}

func (n *NsqOutput) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	config, err := n.Driver.Check(conf)
	if err != nil {
		return err
	}

	n.lock.Lock()
	defer n.lock.Unlock()

	if n.pool != nil {
		n.pool.Close()
	}

	n.topic = config.Topic
	//创建生产者pool
	n.pool, err = CreateProducerPool(config.Address, config.AuthSecret, config.ClientConf)
	if err != nil {
		return err
	}

	//创建formatter
	factory, has := formatter.GetFormatterFactory(config.Type)
	if !has {
		return errFormatterType
	}
	n.formatter, err = factory.Create(config.Formatter)
	if err != nil {
		return err
	}

	return nil
}

func (n *NsqOutput) Stop() error {
	n.lock.Lock()
	defer n.lock.Unlock()

	n.pool.Close()
	n.formatter = nil
	n.pool = nil

	return nil
}

func (n *NsqOutput) CheckSkill(skill string) bool {
	return false
}

func (n *NsqOutput) Output(entry eosc.IEntry) error {
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
