package kafka

import (
	"fmt"
	"github.com/Shopify/sarama"
	"log"
	"testing"
	"time"
)

type SendType string

const (
	SyncSend  SendType = "sync"
	AsyncSend SendType = "async"
)

var (
	addr     = []string{"127.0.0.1:9092"}
	sendType SendType
)

type TestConfig struct {
	address   []string
	topic     string
	partition int32
}

// 先确保已开启生产者再执行该测试
func TestConsumer(t *testing.T) {
	cases := []struct {
		name string
		conf TestConfig
	}{
		{
			name: "baseTest",
			conf: TestConfig{
				topic:     "test",
				partition: 0,
				address:   addr,
			},
		},
	}
	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T) {
			config := sarama.NewConfig()
			config.Consumer.Return.Errors = true
			config.Version = sarama.V0_11_0_0
			config.Net.DialTimeout = 3 * time.Second
			consumer, err := sarama.NewConsumer(cc.conf.address, config)
			if err != nil {
				t.Fatal(err.Error())
			}
			defer consumer.Close()
			//根据消费者获取指定的主题分区的消费者,Offset这里指定为获取最新的消息.
			partitionConsumer, err := consumer.ConsumePartition(cc.conf.topic, cc.conf.partition, sarama.OffsetNewest)
			if err != nil {
				fmt.Println("error get partition consumer", err)
			}
			defer partitionConsumer.Close()
			//循环等待接受消息
			for {
				select {
				case <-time.After(5 * time.Second):
					t.Log("finish!")
					return
				case msg := <-partitionConsumer.Messages():
					fmt.Println("msg offset: ", msg.Offset, " partition: ", msg.Partition, " times: ", msg.Timestamp.Format("2006-Jan-02 15:04"), " value: ", string(msg.Value))
				case err := <-partitionConsumer.Errors():
					fmt.Println(err.Error())
				}
			}
		})
	}
}

type TestProducerConfig struct {
	address   []string
	topic     string
	content   string
	partition int32
}

func sendMessageSync(msg TestProducerConfig) error {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Timeout = 3 * time.Second
	config.Producer.Partitioner = sarama.NewManualPartitioner
	p, err := sarama.NewSyncProducer(msg.address, config)
	if err != nil {
		log.Printf("sarama.NewSyncProducer err, message=%s \n", err)
		return err
	}
	defer p.Close()
	part, offset, err := p.SendMessage(&sarama.ProducerMessage{
		Topic:     msg.topic,
		Value:     sarama.ByteEncoder(msg.content),
		Partition: msg.partition,
	})
	if err != nil {
		log.Printf("send message(%s) err=%v \n", msg.content, err)
		return err
	} else {
		log.Printf("send success, partition=%d, offset=%d \n", part, offset)
	}
	return nil
}
func sendMessageAsync(msg TestProducerConfig) error {
	config := sarama.NewConfig()
	// 等待服务器所有副本都保存成功后，再返回响应
	config.Producer.RequiredAcks = sarama.WaitForLocal
	// 随机向partition发送消息
	config.Producer.Partitioner = sarama.NewManualPartitioner
	// 是否等待成功和失败后的响应，只有上面的RequireAcks设置不是NoResponse，这里才有用。
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true

	// 使用配置，新建一个异步生产者
	producer, err := sarama.NewAsyncProducer(msg.address, config)
	if err != nil {
		fmt.Println("fail to make a producer, error=", err)
		return err
	}
	defer func() {
		// 关闭
		producer.AsyncClose()
		// 关闭的时候确保读完
		for e := range producer.Errors() {
			fmt.Printf("error= %v\n", e.Err)
		}
		for e := range producer.Successes() {
			if e != nil {
				fmt.Printf("succeed, offset=%d, timestamp=%s, partitions=%d\n", e.Offset, e.Timestamp.String(), e.Partition)
			}
		}
	}()
	go func(p sarama.AsyncProducer) {
		for {
			select {
			case <-time.After(5 * time.Second):
				// 监听5秒
				return
			case suc := <-p.Successes():
				if suc != nil {
					fmt.Printf("succeed, offset=%d, timestamp=%s, partitions=%d\n", suc.Offset, suc.Timestamp.String(), suc.Partition)
				}
			case fail := <-p.Errors():
				if fail != nil {
					fmt.Printf("error= %v\n", fail.Err)
				}
			}
		}
	}(producer)

	for i := 0; i < 5; i++ {
		producer.Input() <- &sarama.ProducerMessage{
			Topic:     msg.topic,
			Partition: msg.partition,
			Value:     sarama.ByteEncoder(msg.content),
		}
	}
	return nil
}

func TestMain(m *testing.M) {
	// 同步
	sendType = SyncSend
	// 异步
	//sendType = AsyncSend
	msg := TestProducerConfig{
		address:   addr,
		topic:     "test",
		content:   "test2022",
		partition: 0,
	}
	if sendType == SyncSend {
		err := sendMessageSync(msg)
		if err != nil {
			return
		}
	} else {
		err := sendMessageAsync(msg)
		if err != nil {
			return
		}
	}

	m.Run()
}
