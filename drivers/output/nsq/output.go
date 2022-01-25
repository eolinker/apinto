package nsq

import (
	"context"
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
	pool      *producerPool
	topic     string
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

	if n.pool != nil {
		n.pool.Close()
	}

	if n.cancelFunc == nil {
		ctx, cancelFunc := context.WithCancel(context.Background())
		n.cancelFunc = cancelFunc
		go n.listenAsycInfomation(n.ptChannel, ctx)
	}
	n.topic = config.Topic
	//创建生产者pool
	n.pool, err = CreateProducerPool(config.Address, config.ClientConf)
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
	if n.cancelFunc != nil {
		n.cancelFunc()
		n.cancelFunc = nil
	}
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
			err := n.pool.PublishAsync(n.topic, data, n.ptChannel)
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
				log.Errorf("nsq log error:%s data:%s", pt.Error, pt.Args[0])
			}
		case <-ctx.Done():
			return
		}
	}
}
