package messaging

import (
	"context"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log/slog"
)

type MessageHandler func(ctx context.Context, delivery amqp.Delivery)

type Consumer struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	h    MessageHandler
}

func NewConsumer(conn *amqp.Connection, h MessageHandler) (*Consumer, error) {
	ch, err := conn.Channel()
	if err != nil {
		slog.Error("Failed to open a channel", "error", err)
		return nil, err
	}

	err = ch.Qos(1, 0, false)
	if err != nil {
		slog.Error("Failed to set QoS", "error", err)
		return nil, err
	}

	slog.Info("Consumer created successfully")
	return &Consumer{conn: conn, ch: ch, h: h}, nil
}

func (c *Consumer) Close() error {
	if c.ch != nil {
		return c.ch.Close()
	}

	return nil
}

func (c *Consumer) Start(ctx context.Context, queueName string, consumerTag string) error {
	msgs, err := c.ch.Consume(
		queueName,
		consumerTag,
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		slog.Error("Failed to register a consumer", "error", err)
		return err
	}

	for {
		select {
		case <-ctx.Done():
			slog.Info("Consumer shutting down")
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				slog.Error("Failed to consume message", "error", err)
				return fmt.Errorf("failed to consume message: %w", err)
			}

			c.h(ctx, msg)
		}
	}
}
