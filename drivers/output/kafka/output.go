package kafka

import (
	"context"
	"github.com/Shopify/sarama"
	"github.com/eolinker/eosc"
)

type Output struct {
	*Driver
	id        string
	input     *sarama.ProducerMessage
	producer  sarama.AsyncProducer
	formatter eosc.IFormatter
	cancel    context.CancelFunc
}

func (o *Output) Output(entry eosc.IEntry) error {
	//TODO implement me
	panic("implement me")
}

func (o *Output) Id() string {
	return o.id
}

func (o *Output) Start() error {
	return nil
}

func (o *Output) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (o *Output) Stop() error {
	o.producer.Close()
	o.formatter = nil
	return nil
}

func (o *Output) CheckSkill(skill string) bool {
	return false
}

// CreateProducer 创建kafka的生产者，采用同步生产者的方式
func (o *Output) CreateProducer() {
	// 发送地址 broker address
	//sarama.NewAsyncProducer()
	//// 发送的超时时间
	//
	//// 发送的topic
	//
	//// 分区的选择方式，分四种：
	////sarama.NewManualPartitioner() 返回一个手动选择分区的分割器,也就是获取msg中指定的`partition`，partition
	////sarama.NewRandomPartitioner() 通过随机函数随机获取一个分区号
	////sarama.NewRoundRobinPartitioner() 环形选择,也就是在所有分区中循环选择一个
	////sarama.NewHashPartitioner() 通过msg中的key生成hash值,选择分区，key
	//conf := sarama.NewConfig()
	////等待服务器所有副本都保存成功后的响应
	//conf.Producer.RequiredAcks = sarama.WaitForLocal
	////随机的分区类型
	//conf.Producer.Partitioner = sarama.NewRandomPartitioner
	////是否等待成功和失败后的响应,只有上面的RequireAcks设置不是NoReponse这里才有用.
	////conf.Producer.Return.Successes = true
	//conf.Producer.Return.Errors = true
	//conf.Producer.Timeout =

}
