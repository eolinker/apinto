package nsq

import (
	"context"
	"time"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/formatter"
)

const (
	maxBufSize = 32
)

type Writer struct {
	//pool       *producerPool
	topic      string
	formatter  eosc.IFormatter
	ctx        context.Context
	cancelFunc context.CancelFunc
	//lock       sync.Mutex
	//multiBody     multiBody
	//multiBodySize int64
	//multiBodies   []multiBody
	bodyChan chan []byte
	poolChan chan *producerPool
}

func NewWriter() *Writer {
	ctx, cancel := context.WithCancel(context.Background())
	w := &Writer{
		//multiBody:  make([][]byte, 0),
		bodyChan:   make(chan []byte, maxBufSize),
		poolChan:   make(chan *producerPool, 1),
		ctx:        ctx,
		cancelFunc: cancel,
	}
	go w.doLoop()
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
	//n.lock.Lock()
	//defer n.lock.Unlock()

	//op := n.pool
	//n.pool = pool
	n.poolChan <- pool
	n.topic = config.Topic
	n.formatter = fm
	//if op != nil {
	//	op.Close()
	//}
	return nil
}
func (n *Writer) stop() error {
	//n.lock.Lock()
	//defer n.lock.Unlock()

	n.formatter = nil
	n.cancelFunc()
	close(n.bodyChan)
	close(n.poolChan)
	return nil
}

func (n *Writer) output(entry eosc.IEntry) {

	fm := n.formatter
	if fm == nil {
		return
	}
	data := fm.Format(entry)
	n.bodyChan <- data
	//n.lock.Lock()
	//n.multiBody = append(n.multiBody, data)
	//n.multiBodySize += int64(len(data))
	//if n.multiBodySize > maxBodySize {
	//	n.multiBodies = append(n.multiBodies, n.multiBody)
	//	n.multiBodySize = 0
	//	n.multiBody = make(multiBody, 0)
	//}
	//n.lock.Unlock()
}

func (n *Writer) doLoop() {
	buf := make([][]byte, 0, maxBufSize)
	timer := time.NewTimer(500 * time.Millisecond)
	defer timer.Stop()
	var pool *producerPool
	defer func() {
		if pool != nil {
			pool.Close()
		}
	}()
	for {
		select {
		case <-timer.C:
			if len(buf) < 1 {
				continue
			}
			if pool == nil {
				timer.Reset(500 * time.Millisecond)
				continue
			}

			tmp := buf
			buf = buf[:0]
			err := pool.Publish(n.topic, tmp)
			if err != nil {
				log.Error("nsq publish error: ", err.Error())
			}

		case body, ok := <-n.bodyChan:
			if !ok {
				return
			}

			buf = append(buf, body)
			if pool == nil {
				timer.Reset(500 * time.Millisecond)
				continue
			}
			if len(buf) >= maxBufSize {
				tmp := buf
				buf = buf[:0]

				err := pool.Publish(n.topic, tmp)
				if err != nil {
					log.Error("nsq publish error: ", err.Error())
				}
				// 触发批量发送逻辑
				continue
			}
			timer.Reset(500 * time.Millisecond)
		case p, ok := <-n.poolChan:
			if !ok {
				return
			}
			if pool != nil {
				pool.Close()
			}

			pool = p
		case <-n.ctx.Done():
			if len(buf) > 0 {
				if pool == nil {
					log.Error("data not send, pool is nil,data length ", len(buf))
					return
				}
				err := pool.Publish(n.topic, buf)
				if err != nil {
					log.Error("nsq publish error: ", err.Error())
				}
			}

			return
		}
	}
}
