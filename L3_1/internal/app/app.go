package app

import (
	"context"
	"github.com/1lostsun/L3/internal/handler"
	"github.com/1lostsun/L3/internal/infra/cache"
	"github.com/1lostsun/L3/internal/infra/messaging"
	"github.com/1lostsun/L3/internal/usecase"
	"github.com/1lostsun/L3/internal/workers"
	"github.com/joho/godotenv"
	"github.com/wb-go/wbf/redis"
	"log"
	"os"
)

func Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Конфиг Редиса
	rConfig := cache.NewRedisConfig()

	// Структуры Redis и RabbitMQ
	redisCache := redis.New(rConfig.Addr, rConfig.Pass, rConfig.DB)
	rabbitMQ := messaging.NewRabbit()

	// Консюмер и продюссер RabbitMQ
	cons := messaging.NewConsumer(rabbitMQ.Ch, "notifications")
	pub := rabbitMQ.Pub

	uc := usecase.New(redisCache, pub)
	h := handler.New(uc)
	worker := workers.New(cons, redisCache)

	go func() {
		err := worker.Start(ctx)
		if err != nil {
			log.Fatalf("Error starting worker:%v", err)
		}
	}()

	h.InitRoutes()
	if err := h.Engine.Run(os.Getenv("SERVER_PORT")); err != nil {
		log.Fatal(err)
	}

	<-ctx.Done()
}
