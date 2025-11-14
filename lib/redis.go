package lib

import (
	"context"
	"crypto/tls"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

func Redis() *redis.Client {
	options := &redis.Options{
		Addr:     os.Getenv("REDIS_CLIENT"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	}

	if os.Getenv("ENVIRONMENT") != "development" {
		options.TLSConfig = &tls.Config{}
	}

	rdb := redis.NewClient(options)

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Redis connection failed: %v", err)
	} else {
		log.Println("Connected to Redis successfully!")
	}

	return rdb
}
