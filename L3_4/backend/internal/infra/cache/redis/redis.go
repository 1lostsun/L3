package redis

import (
	"github.com/wb-go/wbf/redis"
	"log"
	"os"
	"strconv"
)

type Cache struct {
	Client *redis.Client
}

func New() *Cache {
	addr := os.Getenv("REDIS_ADDR")
	pass := os.Getenv("REDIS_PASS")
	db, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		log.Fatal("Error parsing REDIS_DB")
	}

	client := redis.New(addr, pass, db)
	return &Cache{Client: client}
}
