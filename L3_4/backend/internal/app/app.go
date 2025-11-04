package app

import (
	"backend/internal/handler"
	"backend/internal/infra/cache/redis"
	"backend/internal/infra/queue/kafka/consumer"
	"backend/internal/infra/queue/kafka/producer"
	"backend/internal/storage/MinIO"
	"backend/internal/usecase"
	"backend/internal/workers"
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	port := os.Getenv("PORT")

	minioStorage, err := MinIO.New()
	if err != nil {
		log.Fatal("Error loading MinIO storage")
	}

	redisCache := redis.New()
	kafkaProducer := producer.New()
	kafkaConsumer := consumer.New()

	worker := workers.New(kafkaConsumer, minioStorage, redisCache)

	go func(ctx context.Context) {
		if err := worker.Run(ctx); err != nil {
			errCh <- err
		}
	}(ctx)

	uc := usecase.New(minioStorage, redisCache, kafkaProducer.Producer)
	h := handler.New(uc)
	h.InitRoutes()

	go func() {
		if err := h.Engine.Run(port); err != nil {
			errCh <- fmt.Errorf("error running server: %w", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		log.Printf("error: %v", err)
	case sig := <-sigCh:
		log.Printf("received signal: %v", sig)
	case <-ctx.Done():
		log.Println("context cancelled")
	}

	cancel()
	_ = kafkaConsumer.Close()
	_ = kafkaProducer.Close()
}
