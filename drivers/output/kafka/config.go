package kafka

import (
	"errors"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/eolinker/eosc"
)

var (
	errTopic           = errors.New("topic can not be null. ")
	errAddress         = errors.New("address is invalid. ")
	errorFormatterType = errors.New("error formatter type")
	errorPartitionKey  = errors.New("partition key is invalid")
)

type Config struct {
	Scopes        []string             `json:"scopes" label:"作用域"`
	Topic         string               `json:"topic" yaml:"topic" label:"Topic"`
	Address       string               `json:"address" yaml:"address" label:"请求地址"`
	Timeout       int                  `json:"timeout" yaml:"timeout" label:"超时时间"`
	Version       string               `json:"kafka_version" yaml:"kafka_version" label:"版本" default:"1.0.0.0" enum:"0.8.2.0, 0.8.2.1, 0.8.2.2, 0.9.0.0, 0.9.0.1, 0.10.0.0, 0.10.0.1, 0.10.1.0, 0.10.1.1, 0.10.2.0, 0.10.2.1, 0.10.2.2, 0.11.0.0, 0.11.0.1, 0.11.0.2, 1.0.0.0, 1.0.1.0, 1.0.2.0, 1.1.0.0, 1.1.1.0, 2.0.0.0, 2.0.1.0, 2.1.0.0, 2.1.1.0, 2.2.0.0, 2.2.1.0, 2.2.2.0, 2.3.0.0, 2.3.1.0, 2.4.0.0, 2.4.1.0, 2.5.0.0, 2.5.1.0, 2.6.0.0, 2.6.1.0, 2.6.2.0, 2.7.0.0, 2.7.1.0, 2.8.0.0, 2.8.1.0, 3.0.0.0, 3.1.0.0"`
	PartitionType string               `json:"partition_type" yaml:"partition_type" enum:"robin,hash,manual,random"`
	Partition     int32                `json:"partition" yaml:"partition" switch:"partition_type==='manual'"`
	PartitionKey  string               `json:"partition_key" yaml:"partition_key" switch:"partition_type==='hash'"`
	Type          string               `json:"type" yaml:"type" enum:"json,line" label:"输出格式"`
	Formatter     eosc.FormatterConfig `json:"formatter" yaml:"formatter" label:"格式化配置"`
}

type ProducerConfig struct {
	Address       []string             `json:"address" yaml:"address"`
	Topic         string               `json:"topic" yaml:"topic"`
	Partition     int32                `json:"partition" yaml:"partition"`
	PartitionKey  string               `json:"partition_key" yaml:"partition_key"`
	PartitionType string               `json:"partition_type" yaml:"partition_type"`
	Conf          *sarama.Config       `json:"conf" yaml:"conf"`
	Type          string               `json:"type" yaml:"type"`
	Formatter     eosc.FormatterConfig `json:"formatter" yaml:"formatter"`
}

func (c *Config) doCheck() (*ProducerConfig, error) {
	conf := c
	if conf.Topic == "" {
		return nil, errTopic
	}
	if conf.Address == "" {
		return nil, errAddress
	}
	p := &ProducerConfig{}
	p.Topic = conf.Topic
	s := sarama.NewConfig()
	if conf.Version != "" {
		v, err := sarama.ParseKafkaVersion(conf.Version)
		if err != nil {
			return nil, err
		}
		s.Version = v
	}
	p.PartitionType = conf.PartitionType
	switch conf.PartitionType {
	case "robin":
		s.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	case "hash":
		// 通过hash获取
		key := strings.TrimLeft(conf.PartitionKey, "$")
		if key == "" {
			// key为空则还是用随机
			s.Producer.Partitioner = sarama.NewRandomPartitioner
			p.PartitionType = "random"
		} else {
			if !strings.HasPrefix(conf.PartitionKey, "$") {
				return nil, errorPartitionKey
			}
			s.Producer.Partitioner = sarama.NewHashPartitioner
			p.PartitionKey = key
		}
	case "manual":
		// 手动指定分区
		s.Producer.Partitioner = sarama.NewManualPartitioner
		// 默认为0
		p.Partition = conf.Partition
	default:
		s.Producer.Partitioner = sarama.NewRandomPartitioner
		p.PartitionType = "random"
	}
	// 只监听错误
	s.Producer.Return.Errors = true
	s.Producer.Return.Successes = false
	s.Producer.RequiredAcks = sarama.WaitForLocal

	p.Address = strings.Split(conf.Address, ",")
	if len(p.Address) == 0 {
		return nil, errAddress
	}
	// 超时时间
	if conf.Timeout != 0 {
		s.Producer.Timeout = time.Duration(conf.Timeout) * time.Second
	}

	if conf.Type == "" {
		conf.Type = "line"
	}
	p.Type = conf.Type
	p.Formatter = conf.Formatter
	p.Conf = s
	return p, nil
}
