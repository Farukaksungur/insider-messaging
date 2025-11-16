package cache

import (
	"context"
	"log"
	"time"

	"insider-messaging/internal/config"

	"github.com/go-redis/redis/v8"
)

// NewRedis Redis bağlantısı oluşturur, bağlantı başarısız olursa nil döner
func NewRedis(cfg *config.Config) *redis.Client {
	if cfg.RedisAddr == "" {
		log.Println("Redis disabled (REDIS_ADDR empty)")
		return nil
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.RedisAddr,
		Password:     cfg.RedisPassword,
		DB:           0,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("Redis connection failed: %v", err)
		return nil
	}

	log.Println("Connected to Redis")
	return rdb
}
