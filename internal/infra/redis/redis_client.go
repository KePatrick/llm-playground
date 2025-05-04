package redis

import (
	"context"
	"fmt"
	"kepatrick/llm-playground/internal/config"

	"github.com/redis/go-redis/v9"
)

func InitRedisClient(conf config.Redis) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     conf.Addr,
		Password: conf.Password,
		DB:       conf.Db,
	})

	// connection test
	Ctx := context.Background()
	_, err := client.Ping(Ctx).Result()
	if err != nil {
		panic(fmt.Sprintf("Failed to connect Redis: %v", err))
	}

	return client
}
