package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/1lostsun/L3/internal/entity"
	r "github.com/go-redis/redis/v8"
	"github.com/wb-go/wbf/rabbitmq"
	"github.com/wb-go/wbf/redis"
)

type UseCase struct {
	redis     *redis.Client
	publisher *rabbitmq.Publisher
}

func New(r *redis.Client, pub *rabbitmq.Publisher) *UseCase {
	return &UseCase{
		redis:     r,
		publisher: pub,
	}
}

func (uc *UseCase) CreateNotification(ctx context.Context, notification entity.Notification) error {
	notification.Status = entity.StatusScheduled

	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	if err := uc.redis.Set(ctx, fmt.Sprintf("notify:%d", notification.ID), data); err != nil {
		return fmt.Errorf("failed to set notification '%d': %w", notification.ID, err)
	}

	cmd := uc.redis.ZAdd(ctx, "notify:scheduled", &r.Z{
		Score:  float64(notification.Date.Unix()),
		Member: notification.ID,
	})

	if cmd.Err() != nil {
		return fmt.Errorf("failed to schedule notification '%d': %v", notification.ID, cmd.Err())
	}

	if err := uc.publisher.Publish(data, "notifications", "application/json"); err != nil {
		return fmt.Errorf("failed to publish notification '%d': %w", notification.ID, err)
	}

	return nil
}

func (uc *UseCase) GetNotification(ctx context.Context, id string) (entity.Notification, error) {
	var notification entity.Notification

	data, err := uc.redis.Get(ctx, fmt.Sprintf("notify:%s", id))
	if err != nil {
		return entity.Notification{}, fmt.Errorf("failed to get notification '%s': %w", id, err)
	}

	if err := json.Unmarshal([]byte(data), &notification); err != nil {
		return entity.Notification{}, fmt.Errorf("failed to unmarshal notification '%s': %w", id, err)
	}

	return notification, nil
}

func (uc *UseCase) CancelNotification(ctx context.Context, id string) error {
	var notification entity.Notification

	val, err := uc.redis.Get(ctx, fmt.Sprintf("notify:%s", id))
	if err != nil {
		return fmt.Errorf("failed to get notification '%s': %w", id, err)
	}

	if err := json.Unmarshal([]byte(val), &notification); err != nil {
		return fmt.Errorf("failed to unmarshal notification '%s': %w", id, err)
	}

	notification.Status = entity.StatusCancelled
	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification '%s': %w", id, err)
	}

	notification.Retries = 0
	return uc.publisher.Publish(data, "notifications", "application/json")
}
