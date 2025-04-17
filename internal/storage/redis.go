package storage

import (
	"context"
	"log"
	"time"

	"ai-language-notes/internal/config"

	"github.com/redis/go-redis/v9"
)

func InitRedis(cfg *config.Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB, // use default DB
		PoolSize: 10,          // Adjust pool size as needed
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Printf("Error connecting to Redis: %v\n", err)
		return nil, err
	}

	log.Println("Successfully connected to Redis.")
	return rdb, nil
}
