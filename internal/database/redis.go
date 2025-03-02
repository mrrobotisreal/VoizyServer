package database

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func InitRedis() error {
	RDB = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	if err := RDB.Ping(context.Background()).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}

	fmt.Println("Redis connected.")
	return nil
}
