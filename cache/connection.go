package cache

import (
	"context"
	"fmt"
	"jvpayments/config"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func InitRedis() error {
	config := config.LoadConfig()

	host := config.RedisHost
	port := config.RedisPort
	password := config.RedisPassword
	database := config.RedisDb

	var db int
	fmt.Sscanf(database, "%d", &db)

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("error connecting to Redis: %v", err)
	}

	log.Println("Successfully connected to Redis")
	return nil
}

func CloseRedis() {
	if RedisClient != nil {
		RedisClient.Close()
		log.Println("Redis connection closed")
	}
}
