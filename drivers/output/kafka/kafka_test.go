package kafka

import (
	"fmt"
	"github.com/Shopify/sarama"
	"testing"
)

func TestConsumer(t *testing.T) {
	fmt.Println(sarama.ParseKafkaVersion("0.8.2.0"))
}

func TestProducer(t *testing.T) {

}

func TestClient(t *testing.T) {

}
