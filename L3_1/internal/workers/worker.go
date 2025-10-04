package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/1lostsun/L3/internal/entity"
	r "github.com/go-redis/redis/v8"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/wb-go/wbf/rabbitmq"
	"github.com/wb-go/wbf/redis"
	"log"
	"math"
	"os"
	"strconv"
	"time"
)

type Worker struct {
	cons                *rabbitmq.Consumer
	timeUp              time.Duration
	redis               *redis.Client
	bot                 *tgbotapi.BotAPI
	chatID              int64
	scheduleCheckSignal chan struct{}
}

func New(cons *rabbitmq.Consumer, redis *redis.Client) *Worker {
	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_TOKEN environment variable not set")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}

	chatIDStr := os.Getenv("TELEGRAM_CHAT_ID")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil || chatID == 0 {
		log.Fatalf("Invalid or missing TELEGRAM_CHAT_ID: %v", err)
	}

	return &Worker{
		cons:                cons,
		timeUp:              time.Second * 30,
		redis:               redis,
		bot:                 bot,
		chatID:              chatID,
		scheduleCheckSignal: make(chan struct{}, 1),
	}
}

func (w *Worker) Start(ctx context.Context) error {
	go w.startConsuming(ctx)

	ticker := time.NewTicker(w.timeUp)
	defer ticker.Stop()

	if err := w.checkRedis(ctx); err != nil {
		log.Printf("Initial checkRedis failed: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := w.checkRedis(ctx); err != nil {
				log.Printf("error during checkRedis on ticker: %v", err)
			}
		case <-w.scheduleCheckSignal:
			if err := w.checkRedis(ctx); err != nil {
				log.Printf("error during checkRedis on signal: %v", err)
			}
		}
	}
}

func (w *Worker) checkRedis(ctx context.Context) error {
	now := time.Now().Unix()

	ids, err := w.redis.ZRangeByScore(ctx, "notify:scheduled", &r.ZRangeBy{
		Min: "-inf",
		Max: fmt.Sprintf("%d", now),
	}).Result()
	if err != nil {
		return fmt.Errorf("ZRangeByScore error: %w", err)
	}

	for _, id := range ids {
		intID, err := strconv.Atoi(id)
		if err != nil {
			log.Printf("invalid ID in ZSET: %v", id)
			continue
		}

		key := fmt.Sprintf("notify:%d", intID)

		data, err := w.redis.Get(ctx, key)
		if err != nil {
			fmt.Println("Redis key not found:", key)
			return fmt.Errorf("error getting data from redis: %w", err)
		}

		var n entity.Notification
		if err := json.Unmarshal([]byte(data), &n); err != nil {
			return fmt.Errorf("error unmarshalling data: %w", err)
		}

		if n.Status != entity.StatusCancelled {
			if err := w.sendMessage(n); err != nil {
				log.Printf("ERROR: Failed to send notification %d: %v", intID, err)
				if err := w.handleFailedSend(ctx, n); err != nil {
					log.Printf("ERROR: Failed to handle failed send notification %d: %v", intID, err)
					return fmt.Errorf("critical failure in retry logic: %w", err)
				}
				continue
			}

			n.Status = entity.StatusSent
			newData, err := json.Marshal(n)
			if err != nil {
				return fmt.Errorf("error marshalling data: %w", err)
			}

			if err := w.redis.SetWithExpiration(ctx, key, newData, time.Hour); err != nil {
				return fmt.Errorf("error setting data to redis: %w", err)
			}

		}

		cmd := w.redis.ZRem(ctx, "notify:scheduled", id)
		if err := cmd.Err(); err != nil {
			return fmt.Errorf("error deleting scheduled notification %d: %w", intID, err)
		}
	}

	return nil
}

func (w *Worker) sendMessage(n entity.Notification) error {
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return fmt.Errorf("failed to load location: %w", err)
	}

	localTime := n.Date.In(loc).Format("2006-01-02 15:04:05")

	text := fmt.Sprintf("📨 %s\n🕒Время напоминания по Москве: %s", n.Message, localTime)

	msg := tgbotapi.NewMessage(w.chatID, text)
	_, err = w.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("error sending message: %w", err)
	}

	return nil
}

func (w *Worker) startConsuming(ctx context.Context) {
	msgChan := make(chan []byte)

	go func() {
		if err := w.cons.Consume(msgChan); err != nil {
			log.Fatalf("Error on consume message: %v", err)
		}

		log.Println("RabbitMQ Consumer exited gracefully.")
	}()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping RabbitMQ message processing.")
			return
		case msg := <-msgChan:
			fmt.Println(string(msg))
			w.handlerRabbitMQMessage(ctx, msg)
		}
	}
}

func (w *Worker) handlerRabbitMQMessage(ctx context.Context, msg []byte) {
	var n entity.Notification
	if err := json.Unmarshal(msg, &n); err != nil {
		log.Printf("Error unmarshalling message: %v", err)
		return
	}

	switch n.Status {
	case entity.StatusScheduled:
		w.triggerScheduleCheck()
	case entity.StatusCancelled:
		w.handleCancellation(ctx, n)
	}
}

func (w *Worker) triggerScheduleCheck() {
	select {
	case w.scheduleCheckSignal <- struct{}{}:
		log.Println("Scheduled check signal")
	default:
	}
}

func (w *Worker) handleCancellation(ctx context.Context, n entity.Notification) {
	cmd := w.redis.ZRem(ctx, "notify:scheduled", strconv.FormatUint(n.ID, 10))
	if cmd.Err() != nil {
		log.Printf("Error on deleting scheduled message: %v", cmd.Err())
	} else {
		log.Printf("Removed scheduled message ID=%d", n.ID)
	}
}

func (w *Worker) handleFailedSend(ctx context.Context, n entity.Notification) error {
	n.Retries++
	if n.Retries >= entity.RetriesCount {
		n.Status = entity.StatusFailed
		log.Printf("NOTIFICATION FAILED: ID %d reached max attempts (%d). Status set to FAILED.", n.ID, entity.RetriesCount)

		if data, err := json.Marshal(n); err == nil {
			if err := w.redis.SetWithExpiration(ctx, fmt.Sprintf("notify:%d", n.ID), data, entity.TTL); err != nil {
				return fmt.Errorf("failed to set notification state for retry: %w", err)
			}
		}

		w.redis.ZRem(ctx, "notify:scheduled", n.ID)
		return nil
	}

	baseDelay := 10 * time.Second
	delay := time.Duration(float64(baseDelay) * math.Pow(2, float64(n.Retries-1)))
	newScore := time.Now().Add(delay).Unix()

	log.Printf("NOTIFICATION RETRY: ID %d failed. Rescheduling for %v (Attempt %d).", n.ID, delay, n.Retries)

	n.Status = entity.StatusScheduled

	data, err := json.Marshal(n)
	if err != nil {
		return fmt.Errorf("error marshalling data: %w", err)
	}

	if err := w.redis.SetWithExpiration(ctx, fmt.Sprintf("notify:%d", n.ID), data, entity.TTL); err != nil {
		return fmt.Errorf("failed to set notification state for retry: %w", err)
	}

	w.redis.ZRem(ctx, "notify:scheduled", n.ID)

	if w.redis.ZAdd(ctx, "notify:scheduled", &r.Z{
		Score:  float64(newScore),
		Member: n.ID,
	}).Err() != nil {
		return fmt.Errorf("failed to reschedule notification: %w", err)
	}

	return nil
}
