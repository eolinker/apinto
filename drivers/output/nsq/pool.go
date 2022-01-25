package nsq

import (
	"fmt"
	"github.com/eolinker/eosc/log"
	"github.com/nsqio/go-nsq"
	"sync"
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
	downNodeNum int32

	config     *nsq.Config
	updateTime time.Time
	isClose    bool
	lock       sync.Mutex
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

	pool.updateTime = time.Now()
	return pool, nil
}

func (p *producerPool) PublishAsync(topic string, body []byte) error {
	if p.isClose || int(p.downNodeNum) >= p.size {
		return errNoValidProducer
	}

	if time.Now().Sub(p.updateTime) > time.Second*30 && len(p.nodes) > 0 {
		// 当上次节点更新时间与当前时间间隔超过30s，则检查连接关闭的节点
		go p.Check()
	}

	//使用round-robin进行负载均衡
	n := int(atomic.AddUint32(&p.next, 1))

	go func(n int) {

		for attempt := 0; attempt < p.size; attempt++ {
			//若所有节点都不可用
			if int(p.downNodeNum) >= p.size {
				log.Errorf("err:%s data:%s", errNoValidProducer, fmt.Sprintf("topic:%s data:%s", topic, body))
				break
			}

			isLastAttempt := attempt+1 == p.size
			//轮询
			index := (n + attempt - 1) % p.size
			producerNode := p.nodes[index]

			//若该节点不可用
			if producerNode.status == disconnected {
				if isLastAttempt {
					log.Errorf("nsq log error:%s data:%s", errProducerInvalid, fmt.Sprintf("nsqd_addr:%s topic:%s data:%s", producerNode.producer.String(), topic, body))
					break
				}
				continue
			}

			//确定该节点的连接情况，若连接不通则将状态置为disconnected，并且等待重新连接
			if err := producerNode.producer.Ping(); err != nil {
				p.lock.Lock()
				if err := producerNode.producer.Ping(); err != nil {
					producerNode.status = disconnected
					atomic.AddInt32(&p.downNodeNum, 1)
					if isLastAttempt {
						log.Errorf("nsq log error:%s data:%s", errProducerInvalid, fmt.Sprintf("nsqd_addr:%s topic:%s data:%s", producerNode.producer.String(), topic, body))
						break
					}

					continue
				}
				p.lock.Unlock()
			}

			//发送消息
			if err := producerNode.producer.Publish(topic, body); err != nil {
				if isLastAttempt {
					log.Errorf("nsq log error:%s data:%s", err, fmt.Sprintf("nsqd_addr:%s topic:%s data:%s", producerNode.producer.String(), topic, body))
					break
				}

				continue
			}

			break
		}
	}(n)

	return nil
}

//Check 检查
func (p *producerPool) Check() {
	p.lock.Lock()
	defer p.lock.Unlock()
	for _, n := range p.nodes {
		if n.status == disconnected {
			//解决断线重连的问题
			oldProducer := n.producer
			n.producer, _ = nsq.NewProducer(oldProducer.String(), p.config)
			n.status = connecting
			atomic.AddInt32(&p.downNodeNum, -1)
			oldProducer.Stop()
		}
	}
}

func (p *producerPool) Close() {

	for _, n := range p.nodes {
		n.producer.Stop()
	}
	p.isClose = true
}
