package kafka

import (
	"errors"
	"github.com/Shopify/sarama"
	"github.com/eolinker/eosc"
	"strings"
	"time"
)

var (
	errTopic   = errors.New("topic can not be null. ")
	errAddress = errors.New("address can not be null. ")
	errVersion = errors.New("version format is invalid. ")
)

type Config struct {
	Config *Kafka `json:"config" yaml:"config"`
}
type Kafka struct {
	Topic         string               `json:"topic" yaml:"topic"`
	Address       string               `json:"address" yaml:"address"`
	Timeout       int                  `json:"timeout" yaml:"timeout"`
	Version       string               `json:"version" yaml:"version"`
	PartitionType string               `json:"partition_type" yaml:"partition_type"`
	Partition     int                  `json:"partition" yaml:"partition"`
	PartitionKey  string               `json:"partition_key" yaml:"partition_key"`
	Type          string               `json:"type" yaml:"type"`
	Formatter     eosc.FormatterConfig `json:"formatter" yaml:"formatter"`
}

type ProducerConfig struct {
	Address      []string             `json:"address" yaml:"address"`
	Topic        string               `json:"topic" yaml:"topic"`
	Partition    int                  `json:"partition" yaml:"partition"`
	PartitionKey string               `json:"partition_key" yaml:"partition_key"`
	Conf         *sarama.Config       `json:"conf" yaml:"conf"`
	Type         string               `json:"type" yaml:"type"`
	Formatter    eosc.FormatterConfig `json:"formatter" yaml:"formatter"`
}

func (c *Config) doCheck() (*ProducerConfig, error) {
	conf := c.Config
	if conf.Topic == "" {
		return nil, errTopic
	}
	if conf.Address == "" {
		return nil, errAddress
	}
	p := &ProducerConfig{}
	s := sarama.NewConfig()
	if conf.Version != "" {
		v, err := sarama.ParseKafkaVersion(conf.Version)
		if err != nil {
			return nil, err
		}
		s.Version = v
	}
	switch c.Config.PartitionType {
	case "robin":
		s.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	case "hash":
		s.Producer.Partitioner = sarama.NewHashPartitioner
		p.PartitionKey = conf.PartitionKey
	case "manual":
		s.Producer.Partitioner = sarama.NewManualPartitioner
		p.Partition = conf.Partition
	default:
		s.Producer.Partitioner = sarama.NewRandomPartitioner
	}
	s.Producer.Return.Errors = true
	s.Producer.RequiredAcks = sarama.WaitForLocal

	p.Address = strings.Split(conf.Address, ",")
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
