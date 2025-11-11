package lib

import "github.com/redis/go-redis/v9"

func Redis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		Protocol: 2,
	})
	return rdb
}
