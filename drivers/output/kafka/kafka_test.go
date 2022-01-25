package kafka

import (
	"fmt"
	"github.com/Shopify/sarama"
	"testing"
	"time"
)

var (
	addr = []string{
		"alikafka-post-cn-7mz2jfjap00k-1-vpc.alikafka.aliyuncs.com:9092",
		"alikafka-post-cn-7mz2jfjap00k-2-vpc.alikafka.aliyuncs.com:9092",
		"alikafka-post-cn-7mz2jfjap00k-3-vpc.alikafka.aliyuncs.com:9092"}
)

func beginConsumer(topic string, addr []string, partition int32) {
	config := sarama.NewConfig()
	config.Version = sarama.V0_11_0_2
	config.Net.DialTimeout = 3 * time.Second
	config.Consumer.Return.Errors = true
	consumer, err := sarama.NewConsumer(addr, config)
	if err != nil {
		fmt.Println("create consumer error", err)
		return
	}
	defer consumer.Close()
	partitionConsumer, err := consumer.ConsumePartition(topic, partition, sarama.OffsetOldest)
	if err != nil {
		fmt.Println("error get partition consumer", err)
		return
	}
	defer partitionConsumer.Close()
	fmt.Println("consumer work!")
	for {
		select {

		case msg := <-partitionConsumer.Messages():
			if msg != nil {
				fmt.Println("msg offset: ", msg.Offset, " partition: ", msg.Partition, " times: ", msg.Timestamp.Format("2006-Jan-02 15:04"), " value: ", string(msg.Value))
			}
		case err = <-partitionConsumer.Errors():
			if err != nil {
				fmt.Println(err.Error())
			}
		}
	}
}

type TestProducerConfig struct {
	address   []string
	topic     string
	content   string
	partition int32
}

func TestSendMessageSync(t *testing.T) {
	msg := TestProducerConfig{
		address:   addr,
		topic:     "test",
		content:   "ApintoSync",
		partition: 0,
	}
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Timeout = 3 * time.Second
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Partitioner = sarama.NewManualPartitioner
	config.Version = sarama.V0_11_0_2
	p, err := sarama.NewSyncProducer(msg.address, config)
	if err != nil {
		t.Errorf("sarama.NewSyncProducer err, message=%s \n", err)
	}
	defer p.Close()
	m := &sarama.ProducerMessage{
		Topic:     msg.topic,
		Value:     sarama.ByteEncoder(msg.content),
		Partition: msg.partition,
	}
	part, offset, err := p.SendMessage(m)
	if err != nil {
		t.Errorf("send message(%s) err=%v \n", msg.content, err)
	} else {
		t.Logf("send success, partition=%d, offset=%d \n", part, offset)
	}
}
func TestSendMessageAsync(t *testing.T) {
	msg := TestProducerConfig{
		address:   addr,
		topic:     "test",
		content:   "ApintoAsync",
		partition: 0,
	}
	config := sarama.NewConfig()
	// 等待服务器所有副本都保存成功后，再返回响应
	config.Producer.RequiredAcks = sarama.WaitForLocal
	// 随机向partition发送消息
	config.Producer.Partitioner = sarama.NewManualPartitioner
	// 是否等待成功和失败后的响应，只有上面的RequireAcks设置不是NoResponse，这里才有用。
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Version = sarama.V0_11_0_2
	producer, err := sarama.NewAsyncProducer(msg.address, config)
	if err != nil {
		t.Errorf("fail to make a producer, error=%v", err)
	}
	defer func() {
		// 关闭
		producer.AsyncClose()
		// 关闭的时候确保读完
		for e := range producer.Errors() {
			t.Errorf("error= %v\n", e.Err)
		}
		for e := range producer.Successes() {
			if e != nil {
				t.Errorf("succeed, offset=%d, timestamp=%s, partitions=%d\n", e.Offset, e.Timestamp.String(), e.Partition)
			}
		}
	}()

	producer.Input() <- &sarama.ProducerMessage{
		Topic:     msg.topic,
		Partition: msg.partition,
		Value:     sarama.ByteEncoder(msg.content),
	}
	select {
	case suc := <-producer.Successes():
		if suc != nil {
			t.Logf("succeed, offset=%d, timestamp=%s, partitions=%d\n", suc.Offset, suc.Timestamp.String(), suc.Partition)
		}
	case fail := <-producer.Errors():
		if fail != nil {
			t.Errorf("error= %v\n", fail.Err)
		}
	}
}

func TestMain(m *testing.M) {
	go beginConsumer("test", addr, 0)
	<-time.After(1 * time.Second)
	m.Run()
	<-time.After(60 * time.Second)
}
