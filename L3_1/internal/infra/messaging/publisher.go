package messaging

import "github.com/wb-go/wbf/rabbitmq"

func newPub(ch *rabbitmq.Channel, exchange *rabbitmq.Exchange) *rabbitmq.Publisher {
	return rabbitmq.NewPublisher(ch, exchange.Name())
}
