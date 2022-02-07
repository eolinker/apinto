package nsq

import (
	"context"
	"fmt"
	"github.com/eolinker/eosc/log"
	"github.com/nsqio/go-nsq"
	"sync/atomic"
	"time"
)

const (
	connecting = iota
	disconnected
)

type producerPool struct {
	nodes       []*node
	size        int
	next        uint32

	config     *nsq.Config
	isClose    bool
	cancelFunc context.CancelFunc
}

type node struct {
	producer *nsq.Producer
	status   int
}

//Create
func CreateProducerPool(addrs []string, conf map[string]interface{}) (*producerPool, error) {

	pool := &producerPool{
		nodes: make([]*node, len(addrs)),
		size:  len(addrs),
	}

	nsqConf := nsq.NewConfig()
	//配置nsq_Config
	for k, v := range conf {
		err := nsqConf.Set(k, v)
		if err != nil {
			return nil, err
		}
	}
	pool.config = nsqConf

	for i, addr := range addrs {

		producer, err := nsq.NewProducer(addr, nsqConf)
		if err != nil {
			return nil, err
		}
		pool.nodes[i] = &node{producer: producer, status: connecting}
	}

	go pool.Check()
	return pool, nil
}

func (p *producerPool) PublishAsync(topic string, body []byte) error {
	if p.isClose {
		return errNoValidProducer
	}

	//使用round-robin进行负载均衡
	n := int(atomic.AddUint32(&p.next, 1))

	go func(n int) {

		for attempt := 0; attempt < p.size; attempt++ {
			//轮询
			index := (n + attempt - 1) % p.size
			producerNode := p.nodes[index]
			//若该节点不可用
			if producerNode.status == disconnected {
				continue
			}

			//发送消息
			if err := producerNode.producer.Publish(topic, body); err != nil {
				//发送失败，将该节点状态置为disconnected，等待check重新连接
				producerNode.status = disconnected
				log.Errorf("log output nsqd is invalid. nsqd_addr:%s  error:%s", producerNode.producer.String(), err)
				continue
			}
			break
		}
		log.Errorf("no available nsqd node. data: %s",fmt.Sprintf("topic:%s data:%s", topic, body))
	}(n)

	return nil
}

//Check 检查节点状态
func (p *producerPool) Check() {

	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()
	ctx, cancelFunc :=context.WithCancel(context.Background())
	p.cancelFunc = cancelFunc
	for{

		select {
		case <- ticker.C:
			for _, n := range p.nodes {

				if err := n.producer.Ping(); err != nil {
					//解决断线重连的问题
					//n.producer.Stop()
					//if err = n.producer.Ping();err != nil{
					//	continue
					//}
					//n.status = connecting
					//continue

					oldProducer := n.producer
					newProducer,_ := nsq.NewProducer(oldProducer.String(), p.config)
					if err = newProducer.Ping(); err != nil {
						if n.status == connecting {
							n.status = disconnected
							log.Errorf("log output nsqd is invalid. nsqd_addr:%s  error:%s", oldProducer.String(), err)
						}
						newProducer.Stop()
						continue
					}
					n.producer = newProducer
					n.status = connecting
					oldProducer.Stop()
				}
				n.status = connecting

			}
		case <- ctx.Done():
			return
		}

	}

}

func (p *producerPool) Close() {

	for _, n := range p.nodes {
		n.producer.Stop()
	}
	if p.cancelFunc != nil{
		p.cancelFunc()
		p.cancelFunc = nil
	}
	p.isClose = true
}
