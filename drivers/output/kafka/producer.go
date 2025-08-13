package kafka

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/IBM/sarama"
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
	go o.work(o.producer, ctx, cfg)
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
	o.producer.Close()
	o.producer = nil
	o.formatter = nil
}

func (o *tProducer) output(entry eosc.IEntry) error {
	log.DebugF("kafka output begin...")
	if o.producer == nil && o.formatter == nil {
		log.DebugF("kafka producer and formatter is nil")
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
	log.DebugF("kafka send addr: %s, topic: %s, data: %s", o.conf.Address, o.conf.Topic, data)
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

func (o *tProducer) work(producer sarama.AsyncProducer, ctx context.Context, cfg *ProducerConfig) {
	for {
		select {
		case <-ctx.Done():
			// 读完
			for e := range producer.Errors() {
				log.Warnf("kafka error:%s", e.Error())
			}
			return
		case err := <-producer.Errors():
			log.DebugF("receive error.kafka addr: %s,kafka topic: %s,kafka partition: %d", cfg.Address, cfg.Topic, cfg.Partition)
			if err != nil {
				log.Errorf("kafka error:%s", err.Error())
			}
		case success, ok := <-producer.Successes():
			if !ok {
				return
			}
			log.DebugF("Message sent to partition %d at offset %d\n", success.Partition, success.Offset)
			//key, err := success.Key.Encode()
			//if err != nil {
			//	log.Errorf("kafka error:%s", err.Error())
			//	continue
			//}
			//value, err := success.Value.Encode()
			//if err != nil {
			//	log.Errorf("kafka error:%s", err.Error())
			//	continue
			//}
			//log.DebugF("kafka success addr: %s, topic: %s, partition: %d, key: %s, value: %s", cfg.Address, cfg.Topic, success.Partition, string(key), string(value))
		}
	}
}
