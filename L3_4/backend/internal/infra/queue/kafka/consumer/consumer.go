package consumer

import (
	"fmt"
	"github.com/wb-go/wbf/kafka"
	"os"
)

type Consumer struct {
	*kafka.Consumer
}

func New() *Consumer {
	port := os.Getenv("KAFKA_PORT")
	topic := os.Getenv("KAFKA_TOPIC")
	groupID := os.Getenv("KAFKA_GROUP_ID")

	fmt.Println(port, topic, groupID)

	consumer := kafka.NewConsumer([]string{port}, topic, groupID)
	return &Consumer{Consumer: consumer}
}
