package lib

import (
	"context"
	"crypto/tls"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

func Redis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:      os.Getenv("REDIS_CLIENT"),
		Password:  os.Getenv("REDIS_PASSWORD"),
		DB:        0,
		TLSConfig: &tls.Config{},
	})

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Redis connection failed: %v", err)
	} else {
		log.Println("Connected to Redis successfully!")
	}

	return rdb
}
