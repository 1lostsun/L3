package producer

import (
	"github.com/wb-go/wbf/kafka"
	"os"
)

type Producer struct {
	*kafka.Producer
}

func New() *Producer {
	port := os.Getenv("KAFKA_PORT")
	topic := os.Getenv("KAFKA_TOPIC")

	producer := kafka.NewProducer([]string{port}, topic)
	return &Producer{producer}
}
