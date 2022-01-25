package nsq

import (
	"fmt"
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

func (p *producerPool) PublishAsync(topic string, body []byte, doneChan chan *nsq.ProducerTransaction) error {
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
		ch := make(chan *nsq.ProducerTransaction, 1)
		defer close(ch)

		for attempt := 0; attempt < p.size; attempt++ {
			//若所有节点都不可用
			if int(p.downNodeNum) >= p.size {
				args := []interface{}{fmt.Sprintf("%s topic:%s data:%s", errNoValidProducer, topic, body)}
				doneChan <- &nsq.ProducerTransaction{Error: errNoValidProducer, Args: args}
				break
			}

			isLastAttempt := attempt+1 == p.size
			//轮询
			index := (n + attempt - 1) % p.size
			producerNode := p.nodes[index]

			//若该节点不可用
			if producerNode.status == disconnected {
				if isLastAttempt {
					args := []interface{}{fmt.Sprintf("nsqd_addr:%s topic:%s data:%s", producerNode.producer.String(), topic, body)}
					doneChan <- &nsq.ProducerTransaction{Error: errProducerInvalid, Args: args}
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
				}
				p.lock.Unlock()
			}

			//发送异步消息
			arg := fmt.Sprintf("nsqd_addr:%s topic:%s data:%s", producerNode.producer.String(), topic, body)
			if err := producerNode.producer.PublishAsync(topic, body, ch, arg); err != nil {
				if isLastAttempt {
					args := []interface{}{arg}
					doneChan <- &nsq.ProducerTransaction{Error: err, Args: args}
					break
				}

				continue
			}

			transaction := <-ch
			if transaction.Error != nil && !isLastAttempt {
				continue
			}
			doneChan <- transaction
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
