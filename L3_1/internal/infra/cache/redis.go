package cache

import (
	"log"
	"os"
	"strconv"
)

type RedisConfig struct {
	Addr string
	Pass string
	DB   int
}

func NewRedisConfig() *RedisConfig {
	addr := os.Getenv("REDIS_ADDR")
	pass := os.Getenv("REDIS_PASS")
	db, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		log.Fatal("Error parsing REDIS_DB")
	}

	return &RedisConfig{
		Addr: addr,
		Pass: pass,
		DB:   db,
	}
}
