package redis

import (
	"context"
	"hahacking/config"
	"log"

	"github.com/go-redis/redis/v8"
)

var Client *redis.Client

func NewClient() *redis.Client {
	redisInfo := config.Get("redis")
	return redis.NewClient(&redis.Options{
		Addr:     redisInfo["host"].(string) + ":" + redisInfo["port"].(string),
		Password: redisInfo["password"].(string),
		DB:       int(redisInfo["db"].(float64)),
	})
}

func Setup() {
	// 初始化Redis客户端
	Client = NewClient()
	pong, err := Client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis - " + pong)
}
