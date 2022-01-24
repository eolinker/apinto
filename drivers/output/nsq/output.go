package nsq

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/formatter"
	"github.com/eolinker/eosc/log"
	"github.com/nsqio/go-nsq"
	"runtime/debug"
	"sync"
)

type NsqOutput struct {
	*Driver
	id        string
	config    *NsqConf
	producer  *nsq.Producer
	formatter eosc.IFormatter

	ptChannel  chan *nsq.ProducerTransaction
	cancelFunc context.CancelFunc
	lock       sync.Mutex
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

	if n.config == nil || n.config.isProducerUpdate(config) {
		if n.producer != nil {
			n.producer.Stop()
		}

		if n.cancelFunc == nil {
			ctx, cancelFunc := context.WithCancel(context.Background())
			n.cancelFunc = cancelFunc
			go n.listenAsycInfomation(n.ptChannel, ctx)
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
	n.lock.Lock()
	defer n.lock.Unlock()

	n.producer.Stop()
	n.producer = nil
	n.formatter = nil
	n.config = nil

	if n.cancelFunc != nil {
		n.cancelFunc()
		n.cancelFunc = nil
	}
	return nil
}

func (n *NsqOutput) CheckSkill(skill string) bool {
	return false
}

func (n *NsqOutput) Output(entry eosc.IEntry) error {
	if n.formatter != nil {
		data := n.formatter.Format(entry)
		if n.producer != nil && len(data) > 0 {
			args := []interface{}{n.producer.String(), n.config.Topic, data}
			err := n.producer.PublishAsync(n.config.Topic, data, n.ptChannel, args)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (n *NsqOutput) listenAsycInfomation(ptChannel chan *nsq.ProducerTransaction, ctx context.Context) {
	defer func() {
		if v := recover(); v != nil {
			if err, ok := v.(error); ok {
				fmt.Println("[nsq] log error: ", err)
				debug.PrintStack()
			}
			go n.listenAsycInfomation(ptChannel, ctx)
		}
	}()

	for {
		select {
		case pt := <-ptChannel:
			if pt.Error != nil {
				data, _ := json.Marshal(pt.Args)
				log.Errorf("nsq log error:%s data:%s", pt.Error, data)
			}
		case <-ctx.Done():
			return
		}
	}
}
