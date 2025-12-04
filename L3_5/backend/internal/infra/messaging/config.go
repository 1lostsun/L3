package messaging

import "fmt"

type RabbitMQConfig struct {
	URL   string
	Host  string
	Port  int
	User  string
	Pass  string
	Vhost string
}

func (cfg *RabbitMQConfig) GetURL() string {
	if cfg.URL != "" {
		return cfg.URL
	}

	return fmt.Sprintf("amqp://%s:%s@%s:%d/%s", cfg.User, cfg.Pass, cfg.Host, cfg.Port, cfg.Vhost)
}

const (
	BookingDelayedExchange = "booking.delayed"
	BookingDelayedQueue    = "booking.delayed.queue"

	BookingExpiredExchange = "booking.expired"
	BookingExpiredQueue    = "booking.expired.queue"

	BookingDelayedRoutingKey = "booking.delay"
	BookingExpiredRoutingKey = "booking.expire"
)
