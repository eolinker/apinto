package nsq

import (
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/formatter"
	"github.com/nsqio/go-nsq"
)

type NsqOutput struct {
	*Driver
	id        string
	config    *NsqConf
	producer  *nsq.Producer
	formatter eosc.IFormatter
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

	if n.config == nil || n.config.isProducerUpdate(config) {
		if n.producer != nil {
			n.producer.Stop()
		}
		//创建生产者
		nsqConf := nsq.NewConfig()
		if config.AuthSecret != "" {
			nsqConf.AuthSecret = config.AuthSecret
		}
		n.producer, err = nsq.NewProducer(config.Address, nsqConf)
		if err != nil {
			return err
		}
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

	n.config = config
	return nil
}

func (n *NsqOutput) Stop() error {
	n.producer.Stop()
	n.producer = nil
	n.formatter = nil
	n.config = nil
	return nil
}

func (n *NsqOutput) CheckSkill(skill string) bool {
	return false
}

func (n *NsqOutput) Output(entry eosc.IEntry) error {
	if n.formatter != nil {
		data := n.formatter.Format(entry)
		if n.producer != nil && len(data) > 0 {
			err := n.producer.Publish(n.config.Topic, data)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
