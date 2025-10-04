package messaging

import "github.com/wb-go/wbf/rabbitmq"

func NewConsumer(ch *rabbitmq.Channel, queueName string) *rabbitmq.Consumer {
	conf := rabbitmq.NewConsumerConfig(queueName)
	conf.Consumer = "events-consumer"
	conf.AutoAck = false
	return rabbitmq.NewConsumer(ch, conf)
}
