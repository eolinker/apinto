package kafka

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/Shopify/sarama"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/formatter"
	"github.com/eolinker/eosc/log"
)

type Producer interface {
	reset(config *ProducerConfig) error
	output(entry eosc.IEntry) error
	close()
}
type tProducer struct {
	wg       *sync.WaitGroup
	input    chan<- *sarama.ProducerMessage
	producer sarama.AsyncProducer
	conf     *ProducerConfig
	cancel   context.CancelFunc
	//enable    bool
	formatter eosc.IFormatter
}

func (o *tProducer) reset(cfg *ProducerConfig) (err error) {

	// 新建formatter
	factory, has := formatter.GetFormatterFactory(cfg.Type)
	if !has {
		return errorFormatterType
	}

	if o.producer != nil {
		// 确保关闭
		o.close()
	}

	var extendCfg []byte
	if cfg.Type == "json" {
		extendCfg, _ = json.Marshal(cfg.ContentResize)
	}
	o.formatter, err = factory.Create(cfg.Formatter, extendCfg)

	// 新建生产者
	o.producer, err = sarama.NewAsyncProducer(cfg.Address, cfg.Conf)
	if err != nil {
		return err
	}
	o.conf = cfg
	o.input = o.producer.Input()
	ctx, cancel := context.WithCancel(context.Background())
	o.cancel = cancel
	go o.work(o.producer, ctx)
	return nil
}

func newTProducer(config *ProducerConfig) *tProducer {
	p := &tProducer{}
	p.reset(config)
	return p
}

func (o *tProducer) close() {
	if o.cancel != nil {
		o.cancel()
		o.cancel = nil
	}
	o.producer.AsyncClose()
	o.producer = nil
	o.formatter = nil
}

func (o *tProducer) output(entry eosc.IEntry) error {
	if o.producer == nil && o.formatter == nil {
		return nil
	}

	data := o.formatter.Format(entry)
	msg := &sarama.ProducerMessage{
		Topic: o.conf.Topic,
		Value: sarama.ByteEncoder(data),
	}
	if o.conf.PartitionType == "manual" {
		msg.Partition = o.conf.Partition
	}
	if o.conf.PartitionType == "hash" {
		msg.Key = sarama.StringEncoder(eosc.ReadStringFromEntry(entry, o.conf.PartitionKey))
	}
	o.write(msg)

	return nil
}

func (o *tProducer) write(msg *sarama.ProducerMessage) {
	// 未开启情况下不给写
	//if !o.enable {
	//	return
	//}
	if o.input != nil {
		o.input <- msg
	}

}

func (o *tProducer) work(producer sarama.AsyncProducer, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			// 读完
			for e := range producer.Errors() {
				log.Warnf("kafka error:%s", e.Error())
			}
			return
		case err := <-producer.Errors():
			if err != nil {
				log.Warnf("kafka error:%s", err.Error())
			}
		case success, ok := <-producer.Successes():
			if !ok {
				return
			}
			log.Debug("kafka success:%s", success)
		}
	}
}
