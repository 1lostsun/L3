package messaging

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log/slog"
)

func SetupRabbitMQ(conn *amqp.Connection) error {
	ch, err := conn.Channel()
	if err != nil {
		slog.Error("failed to create channel")
		return fmt.Errorf("failed to create channel: %w", err)
	}

	defer ch.Close()

	err = ch.ExchangeDeclare(
		BookingExpiredExchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		slog.Error("failed to declare expired exchange")
		return fmt.Errorf("failed to declare expired exchange: %w", err)
	}

	_, err = ch.QueueDeclare(
		BookingExpiredQueue,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		slog.Error("failed to declare expired queue")
		return fmt.Errorf("failed to declare expired queue: %w", err)
	}

	err = ch.QueueBind(
		BookingExpiredQueue,
		BookingDelayedRoutingKey,
		BookingExpiredExchange,
		false,
		nil,
	)

	if err != nil {
		slog.Error("failed to bind expired queue")
		return fmt.Errorf("failed to bind expired queue: %w", err)
	}

	err = ch.ExchangeDeclare(
		BookingDelayedExchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		slog.Error("failed to declare delayed exchange: %w", err)
		return fmt.Errorf("failed to declare delayed exchange: %w", err)
	}

	_, err = ch.QueueDeclare(
		BookingDelayedQueue,
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-dead-letter-exchange":    BookingExpiredExchange,
			"x-dead-letter-routing-key": BookingExpiredRoutingKey,
		},
	)

	if err != nil {
		slog.Error("failed to declare delayed queue")
		return fmt.Errorf("failed to declare delayed queue: %w", err)
	}

	err = ch.QueueBind(
		BookingDelayedQueue,
		BookingDelayedRoutingKey,
		BookingDelayedExchange,
		false,
		nil,
	)

	if err != nil {
		slog.Error("failed to bind delayed queue")
		return fmt.Errorf("failed to bind delayed queue: %w", err)
	}

	slog.Info("RabbitMQ setup completed successfully")
	return nil
}
