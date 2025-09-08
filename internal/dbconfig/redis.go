package dbconfig

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var Ctx = context.Background()

var RedisClient *redis.Client

func ConnectRedis() *redis.Client {
	addr := getEnv("REDIS_ADDR", "localhost:6379")
	password := getEnv("REDIS_PASSWORD", "")

	db := 0

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// test connection
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		panic(fmt.Sprintf("❌ Failed to connect to Redis: %v", err))
	}

	RedisClient = rdb
	fmt.Println("Connected to redis successfully")
	return rdb
}
