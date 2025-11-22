package config

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var Rdb *redis.Client

func InitRedis() {
	options, err := redis.ParseURL(os.Getenv("REDIS_URL"))
	if err != nil {
		log.Fatalf("Invalid redis URL: %v", err)
	}

	Rdb = redis.NewClient(options)

	if err := Rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Redis connection failed: %v", err)
	} else {
		log.Println("Connected to Redis successfully!")
	}
}
