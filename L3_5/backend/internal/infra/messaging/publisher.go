package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/1lostsun/L3/internal/entity/messages"
	amqp "github.com/rabbitmq/amqp091-go"
	"log/slog"
	"time"
)

type Publisher struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

// NewPublisher функция-конструктор для отправителя сообщений
func NewPublisher(conn *amqp.Connection) (*Publisher, error) {
	ch, err := conn.Channel()
	if err != nil {
		slog.Error("Failed to open channel", "error", err)
		return nil, err
	}

	if err := ch.Confirm(false); err != nil {
		slog.Error("Failed to write channel confirmation", "error", err)
		ch.Close()
		return nil, err
	}

	slog.Info("publisher created successfully")
	return &Publisher{conn, ch}, nil
}

func (p *Publisher) Close() error {
	if p.ch != nil {
		return p.ch.Close()
	}
	return nil
}

func (p *Publisher) Publish(ctx context.Context, bookingID string, eventID string, placeID string, ttl time.Duration) error {
	msg := messages.BookingExpiryMessage{
		BookingID:  bookingID,
		EventID:    eventID,
		PlaceID:    placeID,
		CreatedAt:  time.Now(),
		TTLMinutes: int(ttl.Minutes()),
	}

	body, err := json.Marshal(msg)
	if err != nil {
		slog.Error("Failed to marshal message", "error", err)
		return err
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = p.ch.PublishWithContext(
		ctxTimeout,
		BookingDelayedExchange,
		BookingDelayedRoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Expiration:   fmt.Sprintf("%d", ttl.Milliseconds()),
			Timestamp:    time.Now(),
			MessageId:    bookingID,
		},
	)

	if err != nil {
		slog.Error("Failed to publish message", "error", err)
		return err
	}

	slog.Info("published message successfully")
	return nil
}
