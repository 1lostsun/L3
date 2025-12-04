package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/1lostsun/L3/internal/entity/messages"
	"github.com/1lostsun/L3/internal/infra/messaging"
	"github.com/1lostsun/L3/internal/repo/pg"
	amqp "github.com/rabbitmq/amqp091-go"
	"log/slog"
)

type BookingExpiryWorker struct {
	r *pg.Repo
}

func NewBookingExpiryWorker(r *pg.Repo) *BookingExpiryWorker {
	return &BookingExpiryWorker{r: r}
}

func (w *BookingExpiryWorker) HandleMessage(ctx context.Context, delivery amqp.Delivery) {
	var message messages.BookingExpiryMessage

	err := json.Unmarshal(delivery.Body, &message)
	if err != nil {
		slog.Error("failed to unmarshal booking expiry message")
		return
	}

	wasCancelled, err := w.r.CancelBooking(ctx, message.BookingID)
	if err != nil {
		slog.Error(fmt.Sprintf("failed to cancel booking id: %s; \nerr:  %w", message.BookingID, err))
		delivery.Nack(false, true)
		return
	}

	if wasCancelled {
		slog.Info(
			fmt.Sprintf(
				"booking %s was EXPIRED and CANCELED; \nfreed place %s",
				message.BookingID, message.PlaceID),
		)
	} else {
		slog.Info(
			fmt.Sprintf(
				"booking %s was skipped; \nplace %s already PAID or CANCELLED",
				message.BookingID, message.PlaceID,
			),
		)
	}

	delivery.Ack(false)
}

func (w *BookingExpiryWorker) Start(ctx context.Context, conn *amqp.Connection) error {
	consumer, err := messaging.NewConsumer(conn, w.HandleMessage)
	if err != nil {
		return err
	}

	defer consumer.Close()

	return consumer.Start(
		ctx,
		messaging.BookingExpiredQueue,
		"booking-expiry-worker",
	)
}
