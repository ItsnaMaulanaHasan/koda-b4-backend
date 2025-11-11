package lib

import (
	"os"

	"github.com/redis/go-redis/v9"
)

func Redis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
		Protocol: 2,
	})
	return rdb
}
