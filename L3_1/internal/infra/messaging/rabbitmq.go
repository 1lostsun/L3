package messaging

import (
	"github.com/wb-go/wbf/rabbitmq"
	"log"
	"os"
	"strconv"
	"time"
)

type Rabbit struct {
	Conn *rabbitmq.Connection
	Ch   *rabbitmq.Channel
	Pub  *rabbitmq.Publisher
}

func NewRabbit() *Rabbit {
	conn := newConn()
	ch, ex := newCh(conn)
	pub := newPub(ch, ex)

	return &Rabbit{
		Conn: conn,
		Ch:   ch,
		Pub:  pub,
	}
}

func newCh(conn *rabbitmq.Connection) (*rabbitmq.Channel, *rabbitmq.Exchange) {
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}

	exchange := rabbitmq.NewExchange("message-exchange", "direct")
	exchange.Durable = true

	if err := exchange.BindToChannel(ch); err != nil {
		log.Fatal(err)
	}

	qm := rabbitmq.NewQueueManager(ch)

	_, err = qm.DeclareQueue("notifications", rabbitmq.QueueConfig{Durable: true})
	if err != nil {
		log.Fatal(err)
	}

	if err := ch.QueueBind("notifications",
		"notifications",
		exchange.Name(),
		false,
		nil); err != nil {
		log.Fatal(err)
	}

	return ch, exchange
}

func newConn() *rabbitmq.Connection {
	rabbitUrl := os.Getenv("RABBIT_URL")

	rabbitRetries, err := strconv.Atoi(os.Getenv("RABBIT_RETRIES"))
	if err != nil {
		log.Fatal("Error parsing RABBIT_RETRIES environment variable")
	}

	rabbitPause := time.Second

	conn, err := rabbitmq.Connect(rabbitUrl, rabbitRetries, rabbitPause)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ")
	}

	return conn
}
