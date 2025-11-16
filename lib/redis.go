package lib

import (
	"context"
	"crypto/tls"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

func Redis() *redis.Client {
	options, err := redis.ParseURL(os.Getenv("REDIS_URL"))
	if err != nil {
		log.Fatalf("Invalid redis URL: %v", err)
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
